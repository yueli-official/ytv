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
		if request.URL.Query().Get("wd") != "测试" || request.URL.Query().Get("pg") != "2" {
			t.Fatalf("query = %s", request.URL.RawQuery)
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{
			"pagecount":3,"total":7,"list":[{"vod_id":"42","vod_name":"演示影片","vod_year":2026,
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
	if page.Total != 7 || page.NextPage != 3 || len(page.Items) != 1 {
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
