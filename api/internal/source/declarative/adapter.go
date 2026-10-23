// Package declarative executes format-neutral source plans compiled at deployment time.
package declarative

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/antchfx/htmlquery"
	"github.com/yueli-official/ytv/api/internal/source"
	"github.com/yueli-official/ytv/api/internal/source/browsercapture"
	"github.com/yueli-official/ytv/api/internal/source/rules"
	"github.com/yueli-official/ytv/api/internal/source/verification"
	"golang.org/x/net/html"
)

const maxResponseBytes = 4 << 20

type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

type Adapter struct {
	client         Doer
	clock          func() time.Time
	playbackMu     sync.Mutex
	playbackCache  map[string]cachedPlayback
	browserCapture browsercapture.Resolver
}

func New(client Doer) (*Adapter, error) {
	return NewWithBrowserCapture(client, nil)
}

func NewWithBrowserCapture(client Doer, capture browsercapture.Resolver) (*Adapter, error) {
	if client == nil {
		return nil, fmt.Errorf("declarative source adapter requires an HTTP client")
	}
	return &Adapter{client: client, clock: time.Now, playbackCache: make(map[string]cachedPlayback), browserCapture: capture}, nil
}

func (*Adapter) Browse(context.Context, source.Definition, int) (source.Page, error) {
	return source.Page{}, nil
}

func (adapter *Adapter) Search(ctx context.Context, definition source.Definition, query source.Query) (source.Page, error) {
	plan, err := sourcePlan(definition)
	if err != nil {
		return source.Page{}, err
	}
	keyword := strings.TrimSpace(query.Text)
	if keyword == "" {
		return source.Page{}, fmt.Errorf("%w: search text is required", source.ErrInvalid)
	}
	target := strings.ReplaceAll(plan.Search.Request.URL, "@keyword", url.QueryEscape(keyword))
	document, finalURL, err := adapter.fetch(ctx, definition, plan, plan.Search.Request, target)
	if err != nil {
		return source.Page{}, err
	}
	nodes, err := htmlquery.QueryAll(document, plan.Search.ListXPath)
	if err != nil {
		return source.Page{}, fmt.Errorf("%w: search list XPath failed: %v", source.ErrUnavailable, err)
	}
	items := make([]source.Title, 0, len(nodes))
	for _, node := range nodes {
		name := queryValue(node, scoped(plan.Search.NameXPath))
		link := queryLink(node, scoped(plan.Search.LinkXPath))
		detailURL, resolveErr := resolveURL(finalURL, link)
		if name == "" || resolveErr != nil || !sameHost(definition.Endpoint, detailURL) {
			continue
		}
		items = append(items, source.Title{
			Locator:    source.TitleLocator{SourceKey: definition.Key, ExternalID: encodeURL(detailURL)},
			SourceName: definition.Name, Name: name,
		})
	}
	return source.Page{Items: items, Total: len(items)}, nil
}

func (adapter *Adapter) ResolveTitle(ctx context.Context, definition source.Definition, locator source.TitleLocator) (source.Title, error) {
	plan, err := sourcePlan(definition)
	if err != nil {
		return source.Title{}, err
	}
	detailURL, err := decodeURL(locator.ExternalID)
	if err != nil || locator.SourceKey != definition.Key || !sameHost(definition.Endpoint, detailURL) {
		return source.Title{}, fmt.Errorf("%w: declarative title locator is invalid", source.ErrInvalid)
	}
	document, finalURL, err := adapter.fetch(ctx, definition, plan, rules.RequestPlan{Method: http.MethodGet}, detailURL)
	if err != nil {
		return source.Title{}, err
	}
	episodeNodes, err := episodeNodes(document, plan.Title)
	if err != nil {
		return source.Title{}, fmt.Errorf("%w: episode XPath failed: %v", source.ErrUnavailable, err)
	}
	episodes := make([]source.Episode, 0, len(episodeNodes))
	lineNames := make(map[string]string)
	for _, item := range episodeNodes {
		node := item.node
		link := queryValue(node, ".")
		if href := attribute(node, "href"); href != "" {
			link = href
		}
		episodeURL, resolveErr := resolveURL(finalURL, link)
		if resolveErr != nil || (!sameHost(definition.Endpoint, episodeURL) && !isDirectMedia(episodeURL)) {
			continue
		}
		label := strings.TrimSpace(htmlquery.InnerText(node))
		if label == "" {
			label = "第 " + strconv.Itoa(len(episodes)+1) + " 集"
		}
		if _, ok := lineNames[item.lineKey]; !ok {
			lineNames[item.lineKey] = "线路 " + strconv.Itoa(len(lineNames)+1)
		}
		episodes = append(episodes, source.Episode{Key: encodeURL(episodeURL), Title: label, LineKey: item.lineKey, LineName: lineNames[item.lineKey]})
	}
	if len(episodes) == 0 {
		return source.Title{}, fmt.Errorf("%w: no episodes matched", source.ErrUnavailable)
	}
	name := definition.Name
	if plan.Title.NameXPath != "" {
		if extracted := queryValue(document, plan.Title.NameXPath); extracted != "" {
			name = extracted
		}
	}
	return source.Title{Locator: locator, SourceName: definition.Name, Name: name, EpisodeCount: source.EpisodeTotal(episodes), Episodes: episodes}, nil
}

func (adapter *Adapter) ResolveEpisode(ctx context.Context, definition source.Definition, locator source.EpisodeLocator) (source.Title, source.Episode, error) {
	title, err := adapter.ResolveTitle(ctx, definition, locator.Title)
	if err != nil {
		return source.Title{}, source.Episode{}, err
	}
	var episode source.Episode
	for _, candidate := range title.Episodes {
		if candidate.Key == locator.EpisodeKey {
			episode = candidate
			break
		}
	}
	if episode.Key == "" {
		return source.Title{}, source.Episode{}, fmt.Errorf("%w: episode is missing", source.ErrNotFound)
	}
	episodeURL, err := decodeURL(episode.Key)
	if err != nil {
		return source.Title{}, source.Episode{}, fmt.Errorf("%w: episode URL is invalid", source.ErrInvalid)
	}
	if isDirectMedia(episodeURL) {
		episode.PlaybackURL = episodeURL
		return title, episode, nil
	}
	plan, _ := sourcePlan(definition)
	if plan.Playback.Extractor == "maccms_ecplayer" {
		mediaURL, extractErr := adapter.extractMacCMSECPlayer(ctx, definition, plan, episodeURL)
		if extractErr != nil {
			return source.Title{}, source.Episode{}, extractErr
		}
		episode.PlaybackURL = mediaURL
		return title, episode, nil
	}
	if plan.Playback.Extractor == "maccms_direct" {
		mediaURL, extractErr := adapter.extractMacCMSDirect(ctx, definition, episodeURL)
		if extractErr != nil {
			return source.Title{}, source.Episode{}, extractErr
		}
		episode.PlaybackURL = mediaURL
		return title, episode, nil
	}
	if plan.Playback.Extractor == "html_media" {
		mediaURL, extractErr := adapter.extractHTMLMedia(ctx, definition, episodeURL)
		if extractErr != nil {
			return source.Title{}, source.Episode{}, extractErr
		}
		episode.PlaybackURL = mediaURL
		return title, episode, nil
	}
	if plan.Playback.Extractor == "aniplayer" {
		mediaURL, extractErr := adapter.extractAniPlayer(ctx, definition, plan, episodeURL)
		if extractErr != nil {
			return source.Title{}, source.Episode{}, extractErr
		}
		episode.PlaybackURL = mediaURL
		return title, episode, nil
	}
	if strings.TrimSpace(plan.Playback.MediaXPath) == "" {
		if plan.Playback.UseWebview && adapter.browserCapture != nil {
			captured, captureErr := adapter.browserCapture.Resolve(ctx, browsercapture.Request{
				PageURL: episodeURL, LegacyMode: plan.Playback.UseLegacyParser,
			})
			if captureErr != nil {
				return source.Title{}, source.Episode{}, fmt.Errorf("%w: browser capture failed: %v", source.ErrUnavailable, captureErr)
			}
			episode.PlaybackURL = captured.URL
			return title, episode, nil
		}
		return source.Title{}, source.Episode{}, fmt.Errorf("%w: source needs a deployment media extractor", source.ErrUnavailable)
	}
	document, finalURL, err := adapter.fetch(ctx, definition, plan, rules.RequestPlan{Method: http.MethodGet}, episodeURL)
	if err != nil {
		return source.Title{}, source.Episode{}, err
	}
	mediaURL, err := resolveURL(finalURL, queryValue(document, plan.Playback.MediaXPath))
	if err != nil {
		return source.Title{}, source.Episode{}, fmt.Errorf("%w: media XPath returned no URL", source.ErrUnavailable)
	}
	episode.PlaybackURL = mediaURL
	return title, episode, nil
}

func (adapter *Adapter) Probe(ctx context.Context, definition source.Definition) (source.Health, error) {
	started := adapter.clock()
	plan, err := sourcePlan(definition)
	if err == nil {
		const probeKeyword = "海贼王"
		requestPlan := plan.Search.Request
		requestPlan.URL = strings.ReplaceAll(requestPlan.URL, "@keyword", url.QueryEscape(probeKeyword))
		requestPlan.Body = strings.ReplaceAll(requestPlan.Body, "@keyword", url.QueryEscape(probeKeyword))
		_, _, err = adapter.fetch(ctx, definition, plan, requestPlan, requestPlan.URL)
	}
	health := source.Health{Available: err == nil, CheckedAt: adapter.clock(), Latency: adapter.clock().Sub(started)}
	if err != nil {
		health.Detail = "搜索链路检测失败"
		return health, err
	}
	health.Detail = "搜索链路可用，播放按具体条目解析"
	return health, nil
}

func (adapter *Adapter) fetch(ctx context.Context, definition source.Definition, plan *rules.Plan, requestPlan rules.RequestPlan, rawTarget string) (*html.Node, string, error) {
	target, err := resolveURL(plan.BaseURL, rawTarget)
	if err != nil || !sameHost(definition.Endpoint, target) {
		return nil, "", fmt.Errorf("%w: request left deployment host", source.ErrInvalid)
	}
	method := requestPlan.Method
	if method == "" {
		method = http.MethodGet
	}
	var body io.Reader
	if method == http.MethodPost {
		body = strings.NewReader(requestPlan.Body)
	}
	request, err := http.NewRequestWithContext(ctx, method, target, body)
	if err != nil {
		return nil, "", fmt.Errorf("%w: build declarative request", source.ErrInvalid)
	}
	request.Header.Set("Accept", "text/html,application/xhtml+xml")
	request.Header.Set("User-Agent", first(plan.UserAgent, "YTV/1.0"))
	if requestPlan.Referer != "" {
		request.Header.Set("Referer", requestPlan.Referer)
	}
	response, err := adapter.client.Do(request)
	if err != nil {
		return nil, "", fmt.Errorf("%w: declarative request failed: %v", source.ErrUnavailable, err)
	}
	defer response.Body.Close()
	if response.Request != nil && response.Request.URL != nil && !sameHost(definition.Endpoint, response.Request.URL.String()) {
		return nil, "", fmt.Errorf("%w: redirect left deployment host", source.ErrInvalid)
	}
	if requirement := (verification.Cloudflare{}).Detect(verification.Response{Header: response.Header}); requirement != nil {
		return nil, "", requirement
	}
	// Inspect bounded HTML even on non-2xx responses, before treating it as a
	// status failure. A verified challenge must not become an empty result.
	document, err := htmlquery.Parse(io.LimitReader(response.Body, maxResponseBytes+1))
	if err != nil {
		return nil, "", fmt.Errorf("%w: parse declarative response", source.ErrUnavailable)
	}
	if requirement := (verification.DOM{Kind: verification.Image, All: plan.VerificationXPath}).Detect(verification.Response{Document: document}); requirement != nil {
		return nil, "", requirement
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, "", fmt.Errorf("%w: upstream status %d", source.ErrUnavailable, response.StatusCode)
	}
	finalURL := target
	if response.Request != nil && response.Request.URL != nil {
		finalURL = response.Request.URL.String()
	}
	return document, finalURL, nil
}

func sourcePlan(definition source.Definition) (*rules.Plan, error) {
	if definition.Rule == nil {
		return nil, fmt.Errorf("%w: declarative rule is missing", source.ErrInvalid)
	}
	return definition.Rule, nil
}

type lineEpisodeNode struct {
	node     *html.Node
	lineKey  string
	lineName string
}

func episodeNodes(document *html.Node, plan rules.TitlePlan) ([]lineEpisodeNode, error) {
	roads, err := htmlquery.QueryAll(document, plan.RoadsXPath)
	if err != nil {
		return nil, err
	}
	if len(roads) == 0 {
		roads = []*html.Node{document}
	}
	result := make([]lineEpisodeNode, 0)
	for index, road := range roads {
		nodes, queryErr := htmlquery.QueryAll(road, scoped(plan.EpisodesXPath))
		if queryErr != nil {
			return nil, queryErr
		}
		for _, node := range nodes {
			result = append(result, lineEpisodeNode{node: node, lineKey: strconv.Itoa(index), lineName: "线路 " + strconv.Itoa(index+1)})
		}
		if !plan.MultipleRoads && len(result) > 0 {
			break
		}
	}
	return result, nil
}

func queryValue(node *html.Node, expression string) string {
	matched, err := htmlquery.Query(node, expression)
	if err != nil || matched == nil {
		return ""
	}
	if matched.Type == html.ElementNode {
		return strings.TrimSpace(htmlquery.InnerText(matched))
	}
	return strings.TrimSpace(matched.Data)
}

func queryLink(node *html.Node, expression string) string {
	matched, err := htmlquery.Query(node, expression)
	if err != nil || matched == nil {
		return ""
	}
	if matched.Type == html.ElementNode {
		for _, name := range []string{"href", "src", "data-src"} {
			if value := attribute(matched, name); value != "" {
				return value
			}
		}
		return strings.TrimSpace(htmlquery.InnerText(matched))
	}
	return strings.TrimSpace(matched.Data)
}

func scoped(expression string) string {
	expression = strings.TrimSpace(expression)
	if strings.HasPrefix(expression, "/") {
		return "." + expression
	}
	return expression
}

func attribute(node *html.Node, name string) string {
	if node == nil {
		return ""
	}
	for _, item := range node.Attr {
		if item.Key == name {
			return strings.TrimSpace(item.Val)
		}
	}
	return ""
}

func resolveURL(base, reference string) (string, error) {
	baseURL, err := url.Parse(strings.TrimSpace(base))
	if err != nil || baseURL.Hostname() == "" {
		return "", fmt.Errorf("invalid base URL")
	}
	referenceURL, err := url.Parse(strings.TrimSpace(reference))
	if err != nil || strings.TrimSpace(reference) == "" {
		return "", fmt.Errorf("invalid reference URL")
	}
	return baseURL.ResolveReference(referenceURL).String(), nil
}

func sameHost(base, candidate string) bool {
	left, leftErr := url.Parse(base)
	right, rightErr := url.Parse(candidate)
	return leftErr == nil && rightErr == nil && strings.EqualFold(left.Hostname(), right.Hostname())
}

func encodeURL(raw string) string { return base64.RawURLEncoding.EncodeToString([]byte(raw)) }
func decodeURL(raw string) (string, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	return string(decoded), err
}

func isDirectMedia(raw string) bool {
	parsed, err := url.Parse(raw)
	if err != nil {
		return false
	}
	extension := strings.ToLower(path.Ext(parsed.Path))
	return extension == ".m3u8" || extension == ".mp4" || extension == ".webm"
}

func first(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
