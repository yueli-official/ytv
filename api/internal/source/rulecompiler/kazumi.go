package rulecompiler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/yueli-official/ytv/api/internal/source/rules"
)

type Kazumi struct{}

type kazumiDocument struct {
	SearchMode        string            `json:"searchMode"`
	ChapterMode       string            `json:"chapterMode"`
	SearchAPI         *rules.SearchAPI  `json:"searchApiConfig"`
	ChapterAPI        *rules.ChapterAPI `json:"chapterApiConfig"`
	API               scalarInt         `json:"api"`
	Name              string            `json:"name"`
	BaseURL           string            `json:"baseURL"`
	SearchURL         string            `json:"searchURL"`
	SearchList        string            `json:"searchList"`
	SearchName        string            `json:"searchName"`
	SearchResult      string            `json:"searchResult"`
	ChapterRoads      string            `json:"chapterRoads"`
	ChapterResult     string            `json:"chapterResult"`
	UserAgent         string            `json:"userAgent"`
	Referer           string            `json:"referer"`
	UsePost           bool              `json:"usePost"`
	UseWebview        bool              `json:"useWebview"`
	UseNativePlayer   bool              `json:"useNativePlayer"`
	UseLegacyParser   bool              `json:"useLegacyParser"`
	MultipleSources   bool              `json:"muliSources"`
	AdBlocker         json.RawMessage   `json:"adBlocker"`
	AntiCrawlerConfig json.RawMessage   `json:"antiCrawlerConfig"`
	YTV               struct {
		MediaXPath string `json:"mediaXPath"`
	} `json:"ytv"`
}

func (Kazumi) Compile(document json.RawMessage) (Result, error) {
	decoder := json.NewDecoder(strings.NewReader(string(document)))
	var rule kazumiDocument
	if err := decoder.Decode(&rule); err != nil {
		return Result{}, fmt.Errorf("decode document: %w", err)
	}
	if strings.TrimSpace(rule.Name) == "" {
		return Result{}, fmt.Errorf("name is required")
	}
	if rule.API < 1 {
		return Result{}, fmt.Errorf("api level must be at least 1")
	}
	if err := requireAbsoluteHTTPS("baseURL", rule.BaseURL); err != nil {
		return Result{}, err
	}
	required := map[string]string{}
	if rule.SearchMode == "api" {
		if rule.SearchAPI == nil {
			return Result{}, fmt.Errorf("searchApiConfig is required")
		}
		if err := validateAPI(rule.SearchAPI.Request); err != nil {
			return Result{}, err
		}
		required["listPath"], required["namePath"], required["sourcePath"] = rule.SearchAPI.ListPath, rule.SearchAPI.NamePath, rule.SearchAPI.SourcePath
	} else {
		for field, value := range map[string]string{
			"searchURL": rule.SearchURL, "searchList": rule.SearchList,
			"searchName": rule.SearchName, "searchResult": rule.SearchResult,
		} {
			required[field] = value
		}
	}
	if rule.ChapterMode == "api" {
		if rule.ChapterAPI == nil {
			return Result{}, fmt.Errorf("chapterApiConfig is required")
		}
		if err := validateAPI(rule.ChapterAPI.Request); err != nil {
			return Result{}, err
		}
		if rule.ChapterAPI.Format != "delimited" && rule.ChapterAPI.Format != "nested" {
			return Result{}, fmt.Errorf("unsupported chapter API format %q", rule.ChapterAPI.Format)
		}
		fields := map[string]string{"roadNamesPath": rule.ChapterAPI.RoadNamesPath, "roadEpisodesPath": rule.ChapterAPI.RoadEpisodesPath, "roadSeparator": rule.ChapterAPI.RoadSeparator, "episodeSeparator": rule.ChapterAPI.EpisodeSeparator, "fieldSeparator": rule.ChapterAPI.FieldSeparator}
		if rule.ChapterAPI.Format == "nested" {
			fields = map[string]string{"roadsPath": rule.ChapterAPI.RoadsPath, "episodesPath": rule.ChapterAPI.EpisodesPath, "episodeNamePath": rule.ChapterAPI.EpisodeNamePath, "episodeUrlPath": rule.ChapterAPI.EpisodeURLPath, "episodePage.url": rule.ChapterAPI.EpisodePage.URL}
		}
		for field, value := range fields {
			required[field] = value
		}
	} else {
		required["chapterRoads"], required["chapterResult"] = rule.ChapterRoads, rule.ChapterResult
	}
	for field, value := range required {
		if strings.TrimSpace(value) == "" {
			return Result{}, fmt.Errorf("%s is required", field)
		}
	}

	method := http.MethodGet
	capabilities := []rules.Capability{rules.CapabilityStaticHTML, rules.CapabilityXPath}
	diagnostics := make([]Diagnostic, 0)
	if rule.UsePost {
		method = http.MethodPost
		capabilities = append(capabilities, rules.CapabilityFixedPost)
		diagnostics = append(diagnostics, Diagnostic{Severity: SeverityNotice, Code: "fixed_post", Message: "搜索使用固定 POST；部署定义需要确认请求正文模板。"})
	}
	if strings.TrimSpace(rule.Referer) != "" {
		capabilities = append(capabilities, rules.CapabilityFixedReferer)
		diagnostics = append(diagnostics, Diagnostic{Severity: SeverityNotice, Code: "fixed_referer", Message: "播放可能依赖固定 Referer；浏览器直连兼容性需要实测。"})
	}
	if rule.UseNativePlayer {
		capabilities = append(capabilities, rules.CapabilityDirectMedia)
	} else if rule.UseWebview {
		capabilities = append(capabilities, rules.CapabilityBrowserCapture)
		diagnostics = append(diagnostics, Diagnostic{Severity: SeverityWarning, Code: "browser_capture", Message: "规则依赖浏览器捕获媒体请求；目录能力可用，播放需要来源专用解析 Adapter。"})
	}
	if rule.UseLegacyParser {
		capabilities = append(capabilities, rules.CapabilityLegacyIframe)
		diagnostics = append(diagnostics, Diagnostic{Severity: SeverityWarning, Code: "legacy_iframe", Message: "规则依赖 iframe 解析；Web 环境可能受跨域或 CSP 限制。"})
	}
	if present(rule.AdBlocker) {
		capabilities = append(capabilities, rules.CapabilityManifestRewrite)
		diagnostics = append(diagnostics, Diagnostic{Severity: SeverityUnsupported, Code: "manifest_rewrite", Message: "规则要求改写媒体清单，违反 YTV 不代理或改写第三方视频流的产品合同。"})
	}
	if capability, diagnostic, err := verificationAdvice(rule.AntiCrawlerConfig); err != nil {
		return Result{}, err
	} else if capability != "" {
		capabilities = append(capabilities, capability)
		diagnostics = append(diagnostics, diagnostic)
	}

	return Result{Plan: rules.Plan{
		Format: "kazumi", FormatLevel: int(rule.API), Name: strings.TrimSpace(rule.Name),
		BaseURL: strings.TrimSpace(rule.BaseURL), UserAgent: strings.TrimSpace(rule.UserAgent),
		Search: rules.SearchPlan{API: rule.SearchAPI, Request: rules.RequestPlan{
			Method: method, URL: strings.TrimSpace(rule.SearchURL), Referer: strings.TrimSpace(rule.Referer),
		}, ListXPath: rule.SearchList, NameXPath: rule.SearchName, LinkXPath: rule.SearchResult},
		Title:             rules.TitlePlan{API: rule.ChapterAPI, RoadsXPath: rule.ChapterRoads, EpisodesXPath: rule.ChapterResult, MultipleRoads: rule.MultipleSources},
		Playback:          rules.PlaybackPlan{UseWebview: rule.UseWebview, UseNativePlayer: rule.UseNativePlayer, UseLegacyParser: rule.UseLegacyParser, MediaXPath: strings.TrimSpace(rule.YTV.MediaXPath)},
		Capabilities:      capabilities,
		VerificationXPath: verificationSelectors(rule.AntiCrawlerConfig),
	}, Diagnostics: diagnostics}, nil
}

func validateAPI(request rules.APIRequest) error {
	if request.Method != "GET" && !(request.Method == "POST" && request.BodyType == "json") {
		return fmt.Errorf("unsupported rule API method %q", request.Method)
	}
	return requireAbsoluteHTTPS("API request URL", request.URL)
}

func verificationSelectors(raw json.RawMessage) []string {
	var config struct {
		Enabled bool   `json:"enabled"`
		Type    int    `json:"captchaType"`
		Image   string `json:"captchaImage"`
		Input   string `json:"captchaInput"`
		Button  string `json:"captchaButton"`
	}
	if json.Unmarshal(raw, &config) != nil || !config.Enabled || (config.Type != 0 && config.Type != 1) {
		return nil
	}
	if strings.TrimSpace(config.Image) == "" || strings.TrimSpace(config.Input) == "" || strings.TrimSpace(config.Button) == "" {
		return nil
	}
	return []string{config.Image, config.Input, config.Button}
}

// Configuration describes a possible requirement, not a challenge observed in a request.
// No supplied script is executed and enabled verification remains unsupported.
func verificationAdvice(raw json.RawMessage) (rules.Capability, Diagnostic, error) {
	if !present(raw) {
		return "", Diagnostic{}, nil
	}
	var config struct {
		Enabled      *bool  `json:"enabled"`
		Type         int    `json:"captchaType"`
		Script       string `json:"captchaScript"`
		LegacyScript string `json:"script"`
	}
	if err := json.Unmarshal(raw, &config); err != nil {
		return "", Diagnostic{}, fmt.Errorf("invalid antiCrawlerConfig: %w", err)
	}
	if config.Enabled != nil && !*config.Enabled {
		return "", Diagnostic{}, nil
	}
	if strings.TrimSpace(config.LegacyScript) != "" || (config.Enabled != nil && *config.Enabled && config.Type == 3) {
		return rules.CapabilityArbitraryScript, Diagnostic{Severity: SeverityUnsupported, Code: "arbitrary_script", Message: "规则启用了自定义验证脚本；当前服务端不执行来源脚本。"}, nil
	}
	if config.Enabled == nil {
		return "", Diagnostic{}, nil
	}
	message := "规则启用了来源验证；当前网页端尚不支持完成此验证，请选择其他来源。"
	if config.Type == 0 || config.Type == 1 {
		message = "规则启用了图片验证码；当前网页端尚不支持提交验证码，请选择其他来源。"
	}
	if config.Type == 2 {
		message = "规则启用了浏览器按钮验证；当前服务端不能复用用户浏览器的验证状态。"
	}
	return rules.CapabilityVerification, Diagnostic{Severity: SeverityUnsupported, Code: "interactive_verification", Message: message}, nil
}

type scalarInt int

func (value *scalarInt) UnmarshalJSON(data []byte) error {
	text := strings.Trim(strings.TrimSpace(string(data)), `"`)
	parsed, err := strconv.Atoi(text)
	if err != nil {
		return fmt.Errorf("invalid integer %q", text)
	}
	*value = scalarInt(parsed)
	return nil
}

func requireAbsoluteHTTPS(field, raw string) error {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Scheme != "https" || parsed.Hostname() == "" {
		return fmt.Errorf("%s must be an absolute HTTPS URL", field)
	}
	return nil
}

func present(raw json.RawMessage) bool {
	value := strings.TrimSpace(string(raw))
	return value != "" && value != "null" && value != "{}" && value != "[]" && value != "false"
}
