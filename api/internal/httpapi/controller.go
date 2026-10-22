package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/yueli-official/foundation/go/auth"
	"github.com/yueli-official/foundation/go/traffic"
	v1 "github.com/yueli-official/ytv/api/api/v1"
	"github.com/yueli-official/ytv/api/internal/analytics"
	"github.com/yueli-official/ytv/api/internal/apierr"
	"github.com/yueli-official/ytv/api/internal/authz"
	"github.com/yueli-official/ytv/api/internal/discovery"
	"github.com/yueli-official/ytv/api/internal/playback"
	"github.com/yueli-official/ytv/api/internal/progress"
	"github.com/yueli-official/ytv/api/internal/searchanalytics"
	"github.com/yueli-official/ytv/api/internal/siteprofile"
	"github.com/yueli-official/ytv/api/internal/source"
	"github.com/yueli-official/ytv/api/internal/sourcecontrol"
	"github.com/yueli-official/ytv/api/internal/sourcepreference"
	"github.com/yueli-official/ytv/api/internal/trending"
)

type Core struct {
	siteKey     string
	authz       *authz.Service
	discovery   *discovery.Module
	catalog     source.Catalog
	playback    *playback.Resolver
	traffic     *analytics.Traffic
	searches    *searchanalytics.Service
	sources     *sourcecontrol.Service
	progress    *progress.Service
	preferences *sourcepreference.Service
	profile     *siteprofile.Service
	trending    *trending.Service
}

type Public struct{ core *Core }
type Account struct{ core *Core }
type Administrator struct{ core *Core }

type Dependencies struct {
	Authorization     *authz.Service
	Discovery         *discovery.Module
	Catalog           source.Catalog
	Playback          *playback.Resolver
	Traffic           *analytics.Traffic
	Searches          *searchanalytics.Service
	Sources           *sourcecontrol.Service
	Progress          *progress.Service
	SourcePreferences *sourcepreference.Service
	SiteProfile       *siteprofile.Service
	Trending          *trending.Service
}

func New(siteKey string, dependencies Dependencies) (*Core, error) {
	if strings.TrimSpace(siteKey) == "" || dependencies.Authorization == nil || dependencies.Discovery == nil || dependencies.Catalog == nil || dependencies.Playback == nil || dependencies.Traffic == nil || dependencies.Searches == nil || dependencies.Sources == nil || dependencies.Progress == nil || dependencies.SiteProfile == nil || dependencies.Trending == nil {
		return nil, errors.New("ytv/httpapi: site key and all product modules are required")
	}
	return &Core{
		siteKey: siteKey, authz: dependencies.Authorization, discovery: dependencies.Discovery,
		catalog: dependencies.Catalog, playback: dependencies.Playback, traffic: dependencies.Traffic,
		searches: dependencies.Searches, sources: dependencies.Sources, progress: dependencies.Progress, preferences: dependencies.SourcePreferences,
		profile: dependencies.SiteProfile, trending: dependencies.Trending,
	}, nil
}

func (controller *Public) Search(ctx context.Context, request *v1.SearchReq) (*v1.SearchRes, error) {
	startedAt := time.Now()
	page, err := controller.core.discovery.Search(ctx, discovery.Query{Text: request.Query, Page: request.Page, SourceKeys: splitSourceKeys(request.Sources)})
	if err != nil && !errors.Is(err, discovery.ErrAllSourcesFailed) {
		return nil, mapError(err)
	}
	response := &v1.SearchRes{
		Query: page.Query.Text, Page: page.Query.Page, Attempted: page.Attempted,
		Succeeded: page.Succeeded, Failed: page.Failed, Complete: page.Complete,
		Groups:  make([]v1.SearchGroupView, 0, len(page.Groups)),
		Sources: make([]v1.SourceResultView, 0, len(page.Sources)),
	}
	for _, result := range page.Sources {
		response.Sources = append(response.Sources, v1.SourceResultView{
			Key: result.Key, Name: result.Name, Outcome: string(result.Outcome),
			DurationMS: result.Duration.Milliseconds(), ItemCount: result.ItemCount,
		})
	}
	for _, group := range page.Groups {
		view := v1.SearchGroupView{Key: group.Key, Title: group.Title, Year: group.Year, Category: group.Category}
		for _, candidate := range group.Candidates {
			view.Candidates = append(view.Candidates, titleView(candidate.Title))
		}
		response.Groups = append(response.Groups, view)
	}
	if _, analyticsErr := controller.core.searches.Record(ctx, searchanalytics.Event{
		ID: request.EventID, Query: page.Query.Text, ResultCount: len(page.Groups),
		SuccessfulSourceCount: page.Succeeded, FailedSourceCount: page.Failed,
		Duration: time.Since(startedAt), OccurredAt: time.Now(),
	}); analyticsErr != nil {
		g.Log().Warningf(ctx, "record YTV search analytics: %v", analyticsErr)
	}
	if err != nil {
		return nil, apierr.Wrap(apierr.CodeSourceUnavailable, err)
	}
	return response, nil
}

func (controller *Public) Discover(ctx context.Context, request *v1.DiscoverReq) (*v1.SearchRes, error) {
	page, err := controller.core.discovery.Browse(ctx, request.Page)
	if err != nil && !errors.Is(err, discovery.ErrAllSourcesFailed) {
		return nil, mapError(err)
	}
	response := &v1.SearchRes{
		Page: page.Query.Page, Attempted: page.Attempted, Succeeded: page.Succeeded,
		Failed: page.Failed, Complete: page.Complete,
		Groups: make([]v1.SearchGroupView, 0, len(page.Groups)), Sources: make([]v1.SourceResultView, 0, len(page.Sources)),
	}
	for _, result := range page.Sources {
		response.Sources = append(response.Sources, v1.SourceResultView{
			Key: result.Key, Name: result.Name, Outcome: string(result.Outcome),
			DurationMS: result.Duration.Milliseconds(), ItemCount: result.ItemCount,
		})
	}
	for _, group := range page.Groups {
		view := v1.SearchGroupView{Key: group.Key, Title: group.Title, Year: group.Year, Category: group.Category}
		for _, candidate := range group.Candidates {
			view.Candidates = append(view.Candidates, titleView(candidate.Title))
		}
		response.Groups = append(response.Groups, view)
	}
	if err != nil {
		return nil, apierr.Wrap(apierr.CodeSourceUnavailable, err)
	}
	return response, nil
}

func (controller *Public) GetPublicSources(ctx context.Context, _ *v1.GetPublicSourcesReq) (*v1.GetPublicSourcesRes, error) {
	values := controller.core.sources.List()
	response := &v1.GetPublicSourcesRes{Items: make([]v1.PublicSourceView, 0, len(values))}
	for _, value := range values {
		if value.Enabled {
			response.Items = append(response.Items, v1.PublicSourceView{Key: value.Key, Name: value.Name})
		}
	}
	return response, nil
}

func (controller *Public) GetTrending(ctx context.Context, request *v1.GetTrendingReq) (*v1.GetTrendingRes, error) {
	page, err := controller.core.trending.List(ctx, trending.Query{
		Type: request.Type, Tag: request.Tag, Page: request.Page, Limit: request.Limit,
	})
	if err != nil {
		return nil, mapError(err)
	}
	response := &v1.GetTrendingRes{
		Type: page.Query.Type, Tag: page.Query.Tag, Page: page.Query.Page,
		Items: make([]v1.TrendingItemView, 0, len(page.Items)),
	}
	for _, item := range page.Items {
		response.Items = append(response.Items, v1.TrendingItemView{
			ID: item.ID, Title: item.Title, Score: item.Score, CoverURL: item.CoverURL, IsNew: item.IsNew,
		})
	}
	return response, nil
}

func (controller *Public) GetTitle(ctx context.Context, request *v1.GetTitleReq) (*v1.GetTitleRes, error) {
	title, err := controller.core.catalog.ResolveTitle(ctx, source.TitleLocator{SourceKey: request.SourceKey, ExternalID: request.ExternalID})
	if err != nil {
		return nil, mapError(err)
	}
	episodes := make([]v1.EpisodeView, 0, len(title.Episodes))
	for _, episode := range title.Episodes {
		lineKey, lineName := episode.LineKey, episode.LineName
		if lineKey == "" {
			lineKey = "0"
		}
		if lineName == "" {
			lineName = "默认线路"
		}
		episodes = append(episodes, v1.EpisodeView{Key: episode.Key, Title: episode.Title, LineKey: lineKey, LineName: lineName})
	}
	return &v1.GetTitleRes{Title: titleView(title), Episodes: episodes}, nil
}

func (controller *Public) ResolvePlayback(ctx context.Context, request *v1.ResolvePlaybackReq) (*v1.ResolvePlaybackRes, error) {
	manifest, err := controller.core.playback.Resolve(ctx, source.EpisodeLocator{
		Title: source.TitleLocator{SourceKey: request.SourceKey, ExternalID: request.ExternalID}, EpisodeKey: request.EpisodeKey,
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &v1.ResolvePlaybackRes{
		SourceName: manifest.SourceName, EpisodeKey: manifest.Locator.EpisodeKey,
		URL: manifest.URL, Engine: string(manifest.Engine), Timeline: manifest.Timeline, MediaType: manifest.MediaType,
	}, nil
}

func (controller *Public) RecordPageView(ctx context.Context, request *v1.RecordPageViewReq) (*v1.RecordPageViewRes, error) {
	result, err := controller.core.traffic.RecordPage(ctx, analytics.RecordInput{
		EventID: request.EventID, ResourceID: request.Path, VisitorID: request.VisitorID, OccurredAt: request.OccurredAt,
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &v1.RecordPageViewRes{Counted: result.Counted, Replay: result.Replay}, nil
}

func (controller *Public) RecordPlaybackStart(ctx context.Context, request *v1.RecordPlaybackStartReq) (*v1.RecordPlaybackStartRes, error) {
	result, err := controller.core.traffic.RecordPlayback(ctx, analytics.RecordInput{
		EventID:    request.EventID,
		ResourceID: strings.Join([]string{request.SourceKey, request.ExternalID, request.EpisodeKey}, ":"),
		VisitorID:  request.VisitorID, OccurredAt: request.OccurredAt,
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &v1.RecordPlaybackStartRes{Counted: result.Counted, Replay: result.Replay}, nil
}

func (core *Core) Public() *Public {
	return &Public{core: core}
}

func (core *Core) Administrator() *Administrator {
	return &Administrator{core: core}
}

func (core *Core) Account() *Account {
	return &Account{core: core}
}

func (controller *Public) GetMeta(ctx context.Context, _ *v1.GetMetaReq) (*v1.GetMetaRes, error) {
	profile, err := controller.core.profile.Get(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.GetMetaRes{
		Name:        profile.Name,
		Description: profile.Description,
		SiteKey:     controller.core.siteKey,
		Capabilities: []string{
			"discovery.search",
			"playback.resolve",
			"progress.local",
			"progress.sync",
		},
	}, nil
}

func (controller *Administrator) GetAdminSession(ctx context.Context, _ *v1.GetAdminSessionReq) (*v1.GetAdminSessionRes, error) {
	userKey, err := controller.core.requireAdministrator(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.GetAdminSessionRes{Allowed: true, UserKey: userKey}, nil
}

func (controller *Account) ListProgress(ctx context.Context, request *v1.ListProgressReq) (*v1.ListProgressRes, error) {
	userKey, err := requiredUser(ctx)
	if err != nil {
		return nil, err
	}
	items, err := controller.core.progress.List(ctx, userKey, request.Limit)
	if err != nil {
		return nil, mapError(err)
	}
	response := &v1.ListProgressRes{Items: make([]v1.PlaybackProgressView, 0, len(items))}
	for _, item := range items {
		response.Items = append(response.Items, playbackProgressView(item))
	}
	return response, nil
}

func (controller *Account) SaveProgress(ctx context.Context, request *v1.SaveProgressReq) (*v1.SaveProgressRes, error) {
	userKey, err := requiredUser(ctx)
	if err != nil {
		return nil, err
	}
	result, err := controller.core.progress.Save(ctx, progress.Record{
		UserKey: userKey, SourceKey: request.SourceKey, ExternalID: request.ExternalID,
		EpisodeKey: request.EpisodeKey, TitleName: request.TitleName, EpisodeTitle: request.EpisodeTitle,
		PositionSeconds: request.PositionSeconds, DurationSeconds: request.DurationSeconds,
		Completed: request.Completed, UpdatedAt: request.UpdatedAt,
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &v1.SaveProgressRes{Item: playbackProgressView(result.Record), Applied: result.Applied}, nil
}

func (controller *Account) GetSourcePreference(ctx context.Context, _ *v1.GetSourcePreferenceReq) (*v1.GetSourcePreferenceRes, error) {
	userKey, err := requiredUser(ctx)
	if err != nil {
		return nil, err
	}
	if controller.core.preferences == nil {
		return nil, errors.New("source preferences are unavailable")
	}
	value, err := controller.core.preferences.Get(ctx, userKey)
	if err != nil {
		return nil, mapError(err)
	}
	return &v1.GetSourcePreferenceRes{Preference: sourcePreferenceView(value)}, nil
}

func (controller *Account) SaveSourcePreference(ctx context.Context, request *v1.SaveSourcePreferenceReq) (*v1.SaveSourcePreferenceRes, error) {
	userKey, err := requiredUser(ctx)
	if err != nil {
		return nil, err
	}
	if controller.core.preferences == nil {
		return nil, errors.New("source preferences are unavailable")
	}
	value, err := controller.core.preferences.Save(ctx, userKey, request.DisabledSourceKeys, request.Revision)
	if err != nil {
		return nil, mapError(err)
	}
	return &v1.SaveSourcePreferenceRes{Preference: sourcePreferenceView(value)}, nil
}

func (controller *Administrator) GetTrafficOverview(ctx context.Context, request *v1.GetTrafficOverviewReq) (*v1.GetTrafficOverviewRes, error) {
	if _, err := controller.core.requireAdministrator(ctx); err != nil {
		return nil, err
	}
	overview, err := controller.core.traffic.Overview(ctx, request.Days)
	if err != nil {
		return nil, mapError(err)
	}
	return &v1.GetTrafficOverviewRes{
		Days:  overview.Days,
		Pages: trafficChannelView(overview.Pages), Playback: trafficChannelView(overview.Playback),
	}, nil
}

func (controller *Administrator) GetSearchOverview(ctx context.Context, request *v1.GetSearchOverviewReq) (*v1.GetSearchOverviewRes, error) {
	if _, err := controller.core.requireAdministrator(ctx); err != nil {
		return nil, err
	}
	overview, err := controller.core.searches.Overview(ctx, request.Days, request.Limit)
	if err != nil {
		return nil, mapError(err)
	}
	response := &v1.GetSearchOverviewRes{
		Days: overview.Days,
		Summary: v1.SearchSummaryView{
			Searches:           overview.Summary.Searches,
			ZeroResultSearches: overview.Summary.ZeroResultSearches,
			IncompleteSearches: overview.Summary.IncompleteSearches,
			AverageDurationMS:  overview.Summary.AverageDurationMS,
		},
		TopQueries: make([]v1.SearchQueryStatView, 0, len(overview.TopQueries)),
		Series:     make([]v1.SearchDayStatView, 0, len(overview.Series)),
	}
	for _, item := range overview.TopQueries {
		response.TopQueries = append(response.TopQueries, v1.SearchQueryStatView{
			Query: item.Query, Searches: item.Searches, ZeroResultSearches: item.ZeroResultSearches,
			AverageResultCount: item.AverageResultCount, AverageDurationMS: item.AverageDurationMS,
			LastSearchedAt: item.LastSearchedAt,
		})
	}
	for _, point := range overview.Series {
		response.Series = append(response.Series, v1.SearchDayStatView{
			Day: point.Day, Searches: point.Searches, ZeroResultSearches: point.ZeroResultSearches,
		})
	}
	return response, nil
}

func (controller *Administrator) GetSources(ctx context.Context, _ *v1.GetSourcesReq) (*v1.GetSourcesRes, error) {
	if _, err := controller.core.requireAdministrator(ctx); err != nil {
		return nil, err
	}
	values := controller.core.sources.List()
	response := &v1.GetSourcesRes{Items: make([]v1.SourceAdminView, 0, len(values))}
	for _, value := range values {
		response.Items = append(response.Items, sourceAdminView(value))
	}
	return response, nil
}

func (controller *Administrator) GetAvailableSources(ctx context.Context, _ *v1.GetAvailableSourcesReq) (*v1.GetAvailableSourcesRes, error) {
	if _, err := controller.core.requireAdministrator(ctx); err != nil {
		return nil, err
	}
	values := controller.core.sources.Available()
	response := &v1.GetAvailableSourcesRes{Items: make([]v1.AvailableSourceView, 0, len(values))}
	for _, value := range values {
		response.Items = append(response.Items, availableSourceView(value))
	}
	return response, nil
}

func (controller *Administrator) InstallSource(ctx context.Context, request *v1.InstallSourceReq) (*v1.InstallSourceRes, error) {
	if _, err := controller.core.requireAdministrator(ctx); err != nil {
		return nil, err
	}
	installed, err := controller.core.sources.Install(ctx, request.SourceKey, request.Revision)
	if err != nil {
		return nil, mapError(err)
	}
	if response := g.RequestFromCtx(ctx); response != nil {
		response.Response.Header().Set("Location", "/api/v1/admin/sources/"+url.PathEscape(installed.Key))
		response.Response.WriteHeader(http.StatusCreated)
	}
	return &v1.InstallSourceRes{Source: sourceAdminView(installed)}, nil
}

func (controller *Administrator) UpdateSource(ctx context.Context, request *v1.UpdateSourceReq) (*v1.UpdateSourceRes, error) {
	if _, err := controller.core.requireAdministrator(ctx); err != nil {
		return nil, err
	}
	updated, err := controller.core.sources.Update(ctx, request.SourceKey, request.Name, request.Enabled, request.HomeVisible, request.Priority, request.Revision)
	if err != nil {
		return nil, mapError(err)
	}
	return &v1.UpdateSourceRes{Source: sourceAdminView(updated)}, nil
}

func (controller *Administrator) DeleteSource(ctx context.Context, request *v1.DeleteSourceReq) (*v1.DeleteSourceRes, error) {
	if _, err := controller.core.requireAdministrator(ctx); err != nil {
		return nil, err
	}
	removed, err := controller.core.sources.Remove(ctx, request.SourceKey, request.Revision)
	if err != nil {
		return nil, mapError(err)
	}
	return &v1.DeleteSourceRes{Source: availableSourceView(removed)}, nil
}

func (controller *Administrator) ProbeSource(ctx context.Context, request *v1.ProbeSourceReq) (*v1.ProbeSourceRes, error) {
	if _, err := controller.core.requireAdministrator(ctx); err != nil {
		return nil, err
	}
	result, probeErr := controller.core.sources.Probe(ctx, request.SourceKey)
	if result.Source.Key == "" {
		return nil, mapError(probeErr)
	}
	return &v1.ProbeSourceRes{
		Source: sourceAdminView(result.Source),
		Health: v1.SourceHealthView{
			Available: result.Health.Available, CheckedAt: result.Health.CheckedAt,
			LatencyMS: result.Health.Latency.Milliseconds(), Detail: result.Health.Detail,
		},
	}, nil
}

func (controller *Administrator) ProbeAllSources(ctx context.Context, _ *v1.ProbeAllSourcesReq) (*v1.ProbeAllSourcesRes, error) {
	if _, err := controller.core.requireAdministrator(ctx); err != nil {
		return nil, err
	}
	values := controller.core.sources.ProbeAll(ctx)
	response := &v1.ProbeAllSourcesRes{Items: make([]v1.ProbeSourceRes, 0, len(values))}
	for _, value := range values {
		response.Items = append(response.Items, v1.ProbeSourceRes{
			Source: sourceAdminView(value.Source),
			Health: v1.SourceHealthView{
				Available: value.Health.Available, CheckedAt: value.Health.CheckedAt,
				LatencyMS: value.Health.Latency.Milliseconds(), Detail: value.Health.Detail,
			},
		})
	}
	return response, nil
}

func (controller *Administrator) GetSiteProfile(ctx context.Context, _ *v1.GetSiteProfileReq) (*v1.GetSiteProfileRes, error) {
	if _, err := controller.core.requireAdministrator(ctx); err != nil {
		return nil, err
	}
	profile, err := controller.core.profile.Get(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.GetSiteProfileRes{Profile: siteProfileView(profile)}, nil
}

func (controller *Administrator) UpdateSiteProfile(ctx context.Context, request *v1.UpdateSiteProfileReq) (*v1.UpdateSiteProfileRes, error) {
	if _, err := controller.core.requireAdministrator(ctx); err != nil {
		return nil, err
	}
	profile, err := controller.core.profile.Update(ctx, request.Name, request.Description, request.Revision)
	if err != nil {
		return nil, mapError(err)
	}
	return &v1.UpdateSiteProfileRes{Profile: siteProfileView(profile)}, nil
}

func (controller *Administrator) GetAuthorizationOverview(ctx context.Context, _ *v1.GetAuthorizationOverviewReq) (*v1.GetAuthorizationOverviewRes, error) {
	userKey, err := controller.core.requireAdministrator(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.GetAuthorizationOverviewRes{
		CurrentUserKey:    userKey,
		AdministratorRole: string(authz.RoleAdministrator),
		Capabilities: []string{
			string(authz.CapabilitySourceManage), string(authz.CapabilityAnalyticsRead),
			string(authz.CapabilitySettingsManage), string(authz.CapabilityProgressSync),
		},
		InitialClaimAvailable: false,
		BootstrapMode:         true,
	}, nil
}

func requiredUser(ctx context.Context) (string, error) {
	principal, ok := auth.FromContext(ctx)
	if !ok {
		return "", apierr.Error(apierr.Unauthorized)
	}
	if !principal.IsUser() || strings.TrimSpace(principal.Subject) == "" {
		return "", apierr.Error(apierr.Forbidden)
	}
	return principal.Subject, nil
}

func (core *Core) requireAdministrator(ctx context.Context) (string, error) {
	userKey, err := requiredUser(ctx)
	if err != nil {
		return "", err
	}
	if !core.authz.IsAdministrator(ctx) {
		return "", apierr.Error(apierr.Forbidden)
	}
	return userKey, nil
}

func titleView(title source.Title) v1.SearchCandidateView {
	return v1.SearchCandidateView{
		SourceKey: title.Locator.SourceKey, SourceName: title.SourceName, ExternalID: title.Locator.ExternalID,
		Name: title.Name, Year: title.Year, Category: title.Category, Area: title.Area, Language: title.Language, Score: title.Score,
		Remarks: title.Remarks, Summary: title.Summary, ImageURL: title.ImageURL, EpisodeCount: title.EpisodeCount,
	}
}

func trafficChannelView(channel analytics.ChannelOverview) v1.TrafficChannelView {
	points := make([]v1.TrafficPointView, 0, len(channel.Series))
	for _, point := range channel.Series {
		points = append(points, v1.TrafficPointView{
			Day: point.Day.String(), Views: point.Totals.Views, UniqueVisitorDays: point.Totals.UniqueVisitorDays,
		})
	}
	return v1.TrafficChannelView{
		AllTimeViews: channel.AllTime.Views, AllTimeUniqueVisitorDays: channel.AllTime.UniqueVisitorDays,
		RecentViews: channel.Recent.Views, RecentUniqueVisitorDays: channel.Recent.UniqueVisitorDays,
		Series: points,
	}
}

func sourceAdminView(value sourcecontrol.SourceView) v1.SourceAdminView {
	return v1.SourceAdminView{
		Key: value.Key, Name: value.Name, Kind: string(value.Kind), Enabled: value.Enabled,
		Catalog: value.Catalog, CatalogVersion: value.CatalogVersion,
		HomeVisible: value.HomeVisible, Priority: value.Priority, Revision: value.Revision, UpdatedAt: value.UpdatedAt,
		Compatibility: sourceCompatibilityView(value.Compatibility),
	}
}

func availableSourceView(value sourcecontrol.AvailableSource) v1.AvailableSourceView {
	return v1.AvailableSourceView{Key: value.Key, Name: value.Name, Kind: string(value.Kind), Catalog: value.Catalog, CatalogVersion: value.CatalogVersion, Revision: value.Revision, Compatibility: sourceCompatibilityView(value.Compatibility)}
}

func sourceCompatibilityView(values []source.CompatibilityNotice) []v1.SourceCompatibilityView {
	result := make([]v1.SourceCompatibilityView, 0, len(values))
	for _, value := range values {
		result = append(result, v1.SourceCompatibilityView{Level: value.Level, Code: value.Code, Message: value.Message})
	}
	return result
}

func splitSourceKeys(raw string) []string {
	values := strings.Split(raw, ",")
	seen := make(map[string]bool, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] || len(result) >= 32 {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	return result
}

func playbackProgressView(value progress.Record) v1.PlaybackProgressView {
	return v1.PlaybackProgressView{
		SourceKey: value.SourceKey, ExternalID: value.ExternalID, EpisodeKey: value.EpisodeKey,
		TitleName: value.TitleName, EpisodeTitle: value.EpisodeTitle,
		PositionSeconds: value.PositionSeconds, DurationSeconds: value.DurationSeconds,
		Completed: value.Completed, Revision: value.Revision, UpdatedAt: value.UpdatedAt,
	}
}

func sourcePreferenceView(value sourcepreference.Preference) v1.SourcePreferenceView {
	return v1.SourcePreferenceView{DisabledSourceKeys: value.DisabledSourceKeys, Revision: value.Revision, UpdatedAt: value.UpdatedAt, Exists: value.Exists}
}

func siteProfileView(value siteprofile.Profile) v1.SiteProfileView {
	return v1.SiteProfileView{
		SiteKey: value.SiteKey, Name: value.Name, Description: value.Description,
		Revision: value.Revision, UpdatedAt: value.UpdatedAt,
	}
}

func mapError(err error) error {
	var trafficError *traffic.Error
	if errors.As(err, &trafficError) && trafficError.Kind == traffic.ErrorInvalidInput {
		return apierr.Wrap(apierr.CodeInvalidInput, err)
	}
	switch {
	case errors.Is(err, analytics.ErrInvalid), errors.Is(err, progress.ErrInvalid), errors.Is(err, searchanalytics.ErrInvalid), errors.Is(err, siteprofile.ErrInvalid), errors.Is(err, sourcecontrol.ErrInvalid), errors.Is(err, sourcepreference.ErrInvalid), errors.Is(err, discovery.ErrInvalid), errors.Is(err, source.ErrInvalid):
		return apierr.Wrap(apierr.CodeInvalidInput, err)
	case errors.Is(err, trending.ErrInvalid):
		return apierr.Wrap(apierr.CodeInvalidInput, err)
	case errors.Is(err, source.ErrNotFound), errors.Is(err, playback.ErrEpisodeNotFound):
		return apierr.Wrap(apierr.CodeNotFound, err)
	case errors.Is(err, source.ErrUnavailable), errors.Is(err, discovery.ErrAllSourcesFailed):
		return apierr.Wrap(apierr.CodeSourceUnavailable, err)
	case errors.Is(err, playback.ErrUnsafeURL):
		return apierr.Wrap(apierr.CodePlaybackUnsafe, err)
	case errors.Is(err, sourcecontrol.ErrRevisionConflict):
		return apierr.Wrap(apierr.CodeRevisionConflict, err)
	case errors.Is(err, sourcepreference.ErrRevisionConflict):
		return apierr.Wrap(apierr.CodeRevisionConflict, err)
	case errors.Is(err, sourcecontrol.ErrAlreadyInstalled):
		return apierr.Wrap(apierr.CodeSourceAlreadyInstalled, err)
	case errors.Is(err, sourcecontrol.ErrNotInstalled):
		return apierr.Wrap(apierr.CodeSourceNotInstalled, err)
	case errors.Is(err, siteprofile.ErrRevisionConflict):
		return apierr.Wrap(apierr.CodeRevisionConflict, err)
	default:
		return err
	}
}
