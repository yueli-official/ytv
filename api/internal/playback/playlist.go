package playback

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/yueli-official/ytv/api/internal/source"
)

var playlistClient = &http.Client{
	Timeout:       8 * time.Second,
	CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
}
var playlistURI = regexp.MustCompile(`URI="([^"\r\n]+)"`)

// The resolver calls this only for the fixed, reviewed 7sefun playlist host/path.
// Rebasing metadata lets the browser fetch every segment directly from its CDN.
func readPlaylist(ctx context.Context, target *url.URL) (string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil)
	if err != nil {
		return "", fmt.Errorf("%w: playlist request", source.ErrUnavailable)
	}
	request.Header.Set("Accept", "application/vnd.apple.mpegurl,text/plain")
	response, err := playlistClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("%w: playlist request: %v", source.ErrUnavailable, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: playlist upstream status %d", source.ErrUnavailable, response.StatusCode)
	}
	const limit = 512 << 10
	body, err := io.ReadAll(io.LimitReader(response.Body, limit+1))
	if err != nil || len(body) > limit {
		return "", fmt.Errorf("%w: playlist size", source.ErrUnavailable)
	}
	text := strings.TrimPrefix(string(body), "\ufeff")
	if !strings.HasPrefix(text, "#EXTM3U") || !strings.Contains(text, "#EXTINF:") || strings.Contains(text, "#EXT-X-STREAM-INF:") || !strings.Contains(text, "#EXT-X-ENDLIST") {
		return "", fmt.Errorf("%w: expected a complete media playlist", source.ErrUnavailable)
	}
	absolute := func(raw string) (string, error) {
		ref, err := url.Parse(raw)
		if err != nil {
			return "", err
		}
		u := target.ResolveReference(ref)
		if (u.Scheme != "https" && u.Scheme != "http") || u.Hostname() == "" || u.User != nil {
			return "", ErrUnsafeURL
		}
		return u.String(), nil
	}
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "#") {
			lines[i], err = absolute(line)
		} else {
			lines[i] = playlistURI.ReplaceAllStringFunc(line, func(attribute string) string {
				value, resolveErr := absolute(attribute[5 : len(attribute)-1])
				if resolveErr != nil {
					err = resolveErr
					return attribute
				}
				return `URI="` + value + `"`
			})
		}
		if err != nil {
			return "", fmt.Errorf("%w: invalid playlist reference", source.ErrUnavailable)
		}
	}
	return strings.Join(lines, "\n"), nil
}
