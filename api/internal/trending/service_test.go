package trending

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) Do(request *http.Request) (*http.Response, error) {
	return function(request)
}

func TestTrendingValidatesProviderDataAndCaches(t *testing.T) {
	calls := 0
	now := time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC)
	service, err := New(roundTripFunc(func(request *http.Request) (*http.Response, error) {
		calls++
		if request.URL.Query().Get("tag") != "热门" || request.URL.Query().Get("page_limit") != "12" {
			t.Fatalf("query = %s", request.URL.RawQuery)
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"subjects":[
			{"id":"1","title":"测试电影","rate":"8.1","cover":"https://img9.doubanio.com/view/photo/s_ratio_poster/public/p2933198755.jpg","is_new":true},
			{"id":"2","title":"坏海报","rate":"7.0","cover":"https://private.example/poster.jpg"}
		]}`))}, nil
	}), func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	page, err := service.List(context.Background(), Query{})
	if err != nil || len(page.Items) != 2 || page.Items[0].CoverURL != "/external-poster/douban/p2933198755.jpg" || page.Items[1].CoverURL != "" {
		t.Fatalf("page = %#v, %v", page, err)
	}
	if _, err := service.List(context.Background(), Query{}); err != nil || calls != 1 {
		t.Fatalf("cache calls/error = %d, %v", calls, err)
	}
}

func TestSafeDoubanCoverPathRejectsUntrustedShapes(t *testing.T) {
	for _, value := range []string{
		"http://img1.doubanio.com/view/photo/s_ratio_poster/public/p2933198755.jpg",
		"https://img1.doubanio.com/private/p2933198755.jpg",
		"https://img1.doubanio.com/view/photo/s_ratio_poster/public/not-a-poster.jpg",
		"https://img1.doubanio.com.evil.example/view/photo/s_ratio_poster/public/p2933198755.jpg",
		"https://127.0.0.1/view/photo/s_ratio_poster/public/p2933198755.jpg",
		"https://img1.doubanio.com/view/photo/s_ratio_poster/public/p2933198755.jpg?url=http://127.0.0.1",
	} {
		if got := safeDoubanCoverPath(value); got != "" {
			t.Fatalf("safeDoubanCoverPath(%q) = %q", value, got)
		}
	}
}
