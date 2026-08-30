// Package htmlcatalog implements the deployment-owned HTML catalog protocol
// used by the legacy Omo-style source. Selectors and paths stay inside this
// Adapter; callers only see SourceRegistry types.
package htmlcatalog

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/yueli-official/ytv/api/internal/source"
	"golang.org/x/net/html"
)

const maxResponseBytes = 4 << 20

var (
	detailIDPattern = regexp.MustCompile(`/vod/detail/id/([^/.]+)\.html`)
	playURLPattern  = regexp.MustCompile(`"url"\s*:\s*"([^"]+)"`)
	vodNamePattern  = regexp.MustCompile(`var\s+vod_name\s*=\s*'([^']+)'`)
)

type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

type Adapter struct {
	client Doer
	clock  func() time.Time
}

func New(client Doer) (*Adapter, error) {
	if client == nil {
		return nil, fmt.Errorf("html catalog adapter requires an HTTP client")
	}
	return &Adapter{client: client, clock: time.Now}, nil
}

func (adapter *Adapter) Browse(ctx context.Context, definition source.Definition, page int) (source.Page, error) {
	document, err := adapter.fetch(ctx, definition, "/")
	if err != nil {
		return source.Page{}, err
	}
	return source.Page{Items: parseCards(definition, document)}, nil
}

func (adapter *Adapter) Search(ctx context.Context, definition source.Definition, query source.Query) (source.Page, error) {
	text := strings.TrimSpace(query.Text)
	if text == "" {
		return source.Page{}, fmt.Errorf("%w: search text is required", source.ErrInvalid)
	}
	page := query.Page
	if page < 1 {
		page = 1
	}
	path := fmt.Sprintf("/vod/search/page/%d/wd/%s.html", page, url.PathEscape(text))
	document, err := adapter.fetch(ctx, definition, path)
	if err != nil {
		return source.Page{}, err
	}
	items := parseCards(definition, document)
	return source.Page{Items: items, Total: len(items)}, nil
}

func (adapter *Adapter) ResolveTitle(ctx context.Context, definition source.Definition, locator source.TitleLocator) (source.Title, error) {
	if locator.SourceKey != definition.Key || !validID(locator.ExternalID) {
		return source.Title{}, fmt.Errorf("%w: HTML title locator is invalid", source.ErrInvalid)
	}
	document, err := adapter.fetch(ctx, definition, "/vod/detail/id/"+url.PathEscape(locator.ExternalID)+".html")
	if err != nil {
		return source.Title{}, err
	}
	title := parseDetail(definition, locator, document)
	if title.Name == "" || len(title.Episodes) == 0 {
		return source.Title{}, fmt.Errorf("%w: HTML title structure changed", source.ErrUnavailable)
	}
	return title, nil
}

func (adapter *Adapter) ResolveEpisode(ctx context.Context, definition source.Definition, locator source.EpisodeLocator) (source.Title, source.Episode, error) {
	if locator.Title.SourceKey != definition.Key || !validID(locator.Title.ExternalID) {
		return source.Title{}, source.Episode{}, fmt.Errorf("%w: HTML episode locator is invalid", source.ErrInvalid)
	}
	episodeIndex, err := strconv.Atoi(locator.EpisodeKey)
	if err != nil || episodeIndex < 0 {
		return source.Title{}, source.Episode{}, fmt.Errorf("%w: HTML episode key is invalid", source.ErrInvalid)
	}
	path := fmt.Sprintf("/vod/play/id/%s/sid/8/nid/%d.html", url.PathEscape(locator.Title.ExternalID), episodeIndex+1)
	document, err := adapter.fetch(ctx, definition, path)
	if err != nil {
		return source.Title{}, source.Episode{}, err
	}
	title := parseDetail(definition, locator.Title, document)
	if title.Name == "" {
		title.Name = scriptMatch(document, vodNamePattern)
	}
	if title.Name == "" {
		return source.Title{}, source.Episode{}, fmt.Errorf("%w: HTML playback title is missing", source.ErrUnavailable)
	}
	playbackURL := scriptMatch(document, playURLPattern)
	if playbackURL == "" {
		return source.Title{}, source.Episode{}, fmt.Errorf("%w: HTML playback URL is missing", source.ErrUnavailable)
	}
	playbackURL = strings.ReplaceAll(playbackURL, `\/`, "/")
	var episode source.Episode
	for _, candidate := range title.Episodes {
		if candidate.Key == locator.EpisodeKey {
			episode = candidate
			break
		}
	}
	if episode.Key == "" {
		episode = source.Episode{Key: locator.EpisodeKey, Title: fmt.Sprintf("第 %d 集", episodeIndex+1)}
	}
	episode.PlaybackURL = playbackURL
	return title, episode, nil
}

func (adapter *Adapter) Probe(ctx context.Context, definition source.Definition) (source.Health, error) {
	started := adapter.clock()
	_, err := adapter.fetch(ctx, definition, "/")
	health := source.Health{Available: err == nil, CheckedAt: adapter.clock(), Latency: adapter.clock().Sub(started)}
	if err != nil {
		health.Detail = "HTML catalog request failed"
		return health, err
	}
	health.Detail = "HTML catalog responded"
	return health, nil
}

func (adapter *Adapter) fetch(ctx context.Context, definition source.Definition, path string) (*html.Node, error) {
	base, err := url.Parse(definition.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("%w: parse HTML endpoint", source.ErrInvalid)
	}
	reference, err := url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("%w: parse HTML path", source.ErrInvalid)
	}
	target := base.ResolveReference(reference)
	if !strings.EqualFold(target.Hostname(), base.Hostname()) {
		return nil, fmt.Errorf("%w: HTML request left deployment host", source.ErrInvalid)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("%w: build HTML request", source.ErrInvalid)
	}
	request.Header.Set("Accept", "text/html,application/xhtml+xml")
	request.Header.Set("User-Agent", "YTV/1.0")
	response, err := adapter.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("%w: HTML request failed", source.ErrUnavailable)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("%w: HTML upstream status %d", source.ErrUnavailable, response.StatusCode)
	}
	document, err := html.Parse(io.LimitReader(response.Body, maxResponseBytes+1))
	if err != nil {
		return nil, fmt.Errorf("%w: parse HTML response", source.ErrUnavailable)
	}
	return document, nil
}

func parseCards(definition source.Definition, document *html.Node) []source.Title {
	items := make([]source.Title, 0)
	walk(document, func(node *html.Node) {
		if node.Type != html.ElementNode || !hasClass(node, "module-card-item") {
			return
		}
		anchor := findFirst(node, func(candidate *html.Node) bool {
			return candidate.Type == html.ElementNode && candidate.Data == "a" && hasClass(candidate, "module-card-item-poster")
		})
		match := detailIDPattern.FindStringSubmatch(attribute(anchor, "href"))
		if len(match) != 2 {
			return
		}
		name := textOf(findFirst(node, func(candidate *html.Node) bool { return hasClass(candidate, "module-card-item-title") }))
		if strong := findFirst(node, func(candidate *html.Node) bool {
			return candidate.Type == html.ElementNode && candidate.Data == "strong"
		}); strong != nil {
			name = textOf(strong)
		}
		if name == "" {
			return
		}
		image := findFirst(node, func(candidate *html.Node) bool { return candidate.Type == html.ElementNode && candidate.Data == "img" })
		imageURL := attribute(image, "data-original")
		if imageURL == "" {
			imageURL = attribute(image, "src")
		}
		items = append(items, source.Title{
			Locator:    source.TitleLocator{SourceKey: definition.Key, ExternalID: match[1]},
			SourceName: definition.Name, Name: name,
			Category: textOf(findFirst(node, func(candidate *html.Node) bool { return hasClass(candidate, "module-card-item-class") })),
			Remarks:  textOf(findFirst(node, func(candidate *html.Node) bool { return hasClass(candidate, "module-item-note") })),
			ImageURL: source.AllowedImageURL(definition, imageURL),
		})
	})
	return items
}

func parseDetail(definition source.Definition, locator source.TitleLocator, document *html.Node) source.Title {
	title := source.Title{Locator: locator, SourceName: definition.Name}
	if heading := findFirst(document, func(node *html.Node) bool {
		return node.Type == html.ElementNode && node.Data == "h1"
	}); heading != nil {
		title.Name = textOf(heading)
	}
	if title.Name == "" {
		title.Name = scriptMatch(document, vodNamePattern)
	}
	walk(document, func(node *html.Node) {
		if node.Type != html.ElementNode || node.Data != "a" || !hasClass(node, "module-play-list-link") {
			return
		}
		href := attribute(node, "href")
		marker := "/nid/"
		position := strings.Index(href, marker)
		if position < 0 {
			return
		}
		value := strings.TrimSuffix(href[position+len(marker):], ".html")
		nid, err := strconv.Atoi(value)
		if err != nil || nid < 1 {
			return
		}
		title.Episodes = append(title.Episodes, source.Episode{Key: strconv.Itoa(nid - 1), Title: textOf(node)})
	})
	title.EpisodeCount = len(title.Episodes)
	return title
}

func validID(value string) bool {
	if value = strings.TrimSpace(value); value == "" || len(value) > 100 {
		return false
	}
	for _, current := range value {
		if (current < 'a' || current > 'z') && (current < 'A' || current > 'Z') && (current < '0' || current > '9') && current != '-' && current != '_' {
			return false
		}
	}
	return true
}

func walk(node *html.Node, visit func(*html.Node)) {
	if node == nil {
		return
	}
	visit(node)
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		walk(child, visit)
	}
}

func findFirst(node *html.Node, match func(*html.Node) bool) *html.Node {
	var result *html.Node
	walk(node, func(candidate *html.Node) {
		if result == nil && match(candidate) {
			result = candidate
		}
	})
	return result
}

func hasClass(node *html.Node, className string) bool {
	for _, value := range strings.Fields(attribute(node, "class")) {
		if value == className {
			return true
		}
	}
	return false
}

func attribute(node *html.Node, name string) string {
	if node == nil {
		return ""
	}
	for _, attribute := range node.Attr {
		if attribute.Key == name {
			return strings.TrimSpace(attribute.Val)
		}
	}
	return ""
}

func textOf(node *html.Node) string {
	if node == nil {
		return ""
	}
	var builder strings.Builder
	walk(node, func(candidate *html.Node) {
		if candidate.Type == html.TextNode {
			builder.WriteString(" ")
			builder.WriteString(candidate.Data)
		}
	})
	return strings.Join(strings.Fields(builder.String()), " ")
}

func scriptMatch(document *html.Node, pattern *regexp.Regexp) string {
	var result string
	walk(document, func(node *html.Node) {
		if result != "" || node.Type != html.ElementNode || node.Data != "script" {
			return
		}
		match := pattern.FindStringSubmatch(textOf(node))
		if len(match) == 2 {
			if pattern == playURLPattern {
				var decoded string
				if json.Unmarshal([]byte(`"`+match[1]+`"`), &decoded) == nil {
					result = decoded
					return
				}
			}
			result = match[1]
		}
	})
	return strings.TrimSpace(result)
}
