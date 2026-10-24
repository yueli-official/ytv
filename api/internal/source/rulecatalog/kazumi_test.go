package rulecatalog

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

const testRevision = "0123456789abcdef0123456789abcdef01234567"

func TestLoadKazumiCompilesCatalogAndSkipsIndividualFailures(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/" + testRevision + "/index.json":
			fmt.Fprint(response, `[
				{"name":"7sefun","version":"1.2","useNativePlayer":true,"antiCrawlerEnabled":false,"author":"A","lastUpdate":1},
				{"name":"bad/name","version":"1.0","useNativePlayer":true,"antiCrawlerEnabled":false,"author":"","lastUpdate":1},
				{"name":"Broken","version":"1.0","useNativePlayer":true,"antiCrawlerEnabled":false,"author":"","lastUpdate":1}
			]`)
		case "/" + testRevision + "/7sefun.json":
			fmt.Fprintf(response, `{
				"api":1,"name":"7sefun Anime","baseURL":%q,"searchURL":"/search/@keyword",
				"searchList":"//li","searchName":"/a/text()","searchResult":"/a/@href",
				"chapterRoads":"//main","chapterResult":"/a","useNativePlayer":true
			}`, server.URL)
		case "/" + testRevision + "/Broken.json":
			http.Error(response, "broken", http.StatusBadGateway)
		default:
			http.NotFound(response, request)
		}
	}))
	defer server.Close()

	result, err := LoadKazumi(context.Background(), KazumiOptions{
		Client: server.Client(), Revision: testRevision, RawBaseURL: server.URL, MaxParallel: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Definitions) != 1 || len(result.Skipped) != 2 {
		t.Fatalf("result = %#v", result)
	}
	definition := result.Definitions[0]
	if definition.Key != "kazumi-7sefun" || definition.Name != "7sefun Anime" || definition.Catalog != "Kazumi" || definition.CatalogVersion != "1.2" {
		t.Fatalf("definition = %#v", definition)
	}
	if !definition.InitiallyUninstalled || definition.Enabled || len(definition.Compatibility) == 0 || len(definition.PlaybackHosts) != 1 {
		t.Fatalf("definition defaults = %#v", definition)
	}
}

func TestLoadKazumiRequiresImmutableRevision(t *testing.T) {
	_, err := LoadKazumi(context.Background(), KazumiOptions{Client: http.DefaultClient, Revision: "main"})
	if err == nil {
		t.Fatal("mutable catalog revision unexpectedly accepted")
	}
}

func TestLoadKazumiSnapshotDoesNotRequireNetwork(t *testing.T) {
	result, err := LoadKazumiSnapshot("0972545266e7ac7cc06e7e522892239824f7d713")
	if err != nil {
		t.Fatalf("LoadKazumiSnapshot() error = %v", err)
	}
	if len(result.Definitions) != 11 || len(result.Skipped) != 9 {
		t.Fatalf("definitions/skipped = %d/%d, want 11/9", len(result.Definitions), len(result.Skipped))
	}
	for _, definition := range result.Definitions {
		if definition.Catalog != "Kazumi" || !definition.InitiallyUninstalled {
			t.Fatalf("unexpected embedded definition: %#v", definition)
		}
	}
}

func TestLoadKazumiSnapshotRejectsUnbundledRevision(t *testing.T) {
	_, err := LoadKazumiSnapshot("0000000000000000000000000000000000000000")
	if err == nil {
		t.Fatal("unbundled revision unexpectedly accepted")
	}
}

func TestLoadKazumiOfficialSnapshot(t *testing.T) {
	if os.Getenv("YTV_TEST_LIVE_KAZUMI") != "true" {
		t.Skip("set YTV_TEST_LIVE_KAZUMI=true for the non-CI catalog smoke")
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSHandshakeTimeout = 20 * time.Second
	client := &http.Client{Timeout: 25 * time.Second, Transport: transport}
	result, err := LoadKazumi(context.Background(), KazumiOptions{
		Client: client, Revision: "0972545266e7ac7cc06e7e522892239824f7d713", MaxParallel: 4,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Definitions) == 0 {
		t.Fatalf("official snapshot compiled no definitions; skipped = %#v", result.Skipped)
	}
	t.Logf("compiled %d Kazumi rules; skipped %d", len(result.Definitions), len(result.Skipped))
	if len(result.Skipped) > 0 {
		t.Logf("skipped rules: %#v", result.Skipped)
	}
}
