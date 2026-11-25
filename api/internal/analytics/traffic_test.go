package analytics

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestPageAndPlaybackCountersRemainSeparate(t *testing.T) {
	now := time.Date(2026, 8, 30, 4, 0, 0, 0, time.UTC)
	token := strings.Repeat("T", 32)
	type observation struct {
		ID         string    `json:"id"`
		Metric     string    `json:"metric"`
		Resource   string    `json:"resource"`
		OccurredAt time.Time `json:"occurredAt"`
		VisitorKey string    `json:"visitorKey"`
	}
	requests := make(chan []byte, 3)
	collector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+token {
			t.Error("collector did not receive the site credential")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		switch r.Method + " " + r.URL.Path {
		case "POST /v1/events":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Errorf("read event: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			requests <- body
			if len(requests) == 1 {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			_, _ = io.WriteString(w, `{"counted":true,"replay":false}`)
		case "GET /v1/site-report":
			if r.URL.Query().Get("days") != "7" {
				t.Error("collector did not receive the report period")
			}
			_, _ = io.WriteString(w, `{"days":7,"pages":{"allTimeViews":1},"playback":{"allTimeViews":1}}`)
		default:
			t.Errorf("unexpected collector request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer collector.Close()
	service, err := NewTraffic(collector.URL, token, func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	visitor := "visitor-0000000000000001"
	page, err := service.RecordPage(context.Background(), RecordInput{EventID: "event-page-0000001", ResourceID: "/", VisitorID: visitor})
	if err != nil || !page.Counted {
		t.Fatalf("page result = %#v, %v", page, err)
	}
	playback, err := service.RecordPlayback(context.Background(), RecordInput{EventID: "event-play-0000001", ResourceID: "fixture:ring-sea:0", VisitorID: visitor, OccurredAt: now})
	if err != nil || !playback.Counted {
		t.Fatalf("playback result = %#v, %v", playback, err)
	}
	if len(requests) != 3 {
		t.Fatalf("collector received %d events, expected two attempts and one playback", len(requests))
	}
	first, retry, third := <-requests, <-requests, <-requests
	if !bytes.Equal(first, retry) {
		t.Fatal("retry changed the event payload")
	}
	var pageEvent, playbackEvent observation
	if err := json.Unmarshal(first, &pageEvent); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(third, &playbackEvent); err != nil {
		t.Fatal(err)
	}
	if pageEvent.Metric != "page_view" || playbackEvent.Metric != "playback_start" || pageEvent.ID == playbackEvent.ID ||
		pageEvent.Resource == "/" || playbackEvent.Resource == "fixture:ring-sea:0" || pageEvent.Resource == playbackEvent.Resource ||
		pageEvent.VisitorKey == "" || pageEvent.VisitorKey == visitor || pageEvent.VisitorKey != playbackEvent.VisitorKey ||
		!pageEvent.OccurredAt.Equal(now) || !playbackEvent.OccurredAt.Equal(now) {
		t.Fatal("collector event fields did not preserve traffic semantics")
	}
	overview, err := service.Overview(context.Background(), 7)
	if err != nil || overview.Pages.AllTimeViews != 1 || overview.Playback == nil || overview.Playback.AllTimeViews != 1 {
		t.Fatalf("overview = %#v, %v", overview, err)
	}
}
