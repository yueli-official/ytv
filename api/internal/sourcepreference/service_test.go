package sourcepreference

import (
	"context"
	"testing"
	"time"
)

type memoryStore struct{ value Preference }

func (store *memoryStore) Get(_ context.Context, userKey string) (Preference, error) {
	if !store.value.Exists {
		return Preference{UserKey: userKey, DisabledSourceKeys: []string{}}, nil
	}
	return store.value, nil
}
func (store *memoryStore) Save(_ context.Context, value Preference, expected int64) (Preference, error) {
	if store.value.Exists && store.value.Revision != expected {
		return Preference{}, ErrRevisionConflict
	}
	if !store.value.Exists && expected != 0 {
		return Preference{}, ErrRevisionConflict
	}
	value.Exists = true
	value.Revision = expected + 1
	store.value = value
	return value, nil
}

func TestServiceNormalizesDisabledKeysAndUsesRevision(t *testing.T) {
	store := &memoryStore{}
	service, _ := New(store, func() time.Time { return time.Unix(10, 0) })
	saved, err := service.Save(context.Background(), " user ", []string{"b", "a", "a", "bad key"}, 0)
	if err != nil || saved.Revision != 1 || len(saved.DisabledSourceKeys) != 2 || saved.DisabledSourceKeys[0] != "a" {
		t.Fatalf("saved/error = %#v / %v", saved, err)
	}
	if _, err := service.Save(context.Background(), "user", []string{"a"}, 0); err != ErrRevisionConflict {
		t.Fatalf("conflict = %v", err)
	}
}
