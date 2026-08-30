package httpapi

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/yueli-official/foundation/go/authorization"
	"github.com/yueli-official/foundation/go/traffic"
	v1 "github.com/yueli-official/ytv/api/api/v1"
	"github.com/yueli-official/ytv/api/internal/analytics"
	"github.com/yueli-official/ytv/api/internal/authz"
	"github.com/yueli-official/ytv/api/internal/discovery"
	"github.com/yueli-official/ytv/api/internal/playback"
	"github.com/yueli-official/ytv/api/internal/progress"
	"github.com/yueli-official/ytv/api/internal/searchanalytics"
	"github.com/yueli-official/ytv/api/internal/siteprofile"
	"github.com/yueli-official/ytv/api/internal/source"
	"github.com/yueli-official/ytv/api/internal/source/fixture"
	"github.com/yueli-official/ytv/api/internal/sourcecontrol"
	"github.com/yueli-official/ytv/api/internal/trending"
)

func TestMetaKeepsProductCapabilitiesExplicit(t *testing.T) {
	controller := testController(t)
	response, err := controller.Public().GetMeta(context.Background(), &v1.GetMetaReq{})
	if err != nil {
		t.Fatal(err)
	}
	if response.Name != "YTV" || response.SiteKey != "ytv-test" {
		t.Fatalf("meta = %#v", response)
	}
	if len(response.Capabilities) != 4 {
		t.Fatalf("capabilities = %#v", response.Capabilities)
	}
}

func TestPublicFixtureJourneyHidesPlaybackUntilResolve(t *testing.T) {
	controller := testController(t)

	searchResponse, err := controller.Public().Search(context.Background(), &v1.SearchReq{Query: "环海"})
	if err != nil || len(searchResponse.Groups) != 1 || len(searchResponse.Groups[0].Candidates) != 1 {
		t.Fatalf("search/error = %#v / %v", searchResponse, err)
	}
	if searchResponse.Groups[0].Candidates[0].EpisodeCount != 2 {
		t.Fatalf("candidate = %#v", searchResponse.Groups[0].Candidates[0])
	}
	playbackResponse, err := controller.Public().ResolvePlayback(context.Background(), &v1.ResolvePlaybackReq{
		SourceKey: "fixture", ExternalID: "ring-sea", EpisodeKey: "0",
	})
	if err != nil || playbackResponse.Engine != "hls" || playbackResponse.URL == "" {
		t.Fatalf("playback/error = %#v / %v", playbackResponse, err)
	}
}

func TestPublicDiscoveryBrowsesDeploymentSources(t *testing.T) {
	controller := testController(t)
	response, err := controller.Public().Discover(context.Background(), &v1.DiscoverReq{})
	if err != nil || len(response.Groups) == 0 || response.Query != "" {
		t.Fatalf("discover/error = %#v / %v", response, err)
	}
}

func TestPublicSourcesExposeOnlyViewerSelectableCatalogs(t *testing.T) {
	controller := testController(t)
	response, err := controller.Public().GetPublicSources(context.Background(), &v1.GetPublicSourcesReq{})
	if err != nil || len(response.Sources) != 1 || response.Sources[0].Key != "fixture" {
		t.Fatalf("sources/error = %#v / %v", response, err)
	}
}

func testController(t *testing.T) *Core {
	t.Helper()
	definition, err := authorization.Compile(authz.Definition())
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := authorization.NewMemory(definition, authorization.MemoryOptions{
		RootScopeID: authz.RootScopeID, ProtectedSubjects: []authorization.SubjectRef{{Kind: authorization.SubjectUser, ID: "admin"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	adapter := fixture.New(fixture.Titles())
	registry, err := source.NewRegistry([]source.Definition{{
		Key: "fixture", Name: "Fixture", Kind: source.KindFixture, Endpoint: "https://fixture.invalid/catalog",
		PlaybackHosts: []string{"*.fixture.invalid"}, Enabled: true, HomeVisible: true,
	}}, map[source.Kind]source.Adapter{source.KindFixture: adapter})
	if err != nil {
		t.Fatal(err)
	}
	discoveryModule, err := discovery.New(registry, discovery.Options{})
	if err != nil {
		t.Fatal(err)
	}
	playbackResolver, err := playback.New(registry)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 8, 30, 12, 0, 0, 0, time.FixedZone("CST", 8*60*60))
	pageCatalog := traffic.MustCompile(analytics.Definition(analytics.PageResourceKind))
	playbackCatalog := traffic.MustCompile(analytics.Definition(analytics.PlaybackResourceKind))
	pageModule, err := traffic.NewMemory(pageCatalog, traffic.MemoryOptions{Secret: []byte("page-secret-32-bytes-for-testing-1"), Clock: func() time.Time { return now }})
	if err != nil {
		t.Fatal(err)
	}
	playbackModule, err := traffic.NewMemory(playbackCatalog, traffic.MemoryOptions{Secret: []byte("playback-secret-32-bytes-testing"), Clock: func() time.Time { return now }})
	if err != nil {
		t.Fatal(err)
	}
	trafficService, err := analytics.NewTraffic(pageModule, pageCatalog, playbackModule, playbackCatalog, func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	searchService, err := searchanalytics.New(searchanalytics.NewMemoryStore(), func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	sourceService, err := sourcecontrol.New(context.Background(), registry, sourcecontrol.NewMemoryStore(), func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	progressService, err := progress.New(progress.NewMemoryStore(), func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	profileService, err := siteprofile.New("ytv-test", siteprofile.NewMemoryStore(), func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	trendingService, err := trending.New(http.DefaultClient, func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	controller, err := New("ytv-test", Dependencies{
		Authorization: authz.New(runtime), Discovery: discoveryModule, Catalog: registry,
		Playback: playbackResolver, Traffic: trafficService, Searches: searchService,
		Sources: sourceService, Progress: progressService, SiteProfile: profileService, Trending: trendingService,
	})
	if err != nil {
		t.Fatal(err)
	}
	return controller
}
