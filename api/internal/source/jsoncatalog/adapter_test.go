package jsoncatalog

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/yueli-official/ytv/api/internal/source"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) Do(request *http.Request) (*http.Response, error) {
	return function(request)
}

func TestAdapterMapsCatalogWithoutLeakingProtocolFields(t *testing.T) {
	client := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		query := request.URL.Query()
		if query.Get("ac") == "list" {
			if query.Get("wd") != "测试" || query.Get("pg") != "2" {
				t.Fatalf("search query = %s", request.URL.RawQuery)
			}
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"pagecount":3,"total":7,"list":[{"vod_id":"42","vod_name":"演示影片"}]}`))}, nil
		}
		if query.Get("ac") != "videolist" || query.Get("ids") != "42" || query.Has("wd") || query.Has("pg") {
			t.Fatalf("query = %s", request.URL.RawQuery)
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{
			"pagecount":5376,"total":107520,"list":[{"vod_id":"42","vod_name":"演示影片","vod_year":2026,
			"type_name":"科幻","vod_area":"测试地区","vod_lang":"中文","vod_remarks":"更新中",
			"vod_douban_score":"8.6",
			"vod_blurb":"合成内容","vod_pic":"https://images.example/42.jpg",
			"vod_play_url":"第1集$https://media.example/1.m3u8#第2集$https://media.example/2.m3u8"}]}`))}, nil
	})
	adapter, err := New(client)
	if err != nil {
		t.Fatal(err)
	}
	page, err := adapter.Search(context.Background(), source.Definition{Key: "fixture", Name: "Fixture", Endpoint: "https://catalog.example/api", ImageHosts: []string{"images.example"}}, source.Query{Text: "测试", Page: 2})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 7 || page.TotalPages != 3 || page.NextPage != 3 || len(page.Items) != 1 {
		t.Fatalf("page = %#v", page)
	}
	item := page.Items[0]
	if item.Locator.ExternalID != "42" || item.Year != "2026" || item.Score != "8.6" || item.EpisodeCount != 2 || item.Episodes[1].Title != "第2集" || item.ImageURL != "https://images.example/42.jpg" {
		t.Fatalf("item = %#v", item)
	}
}

func TestAdapterDropsItemsWithoutPlayableEpisodes(t *testing.T) {
	client := roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"pagecount":1,"total":1,"list":[{"vod_id":1,"vod_name":"空条目","vod_play_url":""}]}`))}, nil
	})
	adapter, _ := New(client)
	page, err := adapter.Search(context.Background(), source.Definition{Key: "fixture", Name: "Fixture", Endpoint: "https://catalog.example/api"}, source.Query{Text: "空"})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 0 {
		t.Fatalf("items = %#v", page.Items)
	}
}

func TestMultipleLinesResolveIndependentlyAndKeepLegacyKeys(t *testing.T) {
	client := roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"list":[{"vod_id":"42","vod_name":"电影","vod_play_from":"主线$$$备用","vod_play_url":"正片$https://media.example/main.m3u8$$$国语$https://media.example/backup.mp4"}]}`))}, nil
	})
	adapter, _ := New(client)
	definition := source.Definition{Key: "test", Endpoint: "https://catalog.example/api"}
	locator := source.TitleLocator{SourceKey: "test", ExternalID: "42"}
	title, err := adapter.ResolveTitle(context.Background(), definition, locator)
	if err != nil || title.EpisodeCount != 1 {
		t.Fatalf("movie count: %#v, %v", title, err)
	}
	for _, want := range []struct{ key, url, line, name string }{
		{"0", "https://media.example/main.m3u8", "0", "主线"},
		{"line1-0", "https://media.example/backup.mp4", "1", "备用"},
	} {
		_, episode, err := adapter.ResolveEpisode(context.Background(), definition, source.EpisodeLocator{Title: locator, EpisodeKey: want.key})
		if err != nil || episode.PlaybackURL != want.url || episode.LineKey != want.line || episode.LineName != want.name {
			t.Fatalf("resolve %s: %#v, %v", want.key, episode, err)
		}
	}
}
