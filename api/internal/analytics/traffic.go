package analytics

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/yueli-official/foundation/go/traffic"
)

var ErrInvalid = errors.New("invalid analytics input")

const (
	PageResourceKind     traffic.ResourceKind = "page"
	PlaybackResourceKind traffic.ResourceKind = "episode"
)

type Recorder interface {
	traffic.Recorder
	traffic.VisitorTokenizer
	traffic.Reader
}

type RecordInput struct {
	EventID    string
	ResourceID string
	VisitorID  string
	OccurredAt time.Time
}

type RecordResult struct {
	Counted bool
	Replay  bool
}

type ChannelOverview struct {
	AllTime traffic.Totals
	Recent  traffic.Totals
	Series  []traffic.SeriesPoint
}

type Overview struct {
	Days     int
	Pages    ChannelOverview
	Playback ChannelOverview
}

type Traffic struct {
	pages           Recorder
	playback        Recorder
	pageCatalog     *traffic.Catalog
	playbackCatalog *traffic.Catalog
	clock           func() time.Time
}

func Definition(kind traffic.ResourceKind) traffic.Definition {
	return traffic.Definition{
		Version:       traffic.DefinitionVersion,
		TimeZone:      "Asia/Shanghai",
		ResourceKinds: []traffic.ResourceKindDefinition{{Key: kind}},
		Limits: traffic.Limits{
			MaxPastAge:             30 * 24 * time.Hour,
			ReceiptRetention:       30 * 24 * time.Hour,
			VisitorMarkerRetention: 31 * 24 * time.Hour,
		},
	}
}

func NewTraffic(pages Recorder, pageCatalog *traffic.Catalog, playback Recorder, playbackCatalog *traffic.Catalog, clock func() time.Time) (*Traffic, error) {
	if pages == nil || pageCatalog == nil || playback == nil || playbackCatalog == nil {
		return nil, errors.New("ytv/analytics: page and playback traffic modules are required")
	}
	if clock == nil {
		clock = time.Now
	}
	return &Traffic{pages: pages, pageCatalog: pageCatalog, playback: playback, playbackCatalog: playbackCatalog, clock: clock}, nil
}

func (service *Traffic) RecordPage(ctx context.Context, input RecordInput) (RecordResult, error) {
	path := strings.TrimSpace(input.ResourceID)
	if path == "" || !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") || strings.Contains(path, "\\") || len(path) > 200 {
		return RecordResult{}, fmt.Errorf("%w: page path is invalid", ErrInvalid)
	}
	if strings.HasPrefix(path, "/manage") || strings.HasPrefix(path, "/auth") || strings.HasPrefix(path, "/api/") {
		return RecordResult{}, fmt.Errorf("%w: page path is not public", ErrInvalid)
	}
	return service.record(ctx, service.pages, PageResourceKind, input)
}

func (service *Traffic) RecordPlayback(ctx context.Context, input RecordInput) (RecordResult, error) {
	resourceID := strings.TrimSpace(input.ResourceID)
	if resourceID == "" || len(resourceID) > 200 || strings.Count(resourceID, ":") < 2 {
		return RecordResult{}, fmt.Errorf("%w: playback resource is invalid", ErrInvalid)
	}
	input.ResourceID = resourceID
	return service.record(ctx, service.playback, PlaybackResourceKind, input)
}

func (service *Traffic) record(ctx context.Context, module Recorder, kind traffic.ResourceKind, input RecordInput) (RecordResult, error) {
	occurredAt := input.OccurredAt.UTC()
	if occurredAt.IsZero() {
		occurredAt = service.clock().UTC()
	}
	observation := traffic.Observation{
		EventID:    traffic.EventID(strings.TrimSpace(input.EventID)),
		Resource:   traffic.Resource{Kind: kind, ID: strings.TrimSpace(input.ResourceID)},
		OccurredAt: occurredAt,
		Class:      traffic.VisitHuman,
	}
	if visitorID := strings.TrimSpace(input.VisitorID); visitorID != "" {
		if utf8.RuneCountInString(visitorID) < 16 || utf8.RuneCountInString(visitorID) > 128 {
			return RecordResult{}, fmt.Errorf("%w: visitor ID must contain 16 to 128 characters", ErrInvalid)
		}
		token, err := module.TokenizeVisitor(ctx, occurredAt, []byte("browser\x00"+visitorID))
		if err != nil {
			return RecordResult{}, err
		}
		observation.HasVisitor = true
		observation.VisitorToken = token
	}
	result, err := module.Record(ctx, observation)
	if err != nil {
		return RecordResult{}, err
	}
	return RecordResult{Counted: result.Counted, Replay: result.Replay}, nil
}

func (service *Traffic) Overview(ctx context.Context, days int) (Overview, error) {
	if days == 0 {
		days = 14
	}
	if days != 7 && days != 14 && days != 30 {
		return Overview{}, fmt.Errorf("%w: days must be 7, 14, or 30", ErrInvalid)
	}
	pages, err := service.channelOverview(ctx, service.pages, service.pageCatalog, days)
	if err != nil {
		return Overview{}, err
	}
	playback, err := service.channelOverview(ctx, service.playback, service.playbackCatalog, days)
	if err != nil {
		return Overview{}, err
	}
	return Overview{Days: days, Pages: pages, Playback: playback}, nil
}

func (service *Traffic) channelOverview(ctx context.Context, module Recorder, catalog *traffic.Catalog, days int) (ChannelOverview, error) {
	now := service.clock()
	to := catalog.DayAt(now.AddDate(0, 0, 1))
	from := catalog.DayAt(now.AddDate(0, 0, -(days - 1)))
	rangeValue := traffic.DateRange{From: from, To: to}
	allTime, err := module.Summary(ctx, traffic.SummaryQuery{Scope: traffic.InstanceScope()})
	if err != nil {
		return ChannelOverview{}, err
	}
	recent, err := module.Summary(ctx, traffic.SummaryQuery{Scope: traffic.InstanceScope(), Range: &rangeValue})
	if err != nil {
		return ChannelOverview{}, err
	}
	series, err := module.Series(ctx, traffic.SeriesQuery{Scope: traffic.InstanceScope(), Range: rangeValue})
	if err != nil {
		return ChannelOverview{}, err
	}
	return ChannelOverview{AllTime: allTime.Totals, Recent: recent.Totals, Series: series}, nil
}
