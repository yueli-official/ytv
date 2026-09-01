package declarative

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yueli-official/ytv/api/internal/source"
	"github.com/yueli-official/ytv/api/internal/source/rules"
)

func TestAniPlayerExtractorReturnsMediaURL(t *testing.T) {
	const mediaURL = "https://cdn.example/episode.mp4?signature=opaque"
	var server *httptest.Server
	server = httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/title":
			fmt.Fprint(response, `<main><a href="/play">第01集</a></main>`)
		case "/play":
			fmt.Fprint(response, `<script>var player_aaaa={"url":"Doki-token","encrypt":0}</script>`)
		case "/":
			fmt.Fprint(response, `<script>var config = {"url":"Doki-token","key":"key-value","time":"123","next":""}</script>`)
		case "/api_config.php":
			if err := request.ParseForm(); err != nil || request.Form.Get("url") != "Doki-token" || request.Form.Get("key") != "key-value" {
				t.Fatalf("form = %#v, %v", request.Form, err)
			}
			fmt.Fprintf(response, `{"code":"200","url":%q}`, mediaURL)
		default:
			http.NotFound(response, request)
		}
	}))
	defer server.Close()

	adapter, _ := New(server.Client())
	definition := source.Definition{Key: "ani", Endpoint: server.URL, Rule: &rules.Plan{
		BaseURL:  server.URL,
		Title:    rules.TitlePlan{RoadsXPath: "//main", EpisodesXPath: "/a"},
		Playback: rules.PlaybackPlan{Extractor: "aniplayer", ParserURL: server.URL + "/"},
	}}
	locator := source.TitleLocator{SourceKey: definition.Key, ExternalID: encodeURL(server.URL + "/title")}
	title, err := adapter.ResolveTitle(context.Background(), definition, locator)
	if err != nil {
		t.Fatal(err)
	}
	_, episode, err := adapter.ResolveEpisode(context.Background(), definition, source.EpisodeLocator{Title: locator, EpisodeKey: title.Episodes[0].Key})
	if err != nil || episode.PlaybackURL != mediaURL {
		t.Fatalf("episode = %#v, %v", episode, err)
	}
}
