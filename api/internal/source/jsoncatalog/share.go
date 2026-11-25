package jsoncatalog

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
)

// These metadata pages were reviewed individually; this is not a viewer URL parser.
var sharePages = map[string]struct{ host, prefix, variable string }{
	"jszy":    {"vv.jisuzyv.com", "/play/", "vid"},
	"subo":    {"play.xluuss.com", "/play/", "vid"},
	"xinlang": {"play.xluuss.com", "/play/", "vid"},
	"haohua":  {"play.hhuus.com", "/play/", "vid"},
	"guangsu": {"v.gsuus.com", "/play/", "vid"},
	"lzi":     {"vip1.lz-cdn8.com", "/share/", "main"},
}

var shareMedia = regexp.MustCompile(`\b(?:const|var|let)\s+(vid|main)\s*=\s*(?:("(?:[^"\\]|\\.)*")|'([^'\\\r\n]*)')\s*;`)

func isDirectMedia(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	path := strings.ToLower(u.Path)
	return strings.HasSuffix(path, ".m3u8") || strings.HasSuffix(path, ".mp4") || strings.HasSuffix(path, ".webm")
}

func isReviewedSharePage(key, raw string) bool {
	policy, ok := sharePages[key]
	page, err := url.Parse(raw)
	return ok && err == nil && page.Scheme == "https" && page.Host == policy.host && page.User == nil && strings.HasPrefix(page.Path, policy.prefix)
}

func (adapter *Adapter) resolveSharePage(ctx context.Context, key, raw string) (string, error) {
	if !isReviewedSharePage(key, raw) {
		return "", fmt.Errorf("%w: unsupported catalog playback page", source.ErrUnavailable)
	}
	policy := sharePages[key]
	page, _ := url.Parse(raw)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, page.String(), nil)
	if err != nil {
		return "", fmt.Errorf("%w: invalid playback page", source.ErrInvalid)
	}
	request.Header.Set("Accept", "text/html")
	request.Header.Set("User-Agent", "Mozilla/5.0")
	response, err := adapter.pageClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("%w: playback page request: %v", source.ErrUnavailable, err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 || !strings.Contains(strings.ToLower(response.Header.Get("Content-Type")), "text/html") {
		return "", fmt.Errorf("%w: playback page response status %d", source.ErrUnavailable, response.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, (1<<20)+1))
	if err != nil || len(body) > 1<<20 {
		return "", fmt.Errorf("%w: playback page size", source.ErrUnavailable)
	}
	for _, match := range shareMedia.FindAllSubmatch(body, -1) {
		if string(match[1]) != policy.variable {
			continue
		}
		var value string
		if len(match[2]) > 0 {
			if json.Unmarshal(match[2], &value) != nil {
				continue
			}
		} else {
			value = string(match[3])
		}
		ref, parseErr := url.Parse(value)
		if parseErr != nil {
			continue
		}
		media := page.ResolveReference(ref)
		if media.Scheme != "https" || media.Host != page.Host || media.User != nil || !strings.HasSuffix(strings.ToLower(media.Path), ".m3u8") {
			continue
		}
		return media.String(), nil
	}
	return "", fmt.Errorf("%w: playback page has no declared HLS URL", source.ErrUnavailable)
}
