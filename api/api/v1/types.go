package v1

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

type GetMetaReq struct {
	g.Meta `path:"/api/v1/meta" method:"get" tags:"Public" summary:"Read YTV service metadata"`
}

type GetMetaRes struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	SiteKey      string   `json:"siteKey"`
	Capabilities []string `json:"capabilities"`
}

type GetAdminSessionReq struct {
	g.Meta `path:"/api/v1/admin/session" method:"get" tags:"Administration" summary:"Check administrator access"`
}

type GetAdminSessionRes struct {
	Allowed bool   `json:"allowed"`
	UserKey string `json:"userKey"`
}

type SearchReq struct {
	g.Meta  `path:"/api/v1/search" method:"get" tags:"Discovery" summary:"Search enabled source catalogs"`
	Query   string `json:"q" in:"query" v:"required|length:1,100"`
	Page    int    `json:"page" in:"query"`
	Sources string `json:"sources" in:"query" v:"length:0,1000"`
	EventID string `json:"eventId" in:"query" v:"length:0,200"`
}

type DiscoverReq struct {
	g.Meta `path:"/api/v1/discover" method:"get" tags:"Discovery" summary:"Browse current items from enabled source catalogs"`
	Page   int `json:"page" in:"query"`
}

type GetTrendingReq struct {
	g.Meta `path:"/api/v1/trending" method:"get" tags:"Discovery" summary:"Read the current external trending list"`
	Type   string `json:"type" in:"query"`
	Tag    string `json:"tag" in:"query"`
	Page   int    `json:"page" in:"query"`
	Limit  int    `json:"limit" in:"query"`
}

type TrendingItemView struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Score    string `json:"score"`
	CoverURL string `json:"coverUrl"`
	IsNew    bool   `json:"isNew"`
}

type GetTrendingRes struct {
	Type  string             `json:"type"`
	Tag   string             `json:"tag"`
	Page  int                `json:"page"`
	Items []TrendingItemView `json:"items"`
}

type SourceResultView struct {
	Key              string `json:"key"`
	Name             string `json:"name"`
	Outcome          string `json:"outcome"`
	VerificationKind string `json:"verificationKind,omitempty"`
	DurationMS       int64  `json:"durationMs"`
	ItemCount        int    `json:"itemCount"`
}

type GetPublicSourcesReq struct {
	g.Meta `path:"/api/v1/sources" method:"get" tags:"Discovery" summary:"List enabled source catalogs available to viewers"`
}

type PublicSourceView struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type GetPublicSourcesRes struct {
	Items []PublicSourceView `json:"items"`
}

type SearchCandidateView struct {
	SourceKey    string `json:"sourceKey"`
	SourceName   string `json:"sourceName"`
	ExternalID   string `json:"externalId"`
	Name         string `json:"name"`
	Year         string `json:"year"`
	Category     string `json:"category"`
	Area         string `json:"area"`
	Language     string `json:"language"`
	Score        string `json:"score"`
	Remarks      string `json:"remarks"`
	Summary      string `json:"summary"`
	ImageURL     string `json:"imageUrl"`
	EpisodeCount int    `json:"episodeCount"`
}

type SearchGroupView struct {
	Key        string                `json:"key"`
	Title      string                `json:"title"`
	Year       string                `json:"year"`
	Category   string                `json:"category"`
	Candidates []SearchCandidateView `json:"candidates"`
}

type SearchRes struct {
	Query     string             `json:"query"`
	Page      int                `json:"page"`
	Groups    []SearchGroupView  `json:"groups"`
	Sources   []SourceResultView `json:"sources"`
	Attempted int                `json:"attempted"`
	Succeeded int                `json:"succeeded"`
	Failed    int                `json:"failed"`
	Complete  bool               `json:"complete"`
}

type GetTitleReq struct {
	g.Meta     `path:"/api/v1/titles/{sourceKey}/{externalId}" method:"get" tags:"Discovery" summary:"Resolve one source title"`
	SourceKey  string `json:"sourceKey" in:"path" v:"required"`
	ExternalID string `json:"externalId" in:"path" v:"required"`
}

type EpisodeView struct {
	Key      string `json:"key"`
	Title    string `json:"title"`
	LineKey  string `json:"lineKey"`
	LineName string `json:"lineName"`
}

type GetTitleRes struct {
	Title    SearchCandidateView `json:"title"`
	Episodes []EpisodeView       `json:"episodes"`
}

type ResolvePlaybackReq struct {
	g.Meta     `path:"/api/v1/playback/{sourceKey}/{externalId}/{episodeKey}" method:"get" tags:"Playback" summary:"Resolve a safe playback manifest"`
	SourceKey  string `json:"sourceKey" in:"path" v:"required"`
	ExternalID string `json:"externalId" in:"path" v:"required"`
	EpisodeKey string `json:"episodeKey" in:"path" v:"required"`
}

type ResolvePlaybackRes struct {
	SourceName string `json:"sourceName"`
	EpisodeKey string `json:"episodeKey"`
	URL        string `json:"url"`
	Engine     string `json:"engine"`
	Timeline   string `json:"timeline"`
	MediaType  string `json:"mediaType"`
}

type RecordPageViewReq struct {
	g.Meta     `path:"/api/v1/traffic/views" method:"post" tags:"Traffic" summary:"Record one public page view"`
	EventID    string    `json:"eventId" v:"required|length:16,200"`
	Path       string    `json:"path" v:"required|max-length:200"`
	VisitorID  string    `json:"visitorId" v:"length:0,128"`
	OccurredAt time.Time `json:"occurredAt"`
}

type RecordPageViewRes struct {
	Counted bool `json:"counted"`
	Replay  bool `json:"replay"`
}

type RecordPlaybackStartReq struct {
	g.Meta     `path:"/api/v1/traffic/playback" method:"post" tags:"Traffic" summary:"Record one actual playback start"`
	EventID    string    `json:"eventId" v:"required|length:16,200"`
	SourceKey  string    `json:"sourceKey" v:"required"`
	ExternalID string    `json:"externalId" v:"required"`
	EpisodeKey string    `json:"episodeKey" v:"required"`
	VisitorID  string    `json:"visitorId" v:"length:0,128"`
	OccurredAt time.Time `json:"occurredAt"`
}

type RecordPlaybackStartRes struct {
	Counted bool `json:"counted"`
	Replay  bool `json:"replay"`
}

type TrafficPointView struct {
	Day               string `json:"day"`
	Views             int64  `json:"views"`
	UniqueVisitorDays int64  `json:"uniqueVisitorDays"`
}

type TrafficChannelView struct {
	AllTimeViews             int64              `json:"allTimeViews"`
	AllTimeUniqueVisitorDays int64              `json:"allTimeUniqueVisitorDays"`
	RecentViews              int64              `json:"recentViews"`
	RecentUniqueVisitorDays  int64              `json:"recentUniqueVisitorDays"`
	Series                   []TrafficPointView `json:"series"`
}

type GetTrafficOverviewReq struct {
	g.Meta `path:"/api/v1/admin/traffic" method:"get" tags:"Administration" summary:"Read page and playback traffic"`
	Days   int `json:"days" in:"query"`
}

type GetTrafficOverviewRes struct {
	Days     int                `json:"days"`
	Pages    TrafficChannelView `json:"pages"`
	Playback TrafficChannelView `json:"playback"`
}

type GetSearchOverviewReq struct {
	g.Meta `path:"/api/v1/admin/searches" method:"get" tags:"Administration" summary:"Read search term analytics"`
	Days   int `json:"days" in:"query"`
	Limit  int `json:"limit" in:"query"`
}

type SearchSummaryView struct {
	Searches           int64 `json:"searches"`
	ZeroResultSearches int64 `json:"zeroResultSearches"`
	IncompleteSearches int64 `json:"incompleteSearches"`
	AverageDurationMS  int64 `json:"averageDurationMs"`
}

type SearchQueryStatView struct {
	Query              string    `json:"query"`
	Searches           int64     `json:"searches"`
	ZeroResultSearches int64     `json:"zeroResultSearches"`
	AverageResultCount float64   `json:"averageResultCount"`
	AverageDurationMS  int64     `json:"averageDurationMs"`
	LastSearchedAt     time.Time `json:"lastSearchedAt"`
}

type SearchDayStatView struct {
	Day                string `json:"day"`
	Searches           int64  `json:"searches"`
	ZeroResultSearches int64  `json:"zeroResultSearches"`
}

type GetSearchOverviewRes struct {
	Days       int                   `json:"days"`
	Summary    SearchSummaryView     `json:"summary"`
	TopQueries []SearchQueryStatView `json:"topQueries"`
	Series     []SearchDayStatView   `json:"series"`
}

type SourceAdminView struct {
	Key            string                    `json:"key"`
	Name           string                    `json:"name"`
	Kind           string                    `json:"kind"`
	Catalog        string                    `json:"catalog"`
	CatalogVersion string                    `json:"catalogVersion"`
	Enabled        bool                      `json:"enabled"`
	HomeVisible    bool                      `json:"homeVisible"`
	Priority       int                       `json:"priority"`
	Revision       int64                     `json:"revision"`
	UpdatedAt      time.Time                 `json:"updatedAt"`
	Compatibility  []SourceCompatibilityView `json:"compatibility"`
}

type SourceCompatibilityView struct {
	Level   string `json:"level"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type GetSourcesReq struct {
	g.Meta `path:"/api/v1/admin/sources" method:"get" tags:"Administration" summary:"List deployment-owned sources and runtime settings"`
}

type GetSourcesRes struct {
	Items []SourceAdminView `json:"items"`
}

type AvailableSourceView struct {
	Key            string                    `json:"key"`
	Name           string                    `json:"name"`
	Kind           string                    `json:"kind"`
	Catalog        string                    `json:"catalog"`
	CatalogVersion string                    `json:"catalogVersion"`
	Revision       int64                     `json:"revision"`
	Compatibility  []SourceCompatibilityView `json:"compatibility"`
}

type GetAvailableSourcesReq struct {
	g.Meta `path:"/api/v1/admin/source-catalog" method:"get" tags:"Administration" summary:"List approved source definitions available to install"`
}

type GetAvailableSourcesRes struct {
	Items []AvailableSourceView `json:"items"`
}

type InstallSourceReq struct {
	g.Meta    `path:"/api/v1/admin/sources" method:"post" tags:"Administration" summary:"Install one approved source definition"`
	SourceKey string `json:"sourceKey" v:"required"`
	Revision  int64  `json:"revision" v:"min:0"`
}

type InstallSourceRes struct {
	g.Meta `status:"201"`
	Source SourceAdminView `json:"source"`
}

type UpdateSourceReq struct {
	g.Meta      `path:"/api/v1/admin/sources/{sourceKey}" method:"patch" tags:"Administration" summary:"Update source runtime settings"`
	SourceKey   string `json:"sourceKey" in:"path" v:"required"`
	Name        string `json:"name" v:"required|length:1,80"`
	Enabled     bool   `json:"enabled"`
	HomeVisible bool   `json:"homeVisible"`
	Priority    int    `json:"priority" v:"min:0|max:100000"`
	Revision    int64  `json:"revision" v:"min:0"`
}

type UpdateSourceRes struct {
	Source SourceAdminView `json:"source"`
}

type DeleteSourceReq struct {
	g.Meta    `path:"/api/v1/admin/sources/{sourceKey}" method:"delete" tags:"Administration" summary:"Remove one installed source while keeping its approved definition"`
	SourceKey string `json:"sourceKey" in:"path" v:"required"`
	Revision  int64  `json:"revision" v:"min:0"`
}

type DeleteSourceRes struct {
	Source AvailableSourceView `json:"source"`
}

type ProbeSourceReq struct {
	g.Meta    `path:"/api/v1/admin/sources/{sourceKey}/probe" method:"post" tags:"Administration" summary:"Probe one deployment-owned source"`
	SourceKey string `json:"sourceKey" in:"path" v:"required"`
}

type SourceHealthView struct {
	Available bool      `json:"available"`
	CheckedAt time.Time `json:"checkedAt"`
	LatencyMS int64     `json:"latencyMs"`
	Detail    string    `json:"detail"`
}

type ProbeSourceRes struct {
	Source SourceAdminView  `json:"source"`
	Health SourceHealthView `json:"health"`
}

type ProbeAllSourcesReq struct {
	g.Meta `path:"/api/v1/admin/sources/probe-all" method:"post" tags:"Administration" summary:"Probe every installed source"`
}

type ProbeAllSourcesRes struct {
	Items []ProbeSourceRes `json:"items"`
}

type PlaybackProgressView struct {
	SourceKey       string    `json:"sourceKey"`
	ExternalID      string    `json:"externalId"`
	EpisodeKey      string    `json:"episodeKey"`
	TitleName       string    `json:"titleName"`
	EpisodeTitle    string    `json:"episodeTitle"`
	PositionSeconds int       `json:"positionSeconds"`
	DurationSeconds int       `json:"durationSeconds"`
	Completed       bool      `json:"completed"`
	Revision        int64     `json:"revision"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type ListProgressReq struct {
	g.Meta `path:"/api/v1/progress" method:"get" tags:"Account" summary:"List the signed-in user's playback progress"`
	Limit  int `json:"limit" in:"query"`
}

type ListProgressRes struct {
	Items []PlaybackProgressView `json:"items"`
}

type SaveProgressReq struct {
	g.Meta          `path:"/api/v1/progress/{sourceKey}/{externalId}/{episodeKey}" method:"put" tags:"Account" summary:"Merge playback progress for the signed-in user"`
	SourceKey       string    `json:"sourceKey" in:"path" v:"required"`
	ExternalID      string    `json:"externalId" in:"path" v:"required"`
	EpisodeKey      string    `json:"episodeKey" in:"path" v:"required"`
	TitleName       string    `json:"titleName" v:"max-length:200"`
	EpisodeTitle    string    `json:"episodeTitle" v:"max-length:200"`
	PositionSeconds int       `json:"positionSeconds" v:"min:0"`
	DurationSeconds int       `json:"durationSeconds" v:"required|min:1"`
	Completed       bool      `json:"completed"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type SaveProgressRes struct {
	Item    PlaybackProgressView `json:"item"`
	Applied bool                 `json:"applied"`
}

type SourcePreferenceView struct {
	DisabledSourceKeys []string  `json:"disabledSourceKeys"`
	Revision           int64     `json:"revision"`
	UpdatedAt          time.Time `json:"updatedAt"`
	Exists             bool      `json:"exists"`
}

type GetSourcePreferenceReq struct {
	g.Meta `path:"/api/v1/preferences/sources" method:"get" tags:"Account" summary:"Read the signed-in user's source preference"`
}
type GetSourcePreferenceRes struct {
	Preference SourcePreferenceView `json:"preference"`
}

type SaveSourcePreferenceReq struct {
	g.Meta             `path:"/api/v1/preferences/sources" method:"put" tags:"Account" summary:"Save the signed-in user's source preference"`
	DisabledSourceKeys []string `json:"disabledSourceKeys"`
	Revision           int64    `json:"revision" v:"min:0"`
}
type SaveSourcePreferenceRes struct {
	Preference SourcePreferenceView `json:"preference"`
}

type SiteProfileView struct {
	SiteKey     string    `json:"siteKey"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Revision    int64     `json:"revision"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type GetSiteProfileReq struct {
	g.Meta `path:"/api/v1/admin/site-profile" method:"get" tags:"Administration" summary:"Read the public site profile"`
}

type GetSiteProfileRes struct {
	Profile SiteProfileView `json:"profile"`
}

type UpdateSiteProfileReq struct {
	g.Meta      `path:"/api/v1/admin/site-profile" method:"put" tags:"Administration" summary:"Update the public site profile"`
	Name        string `json:"name" v:"required|length:1,80"`
	Description string `json:"description" v:"max-length:300"`
	Revision    int64  `json:"revision" v:"min:0"`
}

type UpdateSiteProfileRes struct {
	Profile SiteProfileView `json:"profile"`
}

type GetAuthorizationOverviewReq struct {
	g.Meta `path:"/api/v1/admin/authorization" method:"get" tags:"Administration" summary:"Read YTV authorization bootstrap state"`
}

type GetAuthorizationOverviewRes struct {
	CurrentUserKey        string   `json:"currentUserKey"`
	AdministratorRole     string   `json:"administratorRole"`
	Capabilities          []string `json:"capabilities"`
	InitialClaimAvailable bool     `json:"initialClaimAvailable"`
	BootstrapMode         bool     `json:"bootstrapMode"`
}
