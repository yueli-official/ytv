package siteprofile

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

var ErrInvalid = errors.New("invalid siteprofile input")

var ErrRevisionConflict = errors.New("site profile revision conflict")

type Profile struct {
	SiteKey     string
	Name        string
	Description string
	Revision    int64
	UpdatedAt   time.Time
}

type Store interface {
	Get(context.Context, string) (Profile, bool, error)
	Save(context.Context, Profile, int64) (Profile, error)
}

type Service struct {
	siteKey string
	store   Store
	clock   func() time.Time
}

func New(siteKey string, store Store, clock func() time.Time) (*Service, error) {
	siteKey = strings.TrimSpace(siteKey)
	if siteKey == "" || store == nil {
		return nil, errors.New("ytv/siteprofile: site key and store are required")
	}
	if clock == nil {
		clock = time.Now
	}
	return &Service{siteKey: siteKey, store: store, clock: clock}, nil
}

func (service *Service) Get(ctx context.Context) (Profile, error) {
	profile, exists, err := service.store.Get(ctx, service.siteKey)
	if err != nil {
		return Profile{}, err
	}
	if !exists {
		return Profile{SiteKey: service.siteKey, Name: "YTV", Description: "从多个获准来源搜索并继续观看。"}, nil
	}
	return profile, nil
}

func (service *Service) Update(ctx context.Context, name, description string, expectedRevision int64) (Profile, error) {
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)
	if name == "" || utf8.RuneCountInString(name) > 80 || utf8.RuneCountInString(description) > 300 || expectedRevision < 0 {
		return Profile{}, fmt.Errorf("%w: site profile input is invalid", ErrInvalid)
	}
	current, err := service.Get(ctx)
	if err != nil {
		return Profile{}, err
	}
	if current.Revision != expectedRevision {
		return Profile{}, ErrRevisionConflict
	}
	return service.store.Save(ctx, Profile{
		SiteKey: service.siteKey, Name: name, Description: description, UpdatedAt: service.clock().UTC(),
	}, expectedRevision)
}
