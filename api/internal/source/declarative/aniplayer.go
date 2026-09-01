package declarative

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/yueli-official/ytv/api/internal/source"
	"github.com/yueli-official/ytv/api/internal/source/rules"
)

var aniPlayerConfigPattern = regexp.MustCompile(`(?s)var\s+config\s*=\s*\{(.*?)\}`)

type aniPlayerResponse struct {
	Code string `json:"code"`
	URL  string `json:"url"`
}

func (adapter *Adapter) extractAniPlayer(ctx context.Context, definition source.Definition, plan *rules.Plan, episodeURL string) (string, error) {
	episodePage, err := adapter.fetchText(ctx, episodeURL, definition.Endpoint, definition.Endpoint)
	if err != nil {
		return "", err
	}
	playerMatch := macPlayerPattern.FindStringSubmatch(episodePage)
	if len(playerMatch) != 2 {
		return "", fmt.Errorf("%w: AniPlayer source payload is missing", source.ErrUnavailable)
	}
	var player macPlayerPayload
	if err := json.Unmarshal([]byte(playerMatch[1]), &player); err != nil || strings.TrimSpace(player.URL) == "" {
		return "", fmt.Errorf("%w: decode AniPlayer source payload", source.ErrUnavailable)
	}
	parser, err := url.Parse(plan.Playback.ParserURL)
	if err != nil || parser.Scheme != "https" || parser.Hostname() == "" {
		return "", fmt.Errorf("%w: AniPlayer URL is invalid", source.ErrInvalid)
	}
	query := parser.Query()
	query.Set("url", player.URL)
	parser.RawQuery = query.Encode()
	parserPage, err := adapter.fetchText(ctx, parser.String(), "https://"+parser.Host, definition.Endpoint)
	if err != nil {
		return "", err
	}
	config := aniPlayerConfigPattern.FindStringSubmatch(parserPage)
	if len(config) != 2 {
		return "", fmt.Errorf("%w: AniPlayer config is missing", source.ErrUnavailable)
	}
	key := quotedConfigValue(config[1], "key")
	timestamp := quotedConfigValue(config[1], "time")
	if key == "" || timestamp == "" {
		return "", fmt.Errorf("%w: AniPlayer credentials are missing", source.ErrUnavailable)
	}
	form := url.Values{"url": {player.URL}, "time": {timestamp}, "key": {key}}
	endpoint := parser.ResolveReference(&url.URL{Path: "/api_config.php"}).String()
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("%w: build AniPlayer request", source.ErrInvalid)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("User-Agent", "Mozilla/5.0")
	request.Header.Set("Referer", parser.String())
	response, err := adapter.client.Do(request)
	if err != nil {
		return "", fmt.Errorf("%w: AniPlayer request failed: %v", source.ErrUnavailable, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: AniPlayer status %d", source.ErrUnavailable, response.StatusCode)
	}
	var result aniPlayerResponse
	if err := json.NewDecoder(io.LimitReader(response.Body, maxResponseBytes)).Decode(&result); err != nil || result.Code != "200" {
		return "", fmt.Errorf("%w: decode AniPlayer response", source.ErrUnavailable)
	}
	return result.URL, nil
}

func quotedConfigValue(config, name string) string {
	pattern := regexp.MustCompile(`"` + regexp.QuoteMeta(name) + `"\s*:\s*"([^"]*)"`)
	match := pattern.FindStringSubmatch(config)
	if len(match) != 2 {
		return ""
	}
	return match[1]
}
