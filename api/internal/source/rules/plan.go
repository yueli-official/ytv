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
	Request   RequestPlan
	ListXPath string
	NameXPath string
	LinkXPath string
}

type TitlePlan struct {
	NameXPath     string
	RoadsXPath    string
	EpisodesXPath string
	MultipleRoads bool
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
