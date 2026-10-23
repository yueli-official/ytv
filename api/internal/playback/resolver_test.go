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

func TestResolverReturnsHLSManifestInsideAllowlist(t *testing.T) {
	catalog := &fakeCatalog{
		definition: source.Definition{Key: "fixture", Name: "Fixture", Enabled: true, PlaybackHosts: []string{"*.media.example"}},
		title:      source.Title{SourceName: "Fixture", Episodes: []source.Episode{{Key: "0", PlaybackURL: "https://cdn.media.example/show/1.m3u8?token=opaque"}}},
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

func TestResolverRejectsPrivateOrUnknownPlaybackHosts(t *testing.T) {
	for _, raw := range []string{"http://127.0.0.1/private.m3u8", "https://unknown.example/video.m3u8"} {
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

func TestWildcardDoesNotMatchBareRoot(t *testing.T) {
	if hostMatches("media.example", "*.media.example") {
		t.Fatal("wildcard unexpectedly matched bare root")
	}
}

func Test7sefunReviewedMediaHostDoesNotAllowOtherHosts(t *testing.T) {
	catalog, err := rulecatalog.LoadKazumiSnapshot("5fea5eb84768a2290ddedb156583ef20d9b43561")
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
		}{{"https://v1.adkwai.com/bs2/adVideoLp/episode.mp4", true}, {"https://v2.adkwai.com/episode.mp4", false}, {"https://v1.adkwai.com.attacker.example/episode.mp4", false}, {"http://v1.adkwai.com/episode.mp4", false}} {
			_, _, err := validateURL(tc.url, d)
			if (err == nil) != tc.allowed {
				t.Errorf("%s allowed=%v error=%v", tc.url, tc.allowed, err)
			}
		}
		return
	}
	t.Fatal("missing 7sefun")
}
