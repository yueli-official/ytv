package analytics

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	ErrInvalid     = errors.New("invalid analytics input")
	ErrConflict    = errors.New("analytics event conflicts")
	ErrUnavailable = errors.New("traffic collector unavailable")
)

const (
	pageView      = "page_view"
	playbackStart = "playback_start"
)

type RecordInput struct {
	EventID    string
	ResourceID string
	VisitorID  string
	OccurredAt time.Time
}

type RecordResult struct {
	Counted bool `json:"counted"`
	Replay  bool `json:"replay"`
}

type Point struct {
	Day            string `json:"day"`
	Views          int64  `json:"views"`
	UniqueVisitors int64  `json:"uniqueVisitors"`
}

type ChannelOverview struct {
	AllTimeViews       int64   `json:"allTimeViews"`
	AllTimeVisitorDays int64   `json:"allTimeVisitorDays"`
	Views              int64   `json:"views"`
	VisitorDays        int64   `json:"visitorDays"`
	Series             []Point `json:"series"`
}

type Overview struct {
	Days     int              `json:"days"`
	Pages    ChannelOverview  `json:"pages"`
	Playback *ChannelOverview `json:"playback"`
}

type Traffic struct {
	base  *url.URL
	token string
	http  *http.Client
	clock func() time.Time
}

func NewTraffic(baseURL, token string, clock func() time.Time) (*Traffic, error) {
	base, err := url.Parse(baseURL)
	if err != nil || base.Hostname() == "" || base.User != nil || base.ForceQuery || base.RawQuery != "" || base.Fragment != "" ||
		(base.Scheme != "https" && !(base.Scheme == "http" && (isLoopback(base.Hostname()) || base.Host == "platform-traffic:8096"))) {
		return nil, errors.New("ytv/analytics: collector URL must use HTTPS, loopback HTTP or the private platform collector, without credentials, query or fragment")
	}
	if len(token) < 32 || len(token) > 128 || strings.ContainsFunc(token, func(c rune) bool {
		return !(c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-' || c == '_')
	}) {
		return nil, errors.New("ytv/analytics: collector credential must be a 32–128 character base64url string")
	}
	if clock == nil {
		clock = time.Now
	}
	client := &http.Client{
		Timeout: 1500 * time.Millisecond,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	return &Traffic{base: base, token: token, http: client, clock: clock}, nil
}

func isLoopback(host string) bool {
	return strings.EqualFold(host, "localhost") || net.ParseIP(host).IsLoopback()
}

func (service *Traffic) RecordPage(ctx context.Context, input RecordInput) (RecordResult, error) {
	path := strings.TrimSpace(input.ResourceID)
	if !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") || strings.ContainsAny(path, "\\?#") || len(path) > 200 {
		return RecordResult{}, fmt.Errorf("%w: page path is invalid", ErrInvalid)
	}
	if strings.HasPrefix(path, "/manage") || strings.HasPrefix(path, "/auth") || strings.HasPrefix(path, "/api/") {
		return RecordResult{}, fmt.Errorf("%w: page path is not public", ErrInvalid)
	}
	input.ResourceID = path
	return service.record(ctx, pageView, input)
}

func (service *Traffic) RecordPlayback(ctx context.Context, input RecordInput) (RecordResult, error) {
	resourceID := strings.TrimSpace(input.ResourceID)
	if resourceID == "" || len(resourceID) > 200 || strings.Count(resourceID, ":") < 2 {
		return RecordResult{}, fmt.Errorf("%w: playback resource is invalid", ErrInvalid)
	}
	input.ResourceID = resourceID
	return service.record(ctx, playbackStart, input)
}

func (service *Traffic) record(ctx context.Context, metric string, input RecordInput) (RecordResult, error) {
	eventID := strings.TrimSpace(input.EventID)
	if eventID == "" || len(eventID) > 200 || strings.ContainsRune(eventID, '\x00') {
		return RecordResult{}, fmt.Errorf("%w: event ID is invalid", ErrInvalid)
	}
	occurredAt := input.OccurredAt.UTC()
	if occurredAt.IsZero() {
		occurredAt = service.clock().UTC()
	}
	visitorKey := ""
	if visitorID := strings.TrimSpace(input.VisitorID); visitorID != "" {
		if length := utf8.RuneCountInString(visitorID); length < 16 || length > 128 {
			return RecordResult{}, fmt.Errorf("%w: visitor ID must contain 16 to 128 characters", ErrInvalid)
		}
		visitorKey = opaqueKey(visitorID)
	}
	payload, err := json.Marshal(struct {
		ID         string    `json:"id"`
		Metric     string    `json:"metric"`
		Resource   string    `json:"resource"`
		OccurredAt time.Time `json:"occurredAt"`
		VisitorKey string    `json:"visitorKey,omitempty"`
	}{eventID, metric, opaqueKey(input.ResourceID), occurredAt, visitorKey})
	if err != nil {
		return RecordResult{}, fmt.Errorf("%w: invalid event timestamp", ErrInvalid)
	}
	var result RecordResult
	target := service.base.JoinPath("v1", "events")
	if err := service.request(ctx, http.MethodPost, target, payload, &result); err != nil {
		return RecordResult{}, err
	}
	return result, nil
}

func (service *Traffic) Overview(ctx context.Context, days int) (Overview, error) {
	switch days {
	case 0:
		days = 14
	case 7, 14, 30:
	default:
		return Overview{}, fmt.Errorf("%w: days must be 7, 14, or 30", ErrInvalid)
	}
	target := service.base.JoinPath("v1", "site-report")
	target.RawQuery = url.Values{"days": {strconv.Itoa(days)}}.Encode()
	var report Overview
	if err := service.request(ctx, http.MethodGet, target, nil, &report); err != nil {
		return Overview{}, err
	}
	if report.Days != days || report.Playback == nil {
		return Overview{}, fmt.Errorf("%w: TV site report is incomplete", ErrUnavailable)
	}
	return report, nil
}

// The collector receives stable opaque keys instead of product paths or raw browser identifiers.
func opaqueKey(value string) string {
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:])
}

func (service *Traffic) request(ctx context.Context, method string, target *url.URL, payload []byte, result any) error {
	retry, err := service.send(ctx, method, target, payload, result)
	if !retry || ctx.Err() != nil {
		return err
	}
	// POST retries resend the exact same event ID, occurrence time and payload.
	timer := time.NewTimer(100 * time.Millisecond)
	defer timer.Stop()
	select {
	case <-timer.C:
		_, err = service.send(ctx, method, target, payload, result)
		return err
	case <-ctx.Done():
		return fmt.Errorf("%w: %v", ErrUnavailable, ctx.Err())
	}
}

func (service *Traffic) send(ctx context.Context, method string, target *url.URL, payload []byte, result any) (bool, error) {
	request, err := http.NewRequestWithContext(ctx, method, target.String(), bytes.NewReader(payload))
	if err != nil {
		return false, ErrUnavailable
	}
	request.Header.Set("Authorization", "Bearer "+service.token)
	if method == http.MethodPost {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := service.http.Do(request)
	if err != nil {
		return true, fmt.Errorf("%w: request failed", ErrUnavailable)
	}
	defer response.Body.Close()
	switch response.StatusCode {
	case http.StatusOK:
		if err := json.NewDecoder(io.LimitReader(response.Body, 64<<10)).Decode(result); err != nil {
			return false, fmt.Errorf("%w: invalid response", ErrUnavailable)
		}
		return false, nil
	case http.StatusBadRequest:
		return false, ErrInvalid
	case http.StatusConflict:
		return false, ErrConflict
	case http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true, ErrUnavailable
	default:
		return false, fmt.Errorf("%w: status %d", ErrUnavailable, response.StatusCode)
	}
}
