// Package rules defines the format-neutral source plan produced by rule compilers.
// Runtime adapters consume this plan and never need to understand its source format.
package rules

import "net/http"

type Capability string

const (
	CapabilityStaticHTML      Capability = "static_html"
	CapabilityXPath           Capability = "xpath"
	CapabilityFixedPost       Capability = "fixed_post"
	CapabilityFixedReferer    Capability = "fixed_referer"
	CapabilityDirectMedia     Capability = "direct_media"
	CapabilityBrowserCapture  Capability = "browser_capture"
	CapabilityLegacyIframe    Capability = "legacy_iframe"
	CapabilityArbitraryScript Capability = "arbitrary_script"
	CapabilityVerification    Capability = "interactive_verification"
	CapabilityManifestRewrite Capability = "manifest_rewrite"
)

type Plan struct {
	Format            string
	FormatLevel       int
	Name              string
	BaseURL           string
	UserAgent         string
	Search            SearchPlan
	Title             TitlePlan
	Playback          PlaybackPlan
	Capabilities      []Capability
	VerificationXPath []string
}

type RequestPlan struct {
	Method  string
	URL     string
	Referer string
	Body    string
}

type SearchPlan struct {
	API       *SearchAPI
	Request   RequestPlan
	ListXPath string
	NameXPath string
	LinkXPath string
}

type TitlePlan struct {
	API           *ChapterAPI
	NameXPath     string
	RoadsXPath    string
	EpisodesXPath string
	MultipleRoads bool
}

// API requests are fixed by the reviewed rule document; viewers supply values only.
type APIRequest struct {
	Method   string            `json:"method"`
	URL      string            `json:"url"`
	Query    map[string]any    `json:"query"`
	Headers  map[string]string `json:"headers"`
	BodyType string            `json:"bodyType"`
	Body     map[string]any    `json:"body"`
}
type SearchAPI struct {
	Request        APIRequest `json:"request"`
	ListPath       string     `json:"listPath"`
	NamePath       string     `json:"namePath"`
	SourcePath     string     `json:"sourcePath"`
	SourceTemplate string     `json:"sourceTemplate"`
}
type ChapterAPI struct {
	RoadsPath        string     `json:"roadsPath"`
	RoadNamePath     string     `json:"roadNamePath"`
	EpisodesPath     string     `json:"episodesPath"`
	EpisodeNamePath  string     `json:"episodeNamePath"`
	EpisodeURLPath   string     `json:"episodeUrlPath"`
	EpisodePage      APIRequest `json:"episodePage"`
	Request          APIRequest `json:"request"`
	Format           string     `json:"format"`
	RoadNamesPath    string     `json:"roadNamesPath"`
	RoadEpisodesPath string     `json:"roadEpisodesPath"`
	RoadSeparator    string     `json:"roadSeparator"`
	EpisodeSeparator string     `json:"episodeSeparator"`
	FieldSeparator   string     `json:"fieldSeparator"`
}

type PlaybackPlan struct {
	UseWebview      bool
	UseNativePlayer bool
	UseLegacyParser bool
	DirectMedia     bool
	MediaXPath      string
	Extractor       string
	ParserURL       string
	PlaybackHosts   []string
}

func (plan Plan) Normalize() Plan {
	if plan.Search.Request.Method == "" {
		plan.Search.Request.Method = http.MethodGet
	}
	return plan
}
