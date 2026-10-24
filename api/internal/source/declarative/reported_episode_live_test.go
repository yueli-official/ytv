package declarative_test

import (
	"context"
	"encoding/base64"
	"github.com/yueli-official/ytv/api/internal/playback"
	"github.com/yueli-official/ytv/api/internal/source"
	"github.com/yueli-official/ytv/api/internal/source/declarative"
	"github.com/yueli-official/ytv/api/internal/source/rulecatalog"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"
)

func TestLive7sefunReportedEpisode(t *testing.T) {
	if os.Getenv("YTV_TEST_LIVE_7SEFUN") != "true" {
		t.Skip("explicit live metadata resolution only")
	}
	catalog, err := rulecatalog.LoadKazumiSnapshot("0972545266e7ac7cc06e7e522892239824f7d713")
	if err != nil {
		t.Fatal(err)
	}
	for _, definition := range catalog.Definitions {
		if definition.Key != "kazumi-7sefun" {
			continue
		}
		definition.Enabled = true
		adapter, _ := declarative.New(&http.Client{Timeout: 20 * time.Second})
		locator := source.EpisodeLocator{Title: source.TitleLocator{SourceKey: definition.Key, ExternalID: base64.RawURLEncoding.EncodeToString([]byte("https://www.7sefun.top/voddetail/31667.html"))}, EpisodeKey: base64.RawURLEncoding.EncodeToString([]byte("https://www.7sefun.top/vodplay/31667-1-1.html"))}
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()
		_, episode, err := adapter.ResolveEpisode(ctx, definition, locator)
		if err != nil {
			t.Fatal(err)
		}
		target, err := url.Parse(episode.PlaybackURL)
		if err != nil {
			t.Fatal("invalid media URL")
		}
		t.Logf("resolved scheme=%s host=%s path=%s approved=%v", target.Scheme, target.Hostname(), target.EscapedPath(), definition.PlaybackHosts)
		registry, err := source.NewRegistry([]source.Definition{definition}, map[source.Kind]source.Adapter{source.KindDeclarative: adapter})
		if err != nil {
			t.Fatal(err)
		}
		resolver, _ := playback.New(registry)
		if _, err := resolver.Resolve(ctx, locator); err != nil {
			t.Fatalf("reported episode manifest: %v", err)
		}
		return
	}
	t.Fatal("7sefun missing")
}
