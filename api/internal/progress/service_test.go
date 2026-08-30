package progress

import (
	"context"
	"testing"
	"time"
)

func TestNewestProgressWinsAndCompletionIsDerived(t *testing.T) {
	now := time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC)
	service, err := New(NewMemoryStore(), func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	input := Record{
		UserKey: "user", SourceKey: "fixture", ExternalID: "ring-sea", EpisodeKey: "0",
		PositionSeconds: 96, DurationSeconds: 100, UpdatedAt: now,
	}
	first, err := service.Save(context.Background(), input)
	if err != nil || !first.Applied || !first.Record.Completed || first.Record.Revision != 1 {
		t.Fatalf("first save = %#v, %v", first, err)
	}
	input.PositionSeconds = 20
	input.UpdatedAt = now.Add(-time.Minute)
	stale, err := service.Save(context.Background(), input)
	if err != nil || stale.Applied || stale.Record.PositionSeconds != 96 {
		t.Fatalf("stale save = %#v, %v", stale, err)
	}
	records, err := service.List(context.Background(), "user", 10)
	if err != nil || len(records) != 1 {
		t.Fatalf("records = %#v, %v", records, err)
	}
}
