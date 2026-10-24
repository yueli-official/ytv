package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	_ "github.com/lib/pq"
	"github.com/yueli-official/foundation/go/authorization"
	authorizationpostgres "github.com/yueli-official/foundation/go/authorization/postgres"
	"github.com/yueli-official/foundation/go/traffic"
	trafficpostgres "github.com/yueli-official/foundation/go/traffic/postgres"
	"github.com/yueli-official/ytv/api/internal/analytics"
	"github.com/yueli-official/ytv/api/internal/authz"
	"github.com/yueli-official/ytv/api/internal/discovery"
	"github.com/yueli-official/ytv/api/internal/httpapi"
	"github.com/yueli-official/ytv/api/internal/playback"
	"github.com/yueli-official/ytv/api/internal/progress"
	ytvruntime "github.com/yueli-official/ytv/api/internal/runtime"
	"github.com/yueli-official/ytv/api/internal/searchanalytics"
	"github.com/yueli-official/ytv/api/internal/server"
	"github.com/yueli-official/ytv/api/internal/siteprofile"
	"github.com/yueli-official/ytv/api/internal/source"
	"github.com/yueli-official/ytv/api/internal/source/browsercapture"
	"github.com/yueli-official/ytv/api/internal/source/declarative"
	"github.com/yueli-official/ytv/api/internal/source/fixture"
	"github.com/yueli-official/ytv/api/internal/source/htmlcatalog"
	"github.com/yueli-official/ytv/api/internal/source/jsoncatalog"
	"github.com/yueli-official/ytv/api/internal/source/rulecatalog"
	"github.com/yueli-official/ytv/api/internal/sourceconfig"
	"github.com/yueli-official/ytv/api/internal/sourcecontrol"
	"github.com/yueli-official/ytv/api/internal/sourcepreference"
	"github.com/yueli-official/ytv/api/internal/sourcereview"
	"github.com/yueli-official/ytv/api/internal/trending"
)

func main() {
	if output := os.Getenv("YTV_OPENAPI_OUTPUT"); output != "" {
		must(exportOpenAPI(output))
		return
	}
	ctx := context.Background()
	database, err := sql.Open("postgres", requiredEnvironment("YTV_DATABASE_URL"))
	must(err)
	defer database.Close()
	must(database.PingContext(ctx))

	authorizationDefinition, err := authorization.Compile(authz.Definition())
	must(err)
	bootstrapSubjects := commaSeparated(os.Getenv("YTV_BOOTSTRAP_ADMIN_SUBS"))
	protectedSubjects := make([]authorization.SubjectRef, 0, len(bootstrapSubjects))
	for _, subject := range bootstrapSubjects {
		protectedSubjects = append(protectedSubjects, authorization.SubjectRef{Kind: authorization.SubjectUser, ID: subject})
	}
	authorizationRuntime, err := authorizationpostgres.New(ctx, authorizationDefinition, authorizationpostgres.Options{
		DB:          database,
		InstanceKey: "ytv:" + environment("YTV_SITE_KEY", "ytv-main"),
		Memory: authorization.MemoryOptions{
			RootScopeID:       authz.RootScopeID,
			AllowUnclaimed:    len(protectedSubjects) == 0,
			ProtectedSubjects: protectedSubjects,
		},
	})
	must(err)

	verifier, err := ytvruntime.NewRemoteVerifier(ytvruntime.RemoteVerifierConfig{
		JWKSURL:           requiredEnvironment("YTV_JWKS_URL"),
		Issuer:            requiredEnvironment("YTV_JWKS_ISSUER"),
		Audience:          environment("YTV_JWKS_AUDIENCE", "ytv-api"),
		AllowLoopbackHTTP: environment("YTV_JWKS_ALLOW_LOOPBACK_HTTP", "false") == "true",
	})
	must(err)

	includeFixture := environment("YTV_DEV_FIXTURE", "false") == "true"
	definitions, err := sourceconfig.Load(os.Getenv("YTV_SOURCE_DEFINITIONS_JSON"), includeFixture)
	must(err)
	if !includeFixture && environment("YTV_KAZUMI_RULES_ENABLED", "true") == "true" {
		catalog, catalogErr := rulecatalog.LoadKazumiSnapshot(
			environment("YTV_KAZUMI_RULES_REF", "0972545266e7ac7cc06e7e522892239824f7d713"),
		)
		if catalogErr != nil {
			panic(catalogErr)
		} else {
			definitions = append(definitions, catalog.Definitions...)
			if len(catalog.Skipped) > 0 {
				g.Log().Infof(ctx, "Kazumi rule catalog left %d rules outside the deployment allowlist", len(catalog.Skipped))
			}
		}
	}
	jsonAdapter, err := jsoncatalog.New(&http.Client{Timeout: 5 * time.Second})
	must(err)
	htmlAdapter, err := htmlcatalog.New(&http.Client{Timeout: 5 * time.Second})
	must(err)
	var captureResolver *browsercapture.ChromeResolver
	if environment("YTV_BROWSER_CAPTURE_ENABLED", "true") == "true" {
		captureResolver, err = browsercapture.New(browsercapture.Options{
			Executable: os.Getenv("YTV_BROWSER_EXECUTABLE"), MaxConcurrent: 2, Timeout: 25 * time.Second,
			NoSandbox: environment("YTV_BROWSER_NO_SANDBOX", "false") == "true",
		})
		must(err)
		defer captureResolver.Close()
	}
	declarativeAdapter, err := declarative.NewWithBrowserCapture(&http.Client{Timeout: 5 * time.Second}, captureResolver)
	must(err)
	registry, err := source.NewRegistry(definitions, map[source.Kind]source.Adapter{
		source.KindJSONCatalog: jsonAdapter,
		source.KindHTMLCatalog: htmlAdapter,
		source.KindDeclarative: declarativeAdapter,
		source.KindFixture:     fixture.New(fixture.Titles()),
	})
	must(err)
	sourceStore, err := sourcecontrol.NewPostgresStore(database)
	must(err)
	sourceService, err := sourcecontrol.New(ctx, registry, sourceStore, nil)
	must(err)
	discoveryModule, err := discovery.New(registry, discovery.Options{Timeout: 6 * time.Second, MaxParallel: 4})
	must(err)
	playbackResolver, err := playback.New(registry)
	must(err)
	pageTrafficCatalog, err := traffic.Compile(analytics.Definition(analytics.PageResourceKind))
	must(err)
	pageTraffic, err := trafficpostgres.New(ctx, pageTrafficCatalog, trafficpostgres.Options{
		DB: database, InstanceKey: "ytv:" + environment("YTV_SITE_KEY", "ytv-main") + ":pages",
	})
	must(err)
	playbackTrafficCatalog, err := traffic.Compile(analytics.Definition(analytics.PlaybackResourceKind))
	must(err)
	playbackTraffic, err := trafficpostgres.New(ctx, playbackTrafficCatalog, trafficpostgres.Options{
		DB: database, InstanceKey: "ytv:" + environment("YTV_SITE_KEY", "ytv-main") + ":playback",
	})
	must(err)
	trafficService, err := analytics.NewTraffic(pageTraffic, pageTrafficCatalog, playbackTraffic, playbackTrafficCatalog, nil)
	must(err)
	searchStore, err := searchanalytics.NewPostgresStore(database)
	must(err)
	searchService, err := searchanalytics.New(searchStore, nil)
	must(err)
	progressStore, err := progress.NewPostgresStore(database)
	must(err)
	progressService, err := progress.New(progressStore, nil)
	must(err)
	preferenceStore, err := sourcepreference.NewPostgresStore(database)
	must(err)
	preferenceService, err := sourcepreference.New(preferenceStore, nil)
	must(err)
	profileStore, err := siteprofile.NewPostgresStore(database)
	must(err)
	profileService, err := siteprofile.New(environment("YTV_SITE_KEY", "ytv-main"), profileStore, nil)
	must(err)
	trendingService, err := trending.New(&http.Client{Timeout: 5 * time.Second}, nil)
	must(err)

	controller, err := httpapi.New(environment("YTV_SITE_KEY", "ytv-main"), httpapi.Dependencies{
		Authorization: authz.New(authorizationRuntime, len(protectedSubjects) > 0), Discovery: discoveryModule, Catalog: registry,
		Playback: playbackResolver, Traffic: trafficService, Searches: searchService,
		SourceReviews: sourcereview.New(sourcereview.NewPostgresStore(database)), Sources: sourceService, Progress: progressService, SourcePreferences: preferenceService, SiteProfile: profileService, Trending: trendingService,
	})
	must(err)
	httpServer := g.Server()
	httpServer.SetAddr(environment("YTV_API_ADDRESS", "127.0.0.1:8097"))
	server.Configure(httpServer, server.Dependencies{
		Controller: controller,
		Verifier:   verifier,
		Database:   database,
	})
	httpServer.Run()
}

func commaSeparated(raw string) []string {
	values := make([]string, 0)
	for _, value := range strings.Split(raw, ",") {
		if value = strings.TrimSpace(value); value != "" {
			values = append(values, value)
		}
	}
	return values
}

func environment(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}

func requiredEnvironment(name string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		panic(name + " is required")
	}
	return value
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
