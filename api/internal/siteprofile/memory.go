package siteprofile

import (
	"context"
	"sync"
)

type MemoryStore struct {
	mu      sync.Mutex
	profile *Profile
}

func NewMemoryStore() *MemoryStore { return &MemoryStore{} }

func (store *MemoryStore) Get(_ context.Context, siteKey string) (Profile, bool, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	if store.profile == nil || store.profile.SiteKey != siteKey {
		return Profile{}, false, nil
	}
	return *store.profile, true, nil
}

func (store *MemoryStore) Save(_ context.Context, profile Profile, expectedRevision int64) (Profile, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	if (store.profile == nil && expectedRevision != 0) || (store.profile != nil && store.profile.Revision != expectedRevision) {
		return Profile{}, ErrRevisionConflict
	}
	profile.Revision = expectedRevision + 1
	store.profile = &profile
	return profile, nil
}
