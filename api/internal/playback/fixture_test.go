package playback

import (
	"context"
	"testing"

	"github.com/yueli-official/ytv/api/internal/source"
)

func TestFixtureMayResolveOnlyTheOwnedHLSPath(t *testing.T) {
	catalog := &fakeCatalog{
		definition: source.Definition{Key: "fixture", Kind: source.KindFixture, Enabled: true},
		title:      source.Title{SourceName: "Fixture", Episodes: []source.Episode{{Key: "0", PlaybackURL: "/demo/hls/index.m3u8"}}},
	}
	resolver, _ := New(catalog)
	manifest, err := resolver.Resolve(context.Background(), source.EpisodeLocator{
		Title: source.TitleLocator{SourceKey: "fixture", ExternalID: "42"}, EpisodeKey: "0",
	})
	if err != nil || manifest.URL != "/demo/hls/index.m3u8" || manifest.Engine != EngineHLS {
		t.Fatalf("manifest/error = %#v / %v", manifest, err)
	}
}
