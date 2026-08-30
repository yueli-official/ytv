package progress

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

type Record struct {
	UserKey         string
	SourceKey       string
	ExternalID      string
	EpisodeKey      string
	TitleName       string
	EpisodeTitle    string
	PositionSeconds int
	DurationSeconds int
	Completed       bool
	Revision        int64
	UpdatedAt       time.Time
}

type SaveResult struct {
	Record  Record
	Applied bool
}

type Store interface {
	Save(context.Context, Record) (SaveResult, error)
	List(context.Context, string, int) ([]Record, error)
}

type Service struct {
	store Store
	clock func() time.Time
}

func New(store Store, clock func() time.Time) (*Service, error) {
	if store == nil {
		return nil, errors.New("ytv/progress: store is required")
	}
	if clock == nil {
		clock = time.Now
	}
	return &Service{store: store, clock: clock}, nil
}

func (service *Service) Save(ctx context.Context, record Record) (SaveResult, error) {
	record.UserKey = strings.TrimSpace(record.UserKey)
	record.SourceKey = strings.TrimSpace(record.SourceKey)
	record.ExternalID = strings.TrimSpace(record.ExternalID)
	record.EpisodeKey = strings.TrimSpace(record.EpisodeKey)
	record.TitleName = strings.TrimSpace(record.TitleName)
	record.EpisodeTitle = strings.TrimSpace(record.EpisodeTitle)
	if record.UserKey == "" || record.SourceKey == "" || record.ExternalID == "" || record.EpisodeKey == "" {
		return SaveResult{}, fmt.Errorf("progress identity and locator are required")
	}
	if utf8.RuneCountInString(record.TitleName) > 200 || utf8.RuneCountInString(record.EpisodeTitle) > 200 {
		return SaveResult{}, fmt.Errorf("progress display metadata is too long")
	}
	if record.PositionSeconds < 0 || record.DurationSeconds <= 0 || record.PositionSeconds > record.DurationSeconds+30 {
		return SaveResult{}, fmt.Errorf("progress position or duration is invalid")
	}
	now := service.clock()
	if record.UpdatedAt.IsZero() {
		record.UpdatedAt = now
	}
	if record.UpdatedAt.Before(now.Add(-30*24*time.Hour)) || record.UpdatedAt.After(now.Add(5*time.Minute)) {
		return SaveResult{}, fmt.Errorf("progress timestamp is outside the accepted window")
	}
	record.UpdatedAt = record.UpdatedAt.UTC()
	record.Completed = record.Completed || record.PositionSeconds >= record.DurationSeconds-10
	return service.store.Save(ctx, record)
}

func (service *Service) List(ctx context.Context, userKey string, limit int) ([]Record, error) {
	userKey = strings.TrimSpace(userKey)
	if userKey == "" {
		return nil, fmt.Errorf("progress user key is required")
	}
	if limit == 0 {
		limit = 100
	}
	if limit < 1 || limit > 100 {
		return nil, fmt.Errorf("progress limit must be between 1 and 100")
	}
	return service.store.List(ctx, userKey, limit)
}
