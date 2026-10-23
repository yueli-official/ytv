// Package verification identifies observed upstream challenges without solving
// them or sharing browser credentials. Detectors never execute upstream scripts.
package verification

import (
	"errors"
	"net/http"
	"strings"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

var ErrRequired = errors.New("source verification is required")

type Kind string

const (
	Image   Kind = "image"
	Browser Kind = "browser"
)

// Required is an observed requirement, not a claim that this host can solve it.
// It deliberately contains no upstream HTML, URL, cookies or executable data.
type Required struct{ Kind Kind }

func (r *Required) Error() string { return ErrRequired.Error() }
func (r *Required) Unwrap() error { return ErrRequired }

type Response struct {
	Header   http.Header
	Document *html.Node
}
type Detector interface{ Detect(Response) *Required }

type Chain []Detector

func (c Chain) Detect(response Response) *Required {
	for _, detector := range c {
		if result := detector.Detect(response); result != nil {
			return result
		}
	}
	return nil
}

type Cloudflare struct{}

func (Cloudflare) Detect(response Response) *Required {
	if strings.EqualFold(strings.TrimSpace(response.Header.Get("cf-mitigated")), "challenge") {
		return &Required{Kind: Browser}
	}
	return nil
}

// DOM requires every deployment-reviewed selector to match. A generic word such
// as captcha in a synopsis, or a HTTP 403 alone, must never trigger detection.
type DOM struct {
	Kind Kind
	All  []string
}

func (d DOM) Detect(response Response) *Required {
	if response.Document == nil || len(d.All) == 0 {
		return nil
	}
	for _, selector := range d.All {
		if strings.TrimSpace(selector) == "" {
			return nil
		}
		node, err := htmlquery.Query(response.Document, selector)
		if err != nil || node == nil {
			return nil
		}
	}
	return &Required{Kind: d.Kind}
}
