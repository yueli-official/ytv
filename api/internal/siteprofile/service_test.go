package siteprofile

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDefaultProfilePersistsWithOptimisticRevision(t *testing.T) {
	now := time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC)
	service, err := New("ytv-main", NewMemoryStore(), func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	initial, err := service.Get(context.Background())
	if err != nil || initial.Name != "YTV" || initial.Revision != 0 {
		t.Fatalf("initial = %#v, %v", initial, err)
	}
	saved, err := service.Update(context.Background(), "YTV 本地", "说明", 0)
	if err != nil || saved.Revision != 1 {
		t.Fatalf("saved = %#v, %v", saved, err)
	}
	if _, err := service.Update(context.Background(), "过期写入", "", 0); !errors.Is(err, ErrRevisionConflict) {
		t.Fatalf("stale error = %v", err)
	}
}
