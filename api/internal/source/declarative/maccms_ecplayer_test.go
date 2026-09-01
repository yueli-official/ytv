package declarative

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"

	"github.com/yueli-official/ytv/api/internal/source"
	"github.com/yueli-official/ytv/api/internal/source/rules"
)

func TestMacCMSECPlayerExtractorReturnsNativeMediaURL(t *testing.T) {
	const mediaURL = "https://media.example/video/episode-1.mp4?token=opaque"
	const uid = "nvcVEg"
	encryptedMedia := encryptECPlayerTestURL(t, mediaURL, uid)
	var parserHits atomic.Int32
	var server *httptest.Server
	server = httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/title":
			fmt.Fprint(response, `<main><a href="/vodplay/42-1-1.html">第01集</a></main>`)
		case "/vodplay/42-1-1.html":
			encoded := base64.StdEncoding.EncodeToString([]byte(url.QueryEscape("https://parser-target.example/play/42.html")))
			fmt.Fprintf(response, `<script>var player_aaaa={"encrypt":2,"url":%q,"from":"lmm"}</script>`, encoded)
		case "/player":
			if parserHits.Add(1) == 1 {
				http.Error(response, "transient", http.StatusBadGateway)
				return
			}
			fmt.Fprintf(response, `<script>let ConFig = {"code":200,"url":%q,"config":{"uid":%q}},box = player;</script>`, encryptedMedia, uid)
		default:
			http.NotFound(response, request)
		}
	}))
	defer server.Close()

	adapter, _ := New(server.Client())
	definition := source.Definition{Key: "special", Name: "Special", Kind: source.KindDeclarative, Endpoint: server.URL, Enabled: true, Rule: &rules.Plan{
		BaseURL:  server.URL,
		Title:    rules.TitlePlan{RoadsXPath: "//main", EpisodesXPath: "/a"},
		Playback: rules.PlaybackPlan{Extractor: "maccms_ecplayer", ParserURL: server.URL + "/player"},
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
	_, second, err := adapter.ResolveEpisode(context.Background(), definition, source.EpisodeLocator{Title: locator, EpisodeKey: title.Episodes[0].Key})
	if err != nil || second.PlaybackURL != mediaURL || parserHits.Load() != 2 {
		t.Fatalf("cached episode/parser hits = %#v/%d, %v", second, parserHits.Load(), err)
	}
}

func encryptECPlayerTestURL(t *testing.T, value, uid string) string {
	t.Helper()
	key := []byte("2890" + uid + "tB959C")
	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	padding := aes.BlockSize - len(value)%aes.BlockSize
	plaintext := append([]byte(value), bytes.Repeat([]byte{byte(padding)}, padding)...)
	encrypted := make([]byte, len(plaintext))
	cipher.NewCBCEncrypter(block, []byte("2F131BE91247866E")).CryptBlocks(encrypted, plaintext)
	return base64.StdEncoding.EncodeToString(encrypted)
}
