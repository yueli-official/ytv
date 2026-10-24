package declarative

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/yueli-official/ytv/api/internal/source"
	"github.com/yueli-official/ytv/api/internal/source/rules"
	"github.com/yueli-official/ytv/api/internal/source/verification"
)

var (
	macPlayerPattern = regexp.MustCompile(`(?s)var\s+player_aaaa\s*=\s*(\{.*?\})\s*</script>`)
	ecConfigPattern  = regexp.MustCompile(`(?s)let\s+ConFig\s*=\s*(\{.*?\})\s*,\s*box\s*=`)
)

type macPlayerPayload struct {
	Encrypt int    `json:"encrypt"`
	URL     string `json:"url"`
	From    string `json:"from"`
}

type ecPlayerPayload struct {
	Code   int    `json:"code"`
	URL    string `json:"url"`
	Config struct {
		UID string `json:"uid"`
	} `json:"config"`
}

type cachedPlayback struct {
	URL       string
	ExpiresAt time.Time
}

func (adapter *Adapter) extractMacCMSECPlayer(ctx context.Context, definition source.Definition, plan *rules.Plan, episodeURL string) (string, error) {
	adapter.playbackMu.Lock()
	defer adapter.playbackMu.Unlock()
	if cached, exists := adapter.playbackCache[episodeURL]; exists && adapter.clock().Before(cached.ExpiresAt) {
		return cached.URL, nil
	}
	var mediaURL string
	var err error
	for attempt := 0; attempt < 3; attempt++ {
		mediaURL, err = adapter.extractMacCMSECPlayerUncached(ctx, definition, plan, episodeURL)
		if err == nil {
			break
		}
		if attempt < 2 {
			timer := time.NewTimer(time.Duration(attempt+1) * 250 * time.Millisecond)
			select {
			case <-ctx.Done():
				timer.Stop()
				return "", ctx.Err()
			case <-timer.C:
			}
		}
	}
	if err != nil {
		return "", err
	}
	adapter.playbackCache[episodeURL] = cachedPlayback{URL: mediaURL, ExpiresAt: adapter.clock().Add(5 * time.Minute)}
	return mediaURL, nil
}

func (adapter *Adapter) extractMacCMSECPlayerUncached(ctx context.Context, definition source.Definition, plan *rules.Plan, episodeURL string) (string, error) {
	page, err := adapter.fetchText(ctx, episodeURL, definition.Endpoint, definition.Endpoint)
	if err != nil {
		return "", err
	}
	match := macPlayerPattern.FindStringSubmatch(page)
	if len(match) != 2 {
		return "", fmt.Errorf("%w: MacCMS player payload is missing", source.ErrUnavailable)
	}
	var player macPlayerPayload
	if err := json.Unmarshal([]byte(match[1]), &player); err != nil {
		return "", fmt.Errorf("%w: decode MacCMS player payload", source.ErrUnavailable)
	}
	target, err := decodeMacPlayerURL(player.URL, player.Encrypt)
	if err != nil {
		return "", fmt.Errorf("%w: decode MacCMS player URL", source.ErrUnavailable)
	}
	parser, err := url.Parse(plan.Playback.ParserURL)
	if err != nil || parser.Scheme != "https" || parser.Hostname() == "" {
		return "", fmt.Errorf("%w: approved parser URL is invalid", source.ErrInvalid)
	}
	query := parser.Query()
	query.Set("from", player.From)
	query.Set("url", target)
	parser.RawQuery = query.Encode()
	parserPage, err := adapter.fetchText(ctx, parser.String(), "https://"+parser.Host, definition.Endpoint)
	if err != nil {
		return "", err
	}
	configMatch := ecConfigPattern.FindStringSubmatch(parserPage)
	if len(configMatch) != 2 {
		return "", fmt.Errorf("%w: ECPlayer config is missing", source.ErrUnavailable)
	}
	var payload ecPlayerPayload
	if err := json.Unmarshal([]byte(configMatch[1]), &payload); err != nil || payload.Code != http.StatusOK {
		return "", fmt.Errorf("%w: decode ECPlayer config", source.ErrUnavailable)
	}
	mediaURL, err := decryptECPlayerURL(payload.URL, payload.Config.UID)
	if err != nil {
		return "", fmt.Errorf("%w: decrypt ECPlayer media URL", source.ErrUnavailable)
	}
	return mediaURL, nil
}

func (adapter *Adapter) fetchText(ctx context.Context, target, approvedBase, referer string) (string, error) {
	if !sameHost(approvedBase, target) {
		return "", fmt.Errorf("%w: special playback request left approved host", source.ErrInvalid)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return "", fmt.Errorf("%w: build special playback request", source.ErrInvalid)
	}
	request.Header.Set("Accept", "text/html,application/xhtml+xml")
	request.Header.Set("User-Agent", "Mozilla/5.0")
	request.Header.Set("Referer", referer)
	response, err := adapter.client.Do(request)
	if err != nil {
		return "", fmt.Errorf("%w: special playback request failed: %v", source.ErrUnavailable, err)
	}
	defer response.Body.Close()
	if requirement := (verification.Cloudflare{}).Detect(verification.Response{Header: response.Header}); requirement != nil {
		return "", requirement
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", fmt.Errorf("%w: special playback upstream status %d", source.ErrUnavailable, response.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes+1))
	if err != nil || len(body) > maxResponseBytes {
		return "", fmt.Errorf("%w: read special playback response", source.ErrUnavailable)
	}
	return string(body), nil
}

func decodeMacPlayerURL(raw string, encryption int) (string, error) {
	value := raw
	if encryption == 2 {
		decoded, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			return "", err
		}
		value = string(decoded)
	}
	if encryption == 1 || encryption == 2 {
		value = decodeJavaScriptUnicodeEscapes(value)
		decoded, err := url.QueryUnescape(value)
		if err != nil {
			return "", err
		}
		value = decoded
	}
	if parsed, err := url.Parse(value); err != nil || parsed.Scheme != "https" || parsed.Hostname() == "" {
		return "", fmt.Errorf("decoded URL is invalid")
	}
	return value, nil
}

func decodeJavaScriptUnicodeEscapes(value string) string {
	return regexp.MustCompile(`%u([0-9A-Fa-f]{4})`).ReplaceAllStringFunc(value, func(match string) string {
		code, err := strconv.ParseInt(match[2:], 16, 32)
		if err != nil {
			return match
		}
		return string(rune(code))
	})
}

func (adapter *Adapter) extractMacCMSDirect(ctx context.Context, definition source.Definition, episodeURL string) (string, error) {
	mediaURL, err := adapter.extractMacCMSPage(ctx, definition, episodeURL)
	if err != nil {
		return "", err
	}
	if !isDirectMedia(mediaURL) {
		return "", fmt.Errorf("%w: MacCMS payload has no direct media URL", source.ErrUnavailable)
	}
	return mediaURL, nil
}

func (adapter *Adapter) extractMacCMSPage(ctx context.Context, definition source.Definition, episodeURL string) (string, error) {
	page, err := adapter.fetchText(ctx, episodeURL, definition.Endpoint, definition.Endpoint)
	if err != nil {
		return "", err
	}
	match := macPlayerPattern.FindStringSubmatch(page)
	if len(match) != 2 {
		return "", fmt.Errorf("%w: MacCMS player payload is missing", source.ErrUnavailable)
	}
	var player macPlayerPayload
	if err := json.Unmarshal([]byte(match[1]), &player); err != nil {
		return "", fmt.Errorf("%w: decode MacCMS player payload", source.ErrUnavailable)
	}
	mediaURL, err := decodeMacPlayerURL(player.URL, player.Encrypt)
	if err != nil {
		return "", fmt.Errorf("%w: MacCMS payload has no playback URL", source.ErrUnavailable)
	}
	return mediaURL, nil
}

var htmlMediaPattern = regexp.MustCompile(`https?:(?:\\/|/){2}[^"'<>[:space:]]+?\.(?:m3u8|mp4|webm)(?:\?[^"'<>[:space:]]*)?`)

func (adapter *Adapter) extractHTMLMedia(ctx context.Context, definition source.Definition, episodeURL string) (string, error) {
	page, err := adapter.fetchText(ctx, episodeURL, definition.Endpoint, definition.Endpoint)
	if err != nil {
		return "", err
	}
	match := htmlMediaPattern.FindString(page)
	if match == "" {
		return "", fmt.Errorf("%w: HTML contains no media URL", source.ErrUnavailable)
	}
	return strings.ReplaceAll(match, `\/`, "/"), nil
}

func decryptECPlayerURL(raw, uid string) (string, error) {
	key := []byte("2890" + strings.TrimSpace(uid) + "tB959C")
	iv := []byte("2F131BE91247866E")
	if len(key) != aes.BlockSize {
		return "", fmt.Errorf("invalid ECPlayer key")
	}
	ciphertext, err := base64.StdEncoding.DecodeString(raw)
	if err != nil || len(ciphertext) == 0 || len(ciphertext)%aes.BlockSize != 0 {
		return "", fmt.Errorf("invalid ECPlayer ciphertext")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	plaintext := make([]byte, len(ciphertext))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(plaintext, ciphertext)
	padding := int(plaintext[len(plaintext)-1])
	if padding < 1 || padding > aes.BlockSize || padding > len(plaintext) {
		return "", fmt.Errorf("invalid ECPlayer padding")
	}
	for _, value := range plaintext[len(plaintext)-padding:] {
		if int(value) != padding {
			return "", fmt.Errorf("invalid ECPlayer padding")
		}
	}
	result := string(plaintext[:len(plaintext)-padding])
	parsed, err := url.Parse(result)
	if err != nil || parsed.Scheme != "https" || parsed.Hostname() == "" {
		return "", fmt.Errorf("invalid ECPlayer media URL")
	}
	return result, nil
}
