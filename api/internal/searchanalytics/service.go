package searchanalytics

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

var ErrInvalid = errors.New("invalid searchanalytics input")

type Event struct {
	ID                    string
	Query                 string
	ResultCount           int
	SuccessfulSourceCount int
	FailedSourceCount     int
	Duration              time.Duration
	OccurredAt            time.Time
}

type RecordResult struct {
	EventID string
	Replay  bool
}

type Summary struct {
	Searches           int64
	ZeroResultSearches int64
	IncompleteSearches int64
	AverageDurationMS  int64
}

type QueryStat struct {
	Query              string
	Searches           int64
	ZeroResultSearches int64
	AverageResultCount float64
	AverageDurationMS  int64
	LastSearchedAt     time.Time
}

type DayStat struct {
	Day                string
	Searches           int64
	ZeroResultSearches int64
}

type Overview struct {
	Days       int
	Summary    Summary
	TopQueries []QueryStat
	Series     []DayStat
}

type Store interface {
	Record(context.Context, Event) (bool, error)
	Overview(context.Context, time.Time, int) (Overview, error)
}

type Service struct {
	store Store
	clock func() time.Time
}

func New(store Store, clock func() time.Time) (*Service, error) {
	if store == nil {
		return nil, errors.New("ytv/searchanalytics: store is required")
	}
	if clock == nil {
		clock = time.Now
	}
	return &Service{store: store, clock: clock}, nil
}

func (service *Service) Record(ctx context.Context, event Event) (RecordResult, error) {
	event.Query = normalizeQuery(event.Query)
	if event.Query == "" || utf8.RuneCountInString(event.Query) > 100 {
		return RecordResult{}, fmt.Errorf("%w: search query must contain 1 to 100 characters", ErrInvalid)
	}
	if event.ResultCount < 0 || event.SuccessfulSourceCount < 0 || event.FailedSourceCount < 0 || event.Duration < 0 {
		return RecordResult{}, fmt.Errorf("%w: search outcome counts and duration must not be negative", ErrInvalid)
	}
	event.ID = strings.TrimSpace(event.ID)
	if event.ID == "" {
		var err error
		event.ID, err = newEventID()
		if err != nil {
			return RecordResult{}, err
		}
	}
	if len(event.ID) < 16 || len(event.ID) > 200 {
		return RecordResult{}, fmt.Errorf("%w: search event ID must contain 16 to 200 bytes", ErrInvalid)
	}
	if event.OccurredAt.IsZero() {
		event.OccurredAt = service.clock()
	}
	event.OccurredAt = event.OccurredAt.UTC()
	if event.OccurredAt.After(service.clock().Add(5 * time.Minute)) {
		return RecordResult{}, fmt.Errorf("%w: search event cannot be in the future", ErrInvalid)
	}
	replay, err := service.store.Record(ctx, event)
	if err != nil {
		return RecordResult{}, err
	}
	return RecordResult{EventID: event.ID, Replay: replay}, nil
}

func (service *Service) Overview(ctx context.Context, days, limit int) (Overview, error) {
	if days == 0 {
		days = 14
	}
	if days != 7 && days != 14 && days != 30 {
		return Overview{}, fmt.Errorf("%w: days must be 7, 14, or 30", ErrInvalid)
	}
	if limit == 0 {
		limit = 10
	}
	if limit < 1 || limit > 50 {
		return Overview{}, fmt.Errorf("%w: limit must be between 1 and 50", ErrInvalid)
	}
	now := service.clock()
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return Overview{}, err
	}
	localNow := now.In(location)
	start := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, location).AddDate(0, 0, -(days - 1))
	overview, err := service.store.Overview(ctx, start.UTC(), limit)
	if err != nil {
		return Overview{}, err
	}
	overview.Days = days
	return overview, nil
}

func normalizeQuery(value string) string {
	return strings.ToLower(strings.Join(strings.Fields(value), " "))
}

func newEventID() (string, error) {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		return "", fmt.Errorf("create search event ID: %w", err)
	}
	return "search_" + hex.EncodeToString(buffer), nil
}
