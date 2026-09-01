package declarative_test

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/yueli-official/ytv/api/internal/source"
	"github.com/yueli-official/ytv/api/internal/source/declarative"
	"github.com/yueli-official/ytv/api/internal/source/rulecatalog"
)

func TestLiveKazumiPlaybackContracts(t *testing.T) {
	if os.Getenv("YTV_TEST_LIVE_KAZUMI_PLAYBACK") != "true" {
		t.Skip("set YTV_TEST_LIVE_KAZUMI_PLAYBACK=true for the live playback audit")
	}
	catalog, err := rulecatalog.LoadKazumiSnapshot("5fea5eb84768a2290ddedb156583ef20d9b43561")
	if err != nil {
		t.Fatal(err)
	}
	adapter, err := declarative.New(&http.Client{Timeout: 20 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	for _, original := range catalog.Definitions {
		definition := original
		t.Run(definition.Key, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
			defer cancel()
			page, err := adapter.Search(ctx, definition, source.Query{Text: "海贼王", Page: 1})
			if err != nil {
				t.Fatalf("search: %v", err)
			}
			if len(page.Items) == 0 {
				t.Skip("source returned no result for audit query")
			}
			title, err := adapter.ResolveTitle(ctx, definition, page.Items[0].Locator)
			if err != nil {
				t.Fatalf("title: %v", err)
			}
			if len(title.Episodes) == 0 {
				t.Fatal("title returned no episodes")
			}
			_, episode, err := adapter.ResolveEpisode(ctx, definition, source.EpisodeLocator{Title: title.Locator, EpisodeKey: title.Episodes[0].Key})
			if err != nil {
				t.Fatalf("playback: %v", err)
			}
			if episode.PlaybackURL == "" {
				t.Fatal("playback returned no media URL")
			}
			t.Logf("media=%s", episode.PlaybackURL)
		})
	}
}
