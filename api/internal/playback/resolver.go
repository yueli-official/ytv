// Package playback owns the Episode Locator to Playback Manifest step.
// It only fetches approved playlist metadata, never video segments or media files.
package playback

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/yueli-official/ytv/api/internal/source"
)

var (
	ErrEpisodeNotFound = errors.New("playback episode is not found")
	ErrUnsafeURL       = errors.New("playback URL is not allowed")
)

type Engine string

const (
	EngineHLS    Engine = "hls"
	EngineDirect Engine = "direct"
)

type Manifest struct {
	Locator      source.EpisodeLocator
	SourceName   string
	URL          string
	Engine       Engine
	Timeline     string
	MediaType    string
	PlaylistText string
}

type Catalog interface {
	Definitions() []source.Definition
	ResolveEpisode(context.Context, source.EpisodeLocator) (source.Title, source.Episode, error)
}

type Resolver struct {
	catalog Catalog
}

func New(catalog Catalog) (*Resolver, error) {
	if catalog == nil {
		return nil, errors.New("playback catalog is required")
	}
	return &Resolver{catalog: catalog}, nil
}

func (resolver *Resolver) Resolve(ctx context.Context, locator source.EpisodeLocator) (Manifest, error) {
	definition, ok := resolver.definition(locator.Title.SourceKey)
	if !ok || strings.TrimSpace(locator.Title.ExternalID) == "" || strings.TrimSpace(locator.EpisodeKey) == "" {
		return Manifest{}, fmt.Errorf("%w: invalid locator", ErrEpisodeNotFound)
	}
	title, episode, err := resolver.catalog.ResolveEpisode(ctx, locator)
	if err != nil {
		return Manifest{}, err
	}
	mediaURL, engine, err := validateURL(episode.PlaybackURL, definition)
	if err != nil {
		return Manifest{}, err
	}
	playlistText := ""
	if definition.Key == "kazumi-7sefun" && mediaURL.Scheme == "https" && mediaURL.Host == "yun.92cj.com" && strings.HasPrefix(mediaURL.Path, "/video/dmhls/") && engine == EngineHLS {
		playlistText, err = readPlaylist(ctx, mediaURL)
		if err != nil {
			return Manifest{}, err
		}
	}
	return Manifest{
		Locator: locator, SourceName: title.SourceName, URL: mediaURL.String(), Engine: engine,
		Timeline:     locator.Title.SourceKey + ":" + locator.Title.ExternalID,
		MediaType:    mediaType(mediaURL, engine),
		PlaylistText: playlistText,
	}, nil
}

func mediaType(target *url.URL, engine Engine) string {
	if engine == EngineHLS {
		return "application/vnd.apple.mpegurl"
	}
	if strings.HasSuffix(strings.ToLower(target.Path), ".mp4") {
		return "video/mp4"
	}
	if strings.HasSuffix(strings.ToLower(target.Path), ".webm") {
		return "video/webm"
	}
	return ""
}

func (resolver *Resolver) definition(key string) (source.Definition, bool) {
	for _, definition := range resolver.catalog.Definitions() {
		if definition.Enabled && definition.Key == key {
			return definition, true
		}
	}
	return source.Definition{}, false
}

func validateURL(raw string, definition source.Definition) (*url.URL, Engine, error) {
	if len(raw) == 0 || len(raw) > 4096 {
		return nil, "", ErrUnsafeURL
	}
	if definition.Kind == source.KindFixture && strings.HasPrefix(raw, "/demo/hls/") && strings.HasSuffix(strings.ToLower(raw), ".m3u8") {
		parsed, err := url.Parse(raw)
		if err != nil {
			return nil, "", ErrUnsafeURL
		}
		return parsed, EngineHLS, nil
	}
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Hostname() == "" {
		return nil, "", ErrUnsafeURL
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return nil, "", ErrUnsafeURL
	}
	// Catalog adapters own the resolved URL. CDN hosts are not a playback policy.
	engine := EngineDirect
	if strings.HasSuffix(strings.ToLower(parsed.Path), ".m3u8") {
		engine = EngineHLS
	}
	return parsed, engine, nil
}
