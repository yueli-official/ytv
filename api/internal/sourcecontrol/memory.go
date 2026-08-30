package sourcecontrol

import (
	"context"
	"sync"
)

type MemoryStore struct {
	mu       sync.Mutex
	settings map[string]Setting
}

func NewMemoryStore(values ...Setting) *MemoryStore {
	settings := make(map[string]Setting, len(values))
	for _, value := range values {
		settings[value.SourceKey] = value
	}
	return &MemoryStore{settings: settings}
}

func (store *MemoryStore) LoadAll(context.Context) ([]Setting, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	result := make([]Setting, 0, len(store.settings))
	for _, setting := range store.settings {
		result = append(result, setting)
	}
	return result, nil
}

func (store *MemoryStore) Save(_ context.Context, setting Setting, expectedRevision int64) (Setting, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	current, exists := store.settings[setting.SourceKey]
	if (exists && current.Revision != expectedRevision) || (!exists && expectedRevision != 0) {
		return Setting{}, ErrRevisionConflict
	}
	setting.Revision = expectedRevision + 1
	store.settings[setting.SourceKey] = setting
	return setting, nil
}
