package trending

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const maxResponseBytes = 1 << 20

var doubanPosterFile = regexp.MustCompile(`^p[0-9]{6,20}\.(?:jpg|webp)$`)

var ErrInvalid = errors.New("trending query is invalid")

var allowedTags = map[string]bool{
	"热门": true, "美剧": true, "英剧": true, "日剧": true,
	"韩剧": true, "港剧": true, "综艺": true, "纪录片": true,
}

type Query struct {
	Type  string
	Tag   string
	Page  int
	Limit int
}

type Item struct {
	ID       string
	Title    string
	Score    string
	CoverURL string
	IsNew    bool
}

type Page struct {
	Query Query
	Items []Item
}

type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

type cacheEntry struct {
	page      Page
	expiresAt time.Time
}

type Service struct {
	client Doer
	clock  func() time.Time
	mu     sync.Mutex
	cache  map[string]cacheEntry
}

func New(client Doer, clock func() time.Time) (*Service, error) {
	if client == nil {
		return nil, errors.New("ytv/trending: HTTP client is required")
	}
	if clock == nil {
		clock = time.Now
	}
	return &Service{client: client, clock: clock, cache: map[string]cacheEntry{}}, nil
}

func (service *Service) List(ctx context.Context, query Query) (Page, error) {
	query.Type = strings.TrimSpace(query.Type)
	if query.Type == "" {
		query.Type = "movie"
	}
	if query.Type != "movie" && query.Type != "tv" {
		return Page{}, fmt.Errorf("%w: type must be movie or tv", ErrInvalid)
	}
	query.Tag = strings.TrimSpace(query.Tag)
	if query.Tag == "" {
		query.Tag = "热门"
	}
	if !allowedTags[query.Tag] {
		return Page{}, fmt.Errorf("%w: tag is not supported", ErrInvalid)
	}
	if query.Page < 1 {
		query.Page = 1
	}
	if query.Limit == 0 {
		query.Limit = 12
	}
	if query.Limit < 1 || query.Limit > 24 {
		return Page{}, fmt.Errorf("%w: limit must be between 1 and 24", ErrInvalid)
	}
	key := strings.Join([]string{query.Type, query.Tag, strconv.Itoa(query.Page), strconv.Itoa(query.Limit)}, "\x00")
	service.mu.Lock()
	if cached, ok := service.cache[key]; ok && service.clock().Before(cached.expiresAt) {
		service.mu.Unlock()
		return cached.page, nil
	}
	service.mu.Unlock()

	endpoint, _ := url.Parse("https://movie.douban.com/j/search_subjects")
	values := endpoint.Query()
	values.Set("type", query.Type)
	values.Set("tag", query.Tag)
	values.Set("sort", "recommend")
	values.Set("page_limit", strconv.Itoa(query.Limit))
	values.Set("page_start", strconv.Itoa((query.Page-1)*query.Limit))
	endpoint.RawQuery = values.Encode()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return Page{}, fmt.Errorf("build trending request: %w", err)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "Mozilla/5.0")
	response, err := service.client.Do(request)
	if err != nil {
		return Page{}, fmt.Errorf("request trending provider: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return Page{}, fmt.Errorf("trending provider status %d", response.StatusCode)
	}
	var payload struct {
		Subjects []struct {
			ID    string `json:"id"`
			Title string `json:"title"`
			Rate  string `json:"rate"`
			Cover string `json:"cover"`
			IsNew bool   `json:"is_new"`
		} `json:"subjects"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, maxResponseBytes+1)).Decode(&payload); err != nil {
		return Page{}, fmt.Errorf("decode trending response: %w", err)
	}
	page := Page{Query: query, Items: make([]Item, 0, len(payload.Subjects))}
	for _, subject := range payload.Subjects {
		if strings.TrimSpace(subject.ID) == "" || strings.TrimSpace(subject.Title) == "" {
			continue
		}
		page.Items = append(page.Items, Item{
			ID: strings.TrimSpace(subject.ID), Title: strings.TrimSpace(subject.Title),
			Score: strings.TrimSpace(subject.Rate), CoverURL: safeDoubanCoverPath(subject.Cover), IsNew: subject.IsNew,
		})
	}
	service.mu.Lock()
	if len(service.cache) >= 32 {
		service.cache = map[string]cacheEntry{}
	}
	service.cache[key] = cacheEntry{page: page, expiresAt: service.clock().Add(30 * time.Minute)}
	service.mu.Unlock()
	return page, nil
}

func safeDoubanCoverPath(raw string) string {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Scheme != "https" || parsed.User != nil || parsed.Port() != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return ""
	}
	host := strings.ToLower(parsed.Hostname())
	if len(host) != len("img1.doubanio.com") || !strings.HasPrefix(host, "img") || host[3] < '1' || host[3] > '9' || host[4:] != ".doubanio.com" {
		return ""
	}
	file := path.Base(parsed.Path)
	if parsed.Path != "/view/photo/s_ratio_poster/public/"+file || !doubanPosterFile.MatchString(file) {
		return ""
	}
	return "/external-poster/douban/" + file
}
