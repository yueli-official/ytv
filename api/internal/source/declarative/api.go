package declarative

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/yueli-official/ytv/api/internal/source"
	"github.com/yueli-official/ytv/api/internal/source/rules"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func (adapter *Adapter) requestAPI(ctx context.Context, plan *rules.Plan, spec rules.APIRequest, variable, value string) (any, error) {
	target, err := url.Parse(strings.ReplaceAll(spec.URL, variable, url.PathEscape(value)))
	if err != nil {
		return nil, err
	}
	query := target.Query()
	for k, v := range spec.Query {
		query.Set(k, strings.ReplaceAll(fmt.Sprint(v), variable, value))
	}
	target.RawQuery = query.Encode()
	body, err := json.Marshal(substitute(spec.Body, variable, value))
	if err != nil {
		return nil, err
	}
	method := spec.Method
	if method == "" {
		method = http.MethodGet
	}
	var requestBody io.Reader
	if method == http.MethodPost {
		requestBody = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, target.String(), requestBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", first(plan.UserAgent, "YTV/1.0"))
	if method == http.MethodPost {
		req.Header.Set("Content-Type", "application/json")
	}
	for key, value := range spec.Headers {
		req.Header.Set(key, value)
	}
	response, err := adapter.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: rule API request: %v", source.ErrUnavailable, err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("%w: rule API status %d", source.ErrUnavailable, response.StatusCode)
	}
	var document any
	decoder := json.NewDecoder(io.LimitReader(response.Body, maxResponseBytes+1))
	decoder.UseNumber()
	if err := decoder.Decode(&document); err != nil {
		return nil, fmt.Errorf("%w: rule API JSON: %v", source.ErrUnavailable, err)
	}
	return document, nil
}

func substitute(value any, variable, replacement string) any {
	switch v := value.(type) {
	case string:
		return strings.ReplaceAll(v, variable, replacement)
	case map[string]any:
		result := map[string]any{}
		for k, item := range v {
			result[k] = substitute(item, variable, replacement)
		}
		return result
	case []any:
		result := make([]any, len(v))
		for i, item := range v {
			result[i] = substitute(item, variable, replacement)
		}
		return result
	default:
		return value
	}
}

// The approved catalogs use property paths and a terminal array wildcard.
func jsonValue(document any, path string) any {
	path = strings.TrimSuffix(path, "[*]")
	if path == "$" {
		return document
	}
	if !strings.HasPrefix(path, "$.") {
		return nil
	}
	for _, key := range strings.Split(strings.TrimPrefix(path, "$."), ".") {
		object, ok := document.(map[string]any)
		if !ok {
			return nil
		}
		document = object[key]
	}
	return document
}
func jsonText(value any) string {
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v)
	case json.Number:
		return v.String()
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	default:
		return ""
	}
}

func (adapter *Adapter) searchAPI(ctx context.Context, definition source.Definition, plan *rules.Plan, keyword string) (source.Page, error) {
	spec := plan.Search.API
	document, err := adapter.requestAPI(ctx, plan, spec.Request, "@keyword", keyword)
	if err != nil {
		return source.Page{}, err
	}
	entries, ok := jsonValue(document, spec.ListPath).([]any)
	if !ok {
		return source.Page{}, fmt.Errorf("%w: rule API result list missing", source.ErrUnavailable)
	}
	items := []source.Title{}
	for _, entry := range entries {
		name := jsonText(jsonValue(entry, spec.NamePath))
		id := jsonText(jsonValue(entry, spec.SourcePath))
		if name == "" || id == "" {
			continue
		}
		target := id
		if plan.Title.API == nil {
			if spec.SourceTemplate != "" {
				target = strings.ReplaceAll(spec.SourceTemplate, "@source", url.PathEscape(id))
			} else {
				target = strings.TrimRight(plan.BaseURL, "/") + "/" + url.PathEscape(id) + ".html"
			}
			if !sameHost(definition.Endpoint, target) {
				continue
			}
		}
		items = append(items, source.Title{Name: name, SourceName: definition.Name, Locator: source.TitleLocator{SourceKey: definition.Key, ExternalID: encodeURL(target)}})
	}
	return source.Page{Items: items, Total: len(items), TotalPages: 1}, nil
}

func (adapter *Adapter) titleAPI(ctx context.Context, definition source.Definition, plan *rules.Plan, locator source.TitleLocator) (source.Title, error) {
	id, err := decodeURL(locator.ExternalID)
	if err != nil || id == "" || locator.SourceKey != definition.Key {
		return source.Title{}, fmt.Errorf("%w: rule API locator", source.ErrInvalid)
	}
	spec := plan.Title.API
	document, err := adapter.requestAPI(ctx, plan, spec.Request, "@source", id)
	if err != nil {
		return source.Title{}, err
	}
	if spec.Format == "nested" {
		return nestedTitle(definition, locator, spec, document, id)
	}
	roads := strings.Split(jsonText(jsonValue(document, spec.RoadEpisodesPath)), spec.RoadSeparator)
	names := strings.Split(jsonText(jsonValue(document, spec.RoadNamesPath)), spec.RoadSeparator)
	episodes := []source.Episode{}
	for i, road := range roads {
		line := strconv.Itoa(i)
		name := "线路 " + strconv.Itoa(i+1)
		if i < len(names) && names[i] != "" {
			name = names[i]
		}
		for _, raw := range strings.Split(road, spec.EpisodeSeparator) {
			parts := strings.SplitN(raw, spec.FieldSeparator, 2)
			if len(parts) != 2 {
				continue
			}
			target := strings.TrimSpace(parts[1])
			if !isDirectMedia(target) && !sameHost(definition.Endpoint, target) {
				continue
			}
			episodes = append(episodes, source.Episode{Key: encodeURL(target), Title: parts[0], LineKey: line, LineName: name})
		}
	}
	if len(episodes) == 0 {
		return source.Title{}, fmt.Errorf("%w: rule API has no episodes", source.ErrUnavailable)
	}
	return source.Title{Locator: locator, Name: definition.Name, SourceName: definition.Name, Episodes: episodes, EpisodeCount: source.EpisodeTotal(episodes)}, nil
}

func nestedTitle(definition source.Definition, locator source.TitleLocator, spec *rules.ChapterAPI, document any, id string) (source.Title, error) {
	value := jsonValue(document, spec.RoadsPath)
	roads := []any{value}
	if strings.HasSuffix(spec.RoadsPath, "[*]") {
		var ok bool
		roads, ok = value.([]any)
		if !ok {
			return source.Title{}, fmt.Errorf("%w: API roads missing", source.ErrUnavailable)
		}
	}
	episodes := []source.Episode{}
	for i, road := range roads {
		items, ok := jsonValue(road, spec.EpisodesPath).([]any)
		if !ok {
			continue
		}
		name := jsonText(jsonValue(road, spec.RoadNamePath))
		if name == "" {
			name = "线路 " + strconv.Itoa(i+1)
		}
		for _, item := range items {
			episodeID := jsonText(jsonValue(item, spec.EpisodeURLPath))
			label := jsonText(jsonValue(item, spec.EpisodeNamePath))
			if episodeID == "" {
				continue
			}
			target := strings.ReplaceAll(strings.ReplaceAll(spec.EpisodePage.URL, "@source", url.PathEscape(id)), "@episodeUrl", url.PathEscape(episodeID))
			u, err := url.Parse(target)
			if err != nil || !sameHost(definition.Endpoint, target) {
				continue
			}
			q := u.Query()
			for k, v := range spec.EpisodePage.Query {
				q.Set(k, strings.ReplaceAll(strings.ReplaceAll(fmt.Sprint(v), "@source", id), "@episodeUrl", episodeID))
			}
			u.RawQuery = q.Encode()
			episodes = append(episodes, source.Episode{Key: encodeURL(u.String()), Title: label, LineKey: strconv.Itoa(i), LineName: name})
		}
	}
	if len(episodes) == 0 {
		return source.Title{}, fmt.Errorf("%w: API episodes missing", source.ErrUnavailable)
	}
	return source.Title{Locator: locator, Name: definition.Name, SourceName: definition.Name, Episodes: episodes, EpisodeCount: source.EpisodeTotal(episodes)}, nil
}
