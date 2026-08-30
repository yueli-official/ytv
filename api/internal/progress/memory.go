package progress

import (
	"context"
	"sort"
	"sync"
)

type MemoryStore struct {
	mu      sync.Mutex
	records map[string]Record
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{records: map[string]Record{}}
}

func recordKey(record Record) string {
	return record.UserKey + "\x00" + record.SourceKey + "\x00" + record.ExternalID + "\x00" + record.EpisodeKey
}

func (store *MemoryStore) Save(_ context.Context, record Record) (SaveResult, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	key := recordKey(record)
	current, exists := store.records[key]
	if exists && !record.UpdatedAt.After(current.UpdatedAt) {
		return SaveResult{Record: current, Applied: false}, nil
	}
	record.Revision = current.Revision + 1
	store.records[key] = record
	return SaveResult{Record: record, Applied: true}, nil
}

func (store *MemoryStore) List(_ context.Context, userKey string, limit int) ([]Record, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	result := make([]Record, 0)
	for _, record := range store.records {
		if record.UserKey == userKey {
			result = append(result, record)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].UpdatedAt.After(result[j].UpdatedAt) })
	if len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}
