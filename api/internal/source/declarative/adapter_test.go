package declarative

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yueli-official/ytv/api/internal/source"
	"github.com/yueli-official/ytv/api/internal/source/browsercapture"
	"github.com/yueli-official/ytv/api/internal/source/rules"
)

type captureStub struct {
	result browsercapture.Result
	err    error
}

func (stub captureStub) Resolve(context.Context, browsercapture.Request) (browsercapture.Result, error) {
	return stub.result, stub.err
}

func TestAdapterExecutesSearchTitleAndMediaExtractionPlan(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/search":
			if request.URL.Query().Get("q") == "" {
				t.Fatalf("query = %q", request.URL.RawQuery)
			}
			fmt.Fprint(response, `<ul><li class="result"><a href="/title/sea"><strong>海之环</strong></a></li></ul>`)
		case "/title/sea":
			fmt.Fprint(response, `<section class="road"><a class="episode" href="/watch/sea-1">第一集</a><a class="episode" href="/watch/sea-2">第二集</a></section>`)
		case "/watch/sea-1":
			fmt.Fprint(response, `<video><source src="https://media.example/sea-1.m3u8"></video>`)
		case "/watch/sea-2":
			fmt.Fprint(response, `<video><source src="https://media.example/sea-2.m3u8"></video>`)
		default:
			http.NotFound(response, request)
		}
	}))
	defer server.Close()

	adapter, err := New(server.Client())
	if err != nil {
		t.Fatal(err)
	}
	definition := source.Definition{
		Key: "anime", Name: "Anime", Kind: source.KindDeclarative, Endpoint: server.URL, Enabled: true,
		Rule: &rules.Plan{
			BaseURL: server.URL,
			Search: rules.SearchPlan{
				Request:   rules.RequestPlan{Method: http.MethodGet, URL: "/search?q=@keyword"},
				ListXPath: "//li[@class='result']", NameXPath: "/a/strong/text()", LinkXPath: "/a/@href",
			},
			Title:    rules.TitlePlan{RoadsXPath: "//section[@class='road']", EpisodesXPath: "/a[@class='episode']"},
			Playback: rules.PlaybackPlan{MediaXPath: "//video/source/@src"},
		},
	}

	page, err := adapter.Search(context.Background(), definition, source.Query{Text: "海"})
	if err != nil || len(page.Items) != 1 || page.Items[0].Name != "海之环" {
		t.Fatalf("search = %#v, %v", page, err)
	}
	title, err := adapter.ResolveTitle(context.Background(), definition, page.Items[0].Locator)
	if err != nil || len(title.Episodes) != 2 || title.Episodes[0].Title != "第一集" {
		t.Fatalf("title = %#v, %v", title, err)
	}
	_, episode, err := adapter.ResolveEpisode(context.Background(), definition, source.EpisodeLocator{Title: title.Locator, EpisodeKey: title.Episodes[0].Key})
	if err != nil || episode.PlaybackURL != "https://media.example/sea-1.m3u8" {
		t.Fatalf("episode = %#v, %v", episode, err)
	}
	health, err := adapter.Probe(context.Background(), definition)
	if err != nil || !health.Available || health.Detail != "搜索链路可用，播放按具体条目解析" {
		t.Fatalf("probe = %#v, %v", health, err)
	}
}

func TestAdapterAllowsDirectMediaEpisodeWithoutExtractor(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		fmt.Fprint(response, `<main><a href="https://media.example/direct.mp4">正片</a></main>`)
	}))
	defer server.Close()
	adapter, _ := New(server.Client())
	definition := source.Definition{Key: "direct", Name: "Direct", Kind: source.KindDeclarative, Endpoint: server.URL, Enabled: true, Rule: &rules.Plan{
		BaseURL: server.URL, Title: rules.TitlePlan{RoadsXPath: "//main", EpisodesXPath: "/a"},
	}}
	locator := source.TitleLocator{SourceKey: definition.Key, ExternalID: encodeURL(server.URL + "/title")}
	title, err := adapter.ResolveTitle(context.Background(), definition, locator)
	if err != nil {
		t.Fatal(err)
	}
	_, episode, err := adapter.ResolveEpisode(context.Background(), definition, source.EpisodeLocator{Title: locator, EpisodeKey: title.Episodes[0].Key})
	if err != nil || episode.PlaybackURL != "https://media.example/direct.mp4" {
		t.Fatalf("episode = %#v, %v", episode, err)
	}
}

func TestAdapterFallsBackToBrowserCaptureForWebviewRule(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		fmt.Fprint(response, `<main><a href="/play/1">第01集</a></main>`)
	}))
	defer server.Close()
	adapter, _ := NewWithBrowserCapture(server.Client(), captureStub{result: browsercapture.Result{URL: "https://media.example/episode.m3u8"}})
	definition := source.Definition{Key: "captured", Endpoint: server.URL, Rule: &rules.Plan{
		BaseURL:  server.URL,
		Title:    rules.TitlePlan{RoadsXPath: "//main", EpisodesXPath: "/a"},
		Playback: rules.PlaybackPlan{UseWebview: true},
	}}
	locator := source.TitleLocator{SourceKey: definition.Key, ExternalID: encodeURL(server.URL + "/title")}
	title, err := adapter.ResolveTitle(context.Background(), definition, locator)
	if err != nil {
		t.Fatal(err)
	}
	_, episode, err := adapter.ResolveEpisode(context.Background(), definition, source.EpisodeLocator{Title: locator, EpisodeKey: title.Episodes[0].Key})
	if err != nil || episode.PlaybackURL != "https://media.example/episode.m3u8" {
		t.Fatalf("episode = %#v, %v", episode, err)
	}
}

func TestAdapterPreservesSeparatePlaybackLines(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<section class="road"><a href="https://media.example/main.mp4">正片</a></section><section class="road"><a href="https://media.example/backup.mp4">正片</a></section>`)
	}))
	defer server.Close()
	adapter, _ := New(server.Client())
	definition := source.Definition{Key: "film", Name: "Films", Endpoint: server.URL, Rule: &rules.Plan{
		BaseURL: server.URL, Title: rules.TitlePlan{RoadsXPath: "//section[@class='road']", EpisodesXPath: "/a", MultipleRoads: true},
	}}
	locator := source.TitleLocator{SourceKey: "film", ExternalID: encodeURL(server.URL + "/title")}
	title, err := adapter.ResolveTitle(context.Background(), definition, locator)
	if err != nil || len(title.Episodes) != 2 || title.EpisodeCount != 1 {
		t.Fatalf("title = %#v, %v", title, err)
	}
	if title.Episodes[0].LineKey == title.Episodes[1].LineKey {
		t.Fatal("alternate lines were merged")
	}
	for index, want := range []string{"https://media.example/main.mp4", "https://media.example/backup.mp4"} {
		if title.Episodes[index].Key != encodeURL(want) {
			t.Fatal("legacy URL keys changed")
		}
		_, episode, err := adapter.ResolveEpisode(context.Background(), definition, source.EpisodeLocator{Title: locator, EpisodeKey: title.Episodes[index].Key})
		if err != nil || episode.PlaybackURL != want {
			t.Fatalf("playback = %#v, %v", episode, err)
		}
	}
}

func TestTitleExtractsNameAndNumbersOnlyPlayableRoads(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<div class="video-p-name">番剧名称<span>热度</span></div><ul></ul><ul><li><a href="/one">第01集</a></li></ul><ul><li><a href="https://other.example/page">无效</a></li></ul><ul><li><a href="/two">第01集</a></li></ul>`)
	}))
	defer server.Close()
	a, _ := New(server.Client())
	d := source.Definition{Key: "anime", Name: "Source", Endpoint: server.URL, Rule: &rules.Plan{BaseURL: server.URL, Title: rules.TitlePlan{NameXPath: "//div[@class='video-p-name']/text()[1]", RoadsXPath: "//ul", EpisodesXPath: "//li/a", MultipleRoads: true}}}
	title, err := a.ResolveTitle(context.Background(), d, source.TitleLocator{SourceKey: "anime", ExternalID: encodeURL(server.URL + "/title")})
	if err != nil {
		t.Fatal(err)
	}
	if title.Name != "番剧名称" || len(title.Episodes) != 2 || title.Episodes[0].LineName != "线路 1" || title.Episodes[1].LineName != "线路 2" {
		t.Fatalf("title=%#v", title)
	}
	if title.Episodes[0].Key != encodeURL(server.URL+"/one") || title.Episodes[1].Key != encodeURL(server.URL+"/two") {
		t.Fatal("episode identity changed")
	}
}
