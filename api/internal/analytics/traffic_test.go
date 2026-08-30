package analytics

import (
	"context"
	"testing"
	"time"

	"github.com/yueli-official/foundation/go/traffic"
)

func TestPageAndPlaybackCountersRemainSeparate(t *testing.T) {
	now := time.Date(2026, 8, 30, 12, 0, 0, 0, time.FixedZone("CST", 8*60*60))
	pageCatalog := traffic.MustCompile(Definition(PageResourceKind))
	playbackCatalog := traffic.MustCompile(Definition(PlaybackResourceKind))
	pageModule, err := traffic.NewMemory(pageCatalog, traffic.MemoryOptions{Secret: []byte("page-secret-32-bytes-for-testing-1"), Clock: func() time.Time { return now }})
	if err != nil {
		t.Fatal(err)
	}
	playbackModule, err := traffic.NewMemory(playbackCatalog, traffic.MemoryOptions{Secret: []byte("playback-secret-32-bytes-testing"), Clock: func() time.Time { return now }})
	if err != nil {
		t.Fatal(err)
	}
	service, err := NewTraffic(pageModule, pageCatalog, playbackModule, playbackCatalog, func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	visitor := "visitor-0000000000000001"
	if _, err := service.RecordPage(context.Background(), RecordInput{EventID: "event-page-0000001", ResourceID: "/", VisitorID: visitor, OccurredAt: now}); err != nil {
		t.Fatal(err)
	}
	if _, err := service.RecordPlayback(context.Background(), RecordInput{EventID: "event-play-0000001", ResourceID: "fixture:ring-sea:0", VisitorID: visitor, OccurredAt: now}); err != nil {
		t.Fatal(err)
	}
	overview, err := service.Overview(context.Background(), 7)
	if err != nil {
		t.Fatal(err)
	}
	if overview.Pages.AllTime.Views != 1 || overview.Playback.AllTime.Views != 1 {
		t.Fatalf("overview = %#v", overview)
	}
}
