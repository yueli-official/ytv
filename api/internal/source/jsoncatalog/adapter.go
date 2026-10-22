package jsoncatalog

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/yueli-official/ytv/api/internal/source"
)

const maxResponseBytes = 4 << 20

type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

type Adapter struct {
	client Doer
	clock  func() time.Time
}

func New(client Doer) (*Adapter, error) {
	if client == nil {
		return nil, fmt.Errorf("json catalog adapter requires an HTTP client")
	}
	return &Adapter{client: client, clock: time.Now}, nil
}

func (adapter *Adapter) Browse(ctx context.Context, definition source.Definition, page int) (source.Page, error) {
	if page < 1 {
		page = 1
	}
	response, err := adapter.fetch(ctx, definition, map[string]string{
		"ac": "videolist", "pg": strconv.Itoa(page),
	})
	if err != nil {
		return source.Page{}, err
	}
	return mapPage(definition, response, page), nil
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
	response, err := adapter.fetch(ctx, definition, map[string]string{
		"ac": "videolist", "wd": text, "pg": strconv.Itoa(page),
	})
	if err != nil {
		return source.Page{}, err
	}
	return mapPage(definition, response, page), nil
}

func mapPage(definition source.Definition, response catalogResponse, page int) source.Page {
	items := make([]source.Title, 0, len(response.List))
	for _, item := range response.List {
		title := mapTitle(definition, item)
		if title.Name == "" || title.Locator.ExternalID == "" || title.EpisodeCount == 0 {
			continue
		}
		items = append(items, title)
	}
	next := 0
	if page < response.PageCount {
		next = page + 1
	}
	return source.Page{Items: items, Total: response.Total, NextPage: next}
}

func (adapter *Adapter) ResolveTitle(ctx context.Context, definition source.Definition, locator source.TitleLocator) (source.Title, error) {
	if strings.TrimSpace(locator.ExternalID) == "" || locator.SourceKey != definition.Key {
		return source.Title{}, fmt.Errorf("%w: source title locator is invalid", source.ErrInvalid)
	}
	response, err := adapter.fetch(ctx, definition, map[string]string{"ac": "videolist", "ids": locator.ExternalID})
	if err != nil {
		return source.Title{}, err
	}
	if len(response.List) == 0 {
		return source.Title{}, fmt.Errorf("%w: %s/%s", source.ErrNotFound, locator.SourceKey, locator.ExternalID)
	}
	title := mapTitle(definition, response.List[0])
	if title.Locator.ExternalID != locator.ExternalID {
		return source.Title{}, fmt.Errorf("%w: upstream returned another title", source.ErrUnavailable)
	}
	return title, nil
}

func (adapter *Adapter) ResolveEpisode(ctx context.Context, definition source.Definition, locator source.EpisodeLocator) (source.Title, source.Episode, error) {
	title, err := adapter.ResolveTitle(ctx, definition, locator.Title)
	if err != nil {
		return source.Title{}, source.Episode{}, err
	}
	for _, episode := range title.Episodes {
		if episode.Key == locator.EpisodeKey {
			return title, episode, nil
		}
	}
	return source.Title{}, source.Episode{}, fmt.Errorf("%w: episode %s", source.ErrNotFound, locator.EpisodeKey)
}

func (adapter *Adapter) Probe(ctx context.Context, definition source.Definition) (source.Health, error) {
	started := adapter.clock()
	_, err := adapter.fetch(ctx, definition, map[string]string{"ac": "list"})
	health := source.Health{Available: err == nil, CheckedAt: adapter.clock(), Latency: adapter.clock().Sub(started)}
	if err != nil {
		health.Detail = "catalog request failed"
		return health, err
	}
	health.Detail = "catalog responded"
	return health, nil
}

func (adapter *Adapter) fetch(ctx context.Context, definition source.Definition, query map[string]string) (catalogResponse, error) {
	endpoint, err := url.Parse(definition.Endpoint)
	if err != nil {
		return catalogResponse{}, fmt.Errorf("%w: parse endpoint", source.ErrInvalid)
	}
	values := endpoint.Query()
	for key, value := range query {
		values.Set(key, value)
	}
	endpoint.RawQuery = values.Encode()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return catalogResponse{}, fmt.Errorf("%w: build request", source.ErrInvalid)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "YTV/1.0")
	response, err := adapter.client.Do(request)
	if err != nil {
		return catalogResponse{}, fmt.Errorf("%w: %v", source.ErrUnavailable, err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return catalogResponse{}, fmt.Errorf("%w: upstream status %d", source.ErrUnavailable, response.StatusCode)
	}
	decoder := json.NewDecoder(io.LimitReader(response.Body, maxResponseBytes+1))
	decoder.UseNumber()
	var payload catalogResponse
	if err := decoder.Decode(&payload); err != nil {
		return catalogResponse{}, fmt.Errorf("%w: decode catalog response: %v", source.ErrUnavailable, err)
	}
	return payload, nil
}

type catalogResponse struct {
	PageCount int       `json:"pagecount"`
	Total     int       `json:"total"`
	List      []vodItem `json:"list"`
}

type scalar string

func (value *scalar) UnmarshalJSON(data []byte) error {
	text := strings.TrimSpace(string(data))
	if text == "null" || text == "" {
		*value = ""
		return nil
	}
	if strings.HasPrefix(text, "\"") {
		var decoded string
		if err := json.Unmarshal(data, &decoded); err != nil {
			return err
		}
		*value = scalar(decoded)
		return nil
	}
	*value = scalar(text)
	return nil
}

type vodItem struct {
	ID       scalar `json:"vod_id"`
	Name     string `json:"vod_name"`
	Year     scalar `json:"vod_year"`
	Category string `json:"type_name"`
	Area     string `json:"vod_area"`
	Language string `json:"vod_lang"`
	Score    scalar `json:"vod_douban_score"`
	Remarks  string `json:"vod_remarks"`
	Summary  string `json:"vod_blurb"`
	ImageURL string `json:"vod_pic"`
	PlayURL  string `json:"vod_play_url"`
	PlayFrom string `json:"vod_play_from"`
}

func mapTitle(definition source.Definition, item vodItem) source.Title {
	episodes := parseLines(item.PlayURL, item.PlayFrom)
	return source.Title{
		Locator:    source.TitleLocator{SourceKey: definition.Key, ExternalID: strings.TrimSpace(string(item.ID))},
		SourceName: definition.Name,
		Name:       strings.TrimSpace(item.Name), Year: strings.TrimSpace(string(item.Year)),
		Category: strings.TrimSpace(item.Category), Area: strings.TrimSpace(item.Area),
		Language: strings.TrimSpace(item.Language), Score: strings.TrimSpace(string(item.Score)), Remarks: strings.TrimSpace(item.Remarks),
		Summary: strings.TrimSpace(item.Summary), ImageURL: source.AllowedImageURL(definition, item.ImageURL),
		EpisodeCount: source.EpisodeTotal(episodes), Episodes: episodes,
	}
}

func parseEpisodes(raw string) []source.Episode {
	parts := strings.Split(raw, "#")
	result := make([]source.Episode, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		fields := strings.SplitN(part, "$", 2)
		title := ""
		playbackURL := strings.TrimSpace(fields[0])
		if len(fields) == 2 {
			title = strings.TrimSpace(fields[0])
			playbackURL = strings.TrimSpace(fields[1])
		}
		if playbackURL == "" {
			continue
		}
		result = append(result, source.Episode{
			Key: strconv.Itoa(len(result)), Title: title, PlaybackURL: playbackURL,
		})
	}
	return result
}

// MacCMS separates lines with $$$, episodes with # and labels with $.
// Keep the first line's numeric keys so existing links and progress still resolve.
func parseLines(raw, names string) []source.Episode {
	result := make([]source.Episode, 0)
	labels := strings.Split(names, "$$$")
	for index, line := range strings.Split(raw, "$$$") {
		key := strconv.Itoa(index)
		name := "线路 " + strconv.Itoa(index+1)
		if index < len(labels) && strings.TrimSpace(labels[index]) != "" {
			name = strings.TrimSpace(labels[index])
		}
		for _, episode := range parseEpisodes(line) {
			if index > 0 {
				episode.Key = "line" + key + "-" + episode.Key
			}
			episode.LineKey, episode.LineName = key, name
			result = append(result, episode)
		}
	}
	return result
}
