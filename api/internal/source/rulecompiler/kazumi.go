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
	API               scalarInt       `json:"api"`
	Name              string          `json:"name"`
	BaseURL           string          `json:"baseURL"`
	SearchURL         string          `json:"searchURL"`
	SearchList        string          `json:"searchList"`
	SearchName        string          `json:"searchName"`
	SearchResult      string          `json:"searchResult"`
	ChapterRoads      string          `json:"chapterRoads"`
	ChapterResult     string          `json:"chapterResult"`
	UserAgent         string          `json:"userAgent"`
	Referer           string          `json:"referer"`
	UsePost           bool            `json:"usePost"`
	UseWebview        bool            `json:"useWebview"`
	UseNativePlayer   bool            `json:"useNativePlayer"`
	UseLegacyParser   bool            `json:"useLegacyParser"`
	MultipleSources   bool            `json:"muliSources"`
	AdBlocker         json.RawMessage `json:"adBlocker"`
	AntiCrawlerConfig json.RawMessage `json:"antiCrawlerConfig"`
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
	for field, value := range map[string]string{
		"searchURL": rule.SearchURL, "searchList": rule.SearchList,
		"searchName": rule.SearchName, "searchResult": rule.SearchResult,
		"chapterRoads": rule.ChapterRoads, "chapterResult": rule.ChapterResult,
	} {
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
	if present(rule.AntiCrawlerConfig) {
		capabilities = append(capabilities, rules.CapabilityArbitraryScript)
		diagnostics = append(diagnostics, Diagnostic{Severity: SeverityUnsupported, Code: "arbitrary_script", Message: "规则包含可执行反爬脚本；必须改写为来源专用、部署审核的 Adapter。"})
	}

	return Result{Plan: rules.Plan{
		Format: "kazumi", FormatLevel: int(rule.API), Name: strings.TrimSpace(rule.Name),
		BaseURL: strings.TrimSpace(rule.BaseURL), UserAgent: strings.TrimSpace(rule.UserAgent),
		Search: rules.SearchPlan{Request: rules.RequestPlan{
			Method: method, URL: strings.TrimSpace(rule.SearchURL), Referer: strings.TrimSpace(rule.Referer),
		}, ListXPath: rule.SearchList, NameXPath: rule.SearchName, LinkXPath: rule.SearchResult},
		Title:        rules.TitlePlan{RoadsXPath: rule.ChapterRoads, EpisodesXPath: rule.ChapterResult, MultipleRoads: rule.MultipleSources},
		Playback:     rules.PlaybackPlan{UseWebview: rule.UseWebview, UseNativePlayer: rule.UseNativePlayer, UseLegacyParser: rule.UseLegacyParser, MediaXPath: strings.TrimSpace(rule.YTV.MediaXPath)},
		Capabilities: capabilities,
	}, Diagnostics: diagnostics}, nil
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
