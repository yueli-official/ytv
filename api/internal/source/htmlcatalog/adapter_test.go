package htmlcatalog

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yueli-official/ytv/api/internal/source"
)

func TestHTMLCatalogSearchDetailAndOnDemandPlayback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Content-Type", "text/html")
		switch request.URL.Path {
		case "/vod/search/page/1/wd/test.html", "/":
			_, _ = response.Write([]byte(`<div class="module-card-item"><a class="module-card-item-poster" href="/vod/detail/id/42.html"></a><div class="module-card-item-title"><a><strong>测试影片</strong></a></div><div class="module-card-item-class">科幻</div><div class="module-item-note">更新至 2 集</div></div>`))
		case "/vod/detail/id/42.html":
			_, _ = response.Write([]byte(`<h1>测试影片</h1><a class="module-play-list-link" href="/vod/play/id/42/sid/8/nid/1.html"><span>第一集</span></a><a class="module-play-list-link" href="/vod/play/id/42/sid/8/nid/2.html"><span>第二集</span></a>`))
		case "/vod/play/id/42/sid/8/nid/2.html":
			_, _ = response.Write([]byte(`<h1>测试影片</h1><a class="module-play-list-link" href="/vod/play/id/42/sid/8/nid/2.html"><span>第二集</span></a><script>var player_aaaa={"url":"https:\/\/media.example\/two.m3u8"}</script>`))
		default:
			http.NotFound(response, request)
		}
	}))
	defer server.Close()
	adapter, err := New(server.Client())
	if err != nil {
		t.Fatal(err)
	}
	definition := source.Definition{Key: "html", Name: "HTML", Kind: source.KindHTMLCatalog, Endpoint: server.URL, Enabled: true, AllowHTTP: true}
	page, err := adapter.Search(context.Background(), definition, source.Query{Text: "test"})
	if err != nil || len(page.Items) != 1 || page.Items[0].Locator.ExternalID != "42" {
		t.Fatalf("search = %#v, %v", page, err)
	}
	title, err := adapter.ResolveTitle(context.Background(), definition, source.TitleLocator{SourceKey: "html", ExternalID: "42"})
	if err != nil || title.EpisodeCount != 2 {
		t.Fatalf("title = %#v, %v", title, err)
	}
	_, episode, err := adapter.ResolveEpisode(context.Background(), definition, source.EpisodeLocator{
		Title: source.TitleLocator{SourceKey: "html", ExternalID: "42"}, EpisodeKey: "1",
	})
	if err != nil || episode.PlaybackURL != "https://media.example/two.m3u8" {
		t.Fatalf("episode = %#v, %v", episode, err)
	}
}
