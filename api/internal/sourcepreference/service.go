package sourcepreference

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
)

var ErrInvalid = errors.New("invalid sourcepreference input")

var ErrRevisionConflict = errors.New("source preference revision conflict")
var sourceKeyPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,100}$`)

type Preference struct {
	UserKey            string
	DisabledSourceKeys []string
	Revision           int64
	UpdatedAt          time.Time
	Exists             bool
}

type Store interface {
	Get(context.Context, string) (Preference, error)
	Save(context.Context, Preference, int64) (Preference, error)
}

type Service struct {
	store Store
	clock func() time.Time
}

func New(store Store, clock func() time.Time) (*Service, error) {
	if store == nil {
		return nil, errors.New("ytv/sourcepreference: store is required")
	}
	if clock == nil {
		clock = time.Now
	}
	return &Service{store: store, clock: clock}, nil
}

func (service *Service) Get(ctx context.Context, userKey string) (Preference, error) {
	userKey = strings.TrimSpace(userKey)
	if userKey == "" {
		return Preference{}, fmt.Errorf("%w: source preference user is required", ErrInvalid)
	}
	return service.store.Get(ctx, userKey)
}

func (service *Service) Save(ctx context.Context, userKey string, disabled []string, expectedRevision int64) (Preference, error) {
	userKey = strings.TrimSpace(userKey)
	if userKey == "" || expectedRevision < 0 {
		return Preference{}, fmt.Errorf("%w: source preference input is invalid", ErrInvalid)
	}
	seen := map[string]bool{}
	keys := make([]string, 0, len(disabled))
	for _, key := range disabled {
		key = strings.TrimSpace(key)
		if !sourceKeyPattern.MatchString(key) || seen[key] {
			continue
		}
		seen[key] = true
		keys = append(keys, key)
		if len(keys) > 64 {
			return Preference{}, fmt.Errorf("%w: source preference has too many disabled sources", ErrInvalid)
		}
	}
	sort.Strings(keys)
	return service.store.Save(ctx, Preference{UserKey: userKey, DisabledSourceKeys: keys, UpdatedAt: service.clock().UTC()}, expectedRevision)
}
