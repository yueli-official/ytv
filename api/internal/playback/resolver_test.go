package playback

import (
	"context"
	"errors"
	"testing"

	"github.com/yueli-official/ytv/api/internal/source"
	"github.com/yueli-official/ytv/api/internal/source/rulecatalog"
)

type fakeCatalog struct {
	definition source.Definition
	title      source.Title
}

func (catalog *fakeCatalog) Definitions() []source.Definition {
	return []source.Definition{catalog.definition}
}
func (catalog *fakeCatalog) ResolveEpisode(_ context.Context, locator source.EpisodeLocator) (source.Title, source.Episode, error) {
	for _, episode := range catalog.title.Episodes {
		if episode.Key == locator.EpisodeKey {
			return catalog.title, episode, nil
		}
	}
	return source.Title{}, source.Episode{}, source.ErrNotFound
}

func TestResolverReturnsCatalogHLSManifest(t *testing.T) {
	catalog := &fakeCatalog{
		definition: source.Definition{Key: "fixture", Name: "Fixture", Enabled: true, PlaybackHosts: []string{"*.media.example"}},
		title:      source.Title{SourceName: "Fixture", Episodes: []source.Episode{{Key: "0", PlaybackURL: "https://vip.ffzy-play7.com/20230101/10759_e65c8db5/index.m3u8"}}},
	}
	resolver, _ := New(catalog)
	manifest, err := resolver.Resolve(context.Background(), source.EpisodeLocator{
		Title: source.TitleLocator{SourceKey: "fixture", ExternalID: "42"}, EpisodeKey: "0",
	})
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Engine != EngineHLS || manifest.SourceName != "Fixture" || manifest.Timeline != "fixture:42" {
		t.Fatalf("manifest = %#v", manifest)
	}
}

func TestResolverRejectsUnsupportedMediaProtocols(t *testing.T) {
	for _, raw := range []string{"javascript:alert(1)", "file:///video.m3u8"} {
		catalog := &fakeCatalog{
			definition: source.Definition{Key: "fixture", Enabled: true, AllowHTTP: true, PlaybackHosts: []string{"media.example"}},
			title:      source.Title{Episodes: []source.Episode{{Key: "0", PlaybackURL: raw}}},
		}
		resolver, _ := New(catalog)
		_, err := resolver.Resolve(context.Background(), source.EpisodeLocator{
			Title: source.TitleLocator{SourceKey: "fixture", ExternalID: "42"}, EpisodeKey: "0",
		})
		if !errors.Is(err, ErrUnsafeURL) {
			t.Fatalf("url %q error = %v", raw, err)
		}
	}
}

func TestApprovedParserAcceptsChangingCDNWithoutHostRegistration(t *testing.T) {
	catalog, err := rulecatalog.LoadKazumiSnapshot("0972545266e7ac7cc06e7e522892239824f7d713")
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range catalog.Definitions {
		if d.Key != "kazumi-7sefun" {
			continue
		}
		for _, tc := range []struct {
			url     string
			allowed bool
		}{
			{"https://v1.adkwai.com/bs2/adVideoLp/episode.mp4", true},
			{"https://v26-cha.bdxiguavod.com/video/tos/episode/?signed=value", true},
			{"https://new-cdn.example/episode.m3u8", true},
			{"http://new-cdn.example/episode.mp4", true},
			{"javascript:alert(1)", false},
			{"file:///episode.mp4", false},
			{"https://127.0.0.1/episode.mp4", true},
			{"https://user:password@cdn.example/episode.mp4", true},
		} {
			_, _, err := validateURL(tc.url, d)
			if (err == nil) != tc.allowed {
				t.Errorf("%s allowed=%v error=%v", tc.url, tc.allowed, err)
			}
		}
		return
	}
	t.Fatal("missing 7sefun")
}
