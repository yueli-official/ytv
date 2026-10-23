// Package rulecatalog loads deployment-pinned external rule catalogs and
// compiles them into ordinary YTV Source Definitions.
package rulecatalog

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/yueli-official/ytv/api/internal/source"
	"github.com/yueli-official/ytv/api/internal/source/rulecompiler"
	"github.com/yueli-official/ytv/api/internal/source/rules"
)

const (
	kazumiCatalog = "Kazumi"
	maxIndexBytes = 1 << 20
	maxRuleBytes  = 256 << 10
)

var (
	commitPattern       = regexp.MustCompile(`^[0-9a-f]{40}$`)
	namePattern         = regexp.MustCompile(`^[A-Za-z0-9_-]{1,80}$`)
	approvedKazumiRules = map[string]bool{
		"7sefun": true, "aafun": true, "akianime": true,
		"moonci": true, "MXdm": true, "xfdmneo": true,
	}
)

type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

type KazumiOptions struct {
	Client      Doer
	Revision    string
	RawBaseURL  string
	MaxParallel int
}

type LoadResult struct {
	Definitions []source.Definition
	Skipped     []SkippedRule
}

type SkippedRule struct {
	Name   string
	Reason string
}

type indexEntry struct {
	Name               string `json:"name"`
	Version            string `json:"version"`
	UseNativePlayer    bool   `json:"useNativePlayer"`
	AntiCrawlerEnabled bool   `json:"antiCrawlerEnabled"`
	Author             string `json:"author"`
	LastUpdate         int64  `json:"lastUpdate"`
}

func LoadKazumi(ctx context.Context, options KazumiOptions) (LoadResult, error) {
	if options.Client == nil {
		return LoadResult{}, fmt.Errorf("Kazumi rule catalog HTTP client is required")
	}
	revision := strings.ToLower(strings.TrimSpace(options.Revision))
	if !commitPattern.MatchString(revision) {
		return LoadResult{}, fmt.Errorf("Kazumi rule catalog revision must be a full commit SHA")
	}
	base := strings.TrimSpace(options.RawBaseURL)
	if base == "" {
		base = "https://raw.githubusercontent.com/Predidit/KazumiRules/"
	}
	baseURL, err := url.Parse(base)
	if err != nil || baseURL.Scheme != "https" || baseURL.Hostname() == "" {
		return LoadResult{}, fmt.Errorf("Kazumi rule catalog base URL must be absolute HTTPS")
	}
	baseURL.Path = strings.TrimSuffix(baseURL.Path, "/") + "/" + revision + "/"

	var entries []indexEntry
	if err := fetchJSON(ctx, options.Client, baseURL.ResolveReference(&url.URL{Path: "index.json"}).String(), maxIndexBytes, &entries); err != nil {
		return LoadResult{}, fmt.Errorf("load Kazumi rule index: %w", err)
	}
	if len(entries) == 0 || len(entries) > 500 {
		return LoadResult{}, fmt.Errorf("Kazumi rule index size is invalid")
	}

	maxParallel := options.MaxParallel
	if maxParallel < 1 {
		maxParallel = 4
	}
	compiler, err := rulecompiler.New(map[string]rulecompiler.Adapter{"kazumi": rulecompiler.Kazumi{}})
	if err != nil {
		return LoadResult{}, err
	}
	results := make([]compiledEntry, len(entries))
	semaphore := make(chan struct{}, maxParallel)
	var wait sync.WaitGroup
	for index, entry := range entries {
		wait.Add(1)
		go func() {
			defer wait.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			results[index] = loadKazumiRule(ctx, options.Client, compiler, baseURL, entry, index)
		}()
	}
	wait.Wait()

	loaded := LoadResult{Definitions: make([]source.Definition, 0, len(entries)), Skipped: make([]SkippedRule, 0)}
	seen := make(map[string]bool, len(entries))
	for _, result := range results {
		if result.skipped.Reason != "" {
			loaded.Skipped = append(loaded.Skipped, result.skipped)
			continue
		}
		if seen[result.definition.Key] {
			loaded.Skipped = append(loaded.Skipped, SkippedRule{Name: result.definition.Name, Reason: "duplicate source key"})
			continue
		}
		seen[result.definition.Key] = true
		loaded.Definitions = append(loaded.Definitions, result.definition)
	}
	sort.SliceStable(loaded.Definitions, func(left, right int) bool { return loaded.Definitions[left].Name < loaded.Definitions[right].Name })
	return loaded, nil
}

func loadKazumiRule(ctx context.Context, client Doer, compiler *rulecompiler.Compiler, base *url.URL, entry indexEntry, order int) compiledEntry {
	name := strings.TrimSpace(entry.Name)
	if !namePattern.MatchString(name) {
		return compiledEntry{skipped: SkippedRule{Name: name, Reason: "invalid rule name"}}
	}
	if !approvedKazumiRules[name] {
		return compiledEntry{skipped: SkippedRule{Name: name, Reason: "playback contract is not approved for deployment"}}
	}
	var document json.RawMessage
	target := base.ResolveReference(&url.URL{Path: name + ".json"}).String()
	if err := fetchJSON(ctx, client, target, maxRuleBytes, &document); err != nil {
		return compiledEntry{skipped: SkippedRule{Name: name, Reason: err.Error()}}
	}
	return compileKazumiRule(compiler, entry, document, order)
}

func compileKazumiRule(compiler *rulecompiler.Compiler, entry indexEntry, document json.RawMessage, order int) compiledEntry {
	name := strings.TrimSpace(entry.Name)
	if !namePattern.MatchString(name) {
		return compiledEntry{skipped: SkippedRule{Name: name, Reason: "invalid rule name"}}
	}
	if !approvedKazumiRules[name] {
		return compiledEntry{skipped: SkippedRule{Name: name, Reason: "playback contract is not approved for deployment"}}
	}
	compiled, err := compiler.Compile("kazumi", document)
	if err != nil {
		return compiledEntry{skipped: SkippedRule{Name: name, Reason: err.Error()}}
	}
	rules.ApplySpecialPlayback(&compiled.Plan)
	endpoint, err := url.Parse(compiled.Plan.BaseURL)
	if err != nil || endpoint.Hostname() == "" {
		return compiledEntry{skipped: SkippedRule{Name: name, Reason: "compiled rule has no endpoint host"}}
	}
	notices := make([]source.CompatibilityNotice, 0, len(compiled.Diagnostics)+1)
	for _, diagnostic := range compiled.Diagnostics {
		notices = append(notices, source.CompatibilityNotice{Level: string(diagnostic.Severity), Code: diagnostic.Code, Message: diagnostic.Message})
	}
	notices = append(notices, source.CompatibilityNotice{
		Level: "notice", Code: "catalog_snapshot",
		Message: "来自部署固定的 Kazumi 规则快照，播放地址由已接入的解析器获取。",
	})
	return compiledEntry{definition: source.Definition{
		Key: "kazumi-" + strings.ToLower(name), Name: first(compiled.Plan.Name, name), Kind: source.KindDeclarative,
		Catalog: kazumiCatalog, CatalogVersion: strings.TrimSpace(entry.Version), Endpoint: compiled.Plan.BaseURL,
		TrustPlaybackURL: true, AllowHTTP: true,
		PlaybackHosts: append([]string{strings.ToLower(endpoint.Hostname())}, compiled.Plan.Playback.PlaybackHosts...), InitiallyUninstalled: true,
		Enabled: false, HomeVisible: false, Priority: 2000 + order, Rule: &compiled.Plan, Compatibility: notices,
	}}
}

type compiledEntry struct {
	definition source.Definition
	skipped    SkippedRule
}

func fetchJSON(ctx context.Context, client Doer, target string, limit int64, destination any) error {
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		lastErr = fetchJSONOnce(ctx, client, target, limit, destination)
		if lastErr == nil || ctx.Err() != nil {
			return lastErr
		}
	}
	return lastErr
}

func fetchJSONOnce(ctx context.Context, client Doer, target string, limit int64, destination any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "YTV/1.0")
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("upstream status %d", response.StatusCode)
	}
	decoder := json.NewDecoder(io.LimitReader(response.Body, limit+1))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return fmt.Errorf("decode JSON: %w", err)
	}
	return nil
}

func first(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
