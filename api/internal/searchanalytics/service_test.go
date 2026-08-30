package searchanalytics

import (
	"context"
	"testing"
	"time"
)

func TestRecordIsIdempotentAndOverviewNormalizesQueries(t *testing.T) {
	now := time.Date(2026, 8, 30, 12, 0, 0, 0, time.FixedZone("CST", 8*60*60))
	service, err := New(NewMemoryStore(), func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	event := Event{
		ID: "search-event-0000000001", Query: "  霜线   计划 ", ResultCount: 2,
		SuccessfulSourceCount: 1, FailedSourceCount: 1, Duration: 120 * time.Millisecond,
	}
	first, err := service.Record(context.Background(), event)
	if err != nil || first.Replay {
		t.Fatalf("first record = %#v, %v", first, err)
	}
	second, err := service.Record(context.Background(), event)
	if err != nil || !second.Replay {
		t.Fatalf("second record = %#v, %v", second, err)
	}
	overview, err := service.Overview(context.Background(), 14, 10)
	if err != nil {
		t.Fatal(err)
	}
	if overview.Summary.Searches != 1 || overview.Summary.IncompleteSearches != 1 || len(overview.TopQueries) != 1 || overview.TopQueries[0].Query != "霜线 计划" {
		t.Fatalf("overview = %#v", overview)
	}
}
