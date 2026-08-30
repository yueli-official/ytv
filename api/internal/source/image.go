package source

import (
	"net"
	"net/url"
	"slices"
	"strings"
)

// AllowedImageURL returns only deployment-allowlisted HTTPS poster URLs. The
// browser never receives arbitrary upstream image locations.
func AllowedImageURL(definition Definition, raw string) string {
	raw = strings.TrimSpace(raw)
	if definition.Kind == KindFixture && strings.HasPrefix(raw, "/demo/") && !strings.HasPrefix(raw, "//") {
		return raw
	}
	if raw == "" || len(raw) > 2048 || len(definition.ImageHosts) == 0 {
		return ""
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme != "https" || parsed.Hostname() == "" || parsed.User != nil {
		return ""
	}
	host := strings.ToLower(parsed.Hostname())
	if address := net.ParseIP(host); address != nil && (address.IsPrivate() || address.IsLoopback() || address.IsUnspecified() || address.IsLinkLocalUnicast()) {
		return ""
	}
	allowed := slices.ContainsFunc(definition.ImageHosts, func(pattern string) bool {
		pattern = strings.ToLower(strings.TrimSpace(pattern))
		if strings.HasPrefix(pattern, "*.") {
			suffix := strings.TrimPrefix(pattern, "*")
			return strings.HasSuffix(host, suffix) && host != strings.TrimPrefix(suffix, ".")
		}
		return host == pattern
	})
	if !allowed {
		return ""
	}
	return parsed.String()
}
