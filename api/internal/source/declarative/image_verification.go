package declarative

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/yueli-official/ytv/api/internal/source"
	"github.com/yueli-official/ytv/api/internal/source/verification"
	"golang.org/x/net/publicsuffix"
)

// NewMutefunSearchVerification is an explicit pilot, not catalog approval. It
// accepts only the reviewed source and endpoints; it never executes source JS.
// Callers must enforce installed-source policy and bind sessions to a principal.
func NewMutefunSearchVerification(base *http.Client, definition source.Definition, query source.Query) (verification.Flow, error) {
	if definition.Key != "kazumi-mutefun" || definition.Rule == nil || definition.Rule.Name != "mutefun" || strings.TrimRight(definition.Endpoint, "/") != "https://www.2kdm.com" || strings.TrimRight(definition.Rule.BaseURL, "/") != "https://www.2kdm.com" {
		return nil, source.ErrInvalid
	}
	return newImageVerification(base, definition, query)
}

type imageVerification struct {
	client     *http.Client
	adapter    *Adapter
	definition source.Definition
	query      source.Query
	origin     *url.URL
	lifetime   context.Context
	cancel     context.CancelFunc
	primed     bool
	accepted   bool
	resumed    bool
}

func newImageVerification(base *http.Client, definition source.Definition, query source.Query) (*imageVerification, error) {
	if base == nil || definition.Rule == nil || strings.TrimSpace(query.Text) == "" || len(query.Text) > 600 {
		return nil, source.ErrInvalid
	}
	origin, err := url.Parse(definition.Endpoint)
	if err != nil || origin.Scheme != "https" || origin.Host == "" || origin.User != nil {
		return nil, source.ErrInvalid
	}
	jar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	client := *base
	client.Jar = jar
	client.Timeout = 20 * time.Second
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 || req.URL.Scheme != origin.Scheme || req.URL.Host != origin.Host {
			return source.ErrInvalid
		}
		return nil
	}
	// Copy mutable rule data so the captured operation cannot change mid-flow.
	plan := *definition.Rule
	plan.VerificationXPath = append([]string(nil), plan.VerificationXPath...)
	definition.Rule = &plan
	adapter, _ := New(&client)
	lifetime, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	return &imageVerification{client: &client, adapter: adapter, definition: definition, query: query, origin: origin, lifetime: lifetime, cancel: cancel}, nil
}

func (f *imageVerification) requestContext(ctx context.Context) (context.Context, func()) {
	child, cancel := context.WithCancel(ctx)
	stop := context.AfterFunc(f.lifetime, cancel)
	if f.lifetime.Err() != nil {
		cancel()
	}
	return child, func() { stop(); cancel() }
}
func (f *imageVerification) Close() { f.cancel() }

func (f *imageVerification) Refresh(ctx context.Context) (verification.CaptchaImage, error) {
	ctx, done := f.requestContext(ctx)
	defer done()
	if f.accepted {
		return verification.CaptchaImage{}, verification.ErrSession
	}
	if !f.primed {
		_, err := f.adapter.Search(ctx, f.definition, f.query)
		var required *verification.Required
		if !errors.As(err, &required) || required.Kind != verification.Image {
			if err != nil {
				return verification.CaptchaImage{}, err
			}
			return verification.CaptchaImage{}, verification.ErrSession
		}
		f.primed = true
	}
	body, contentType, err := f.request(ctx, http.MethodGet, "/index.php/verify/index.html", nil, 128<<10)
	if err != nil {
		return verification.CaptchaImage{}, err
	}
	media, _, err := mime.ParseMediaType(contentType)
	if err != nil || (media != "image/png" && media != "image/jpeg" && media != "image/gif") {
		return verification.CaptchaImage{}, source.ErrUnavailable
	}
	config, format, err := image.DecodeConfig(bytes.NewReader(body))
	if err != nil || config.Width < 1 || config.Height < 1 || config.Width > 1024 || config.Height > 512 || media != "image/"+format {
		return verification.CaptchaImage{}, source.ErrUnavailable
	}
	return verification.CaptchaImage{MIME: media, Bytes: body}, nil
}

func (f *imageVerification) Submit(ctx context.Context, answer string) (bool, error) {
	if !f.primed || f.accepted {
		return false, verification.ErrSession
	}
	answer = strings.TrimSpace(answer)
	if len(answer) < 1 || len(answer) > 32 {
		return false, verification.ErrAnswer
	}
	ctx, done := f.requestContext(ctx)
	defer done()
	// Observed source uses POST with type and answer in query parameters.
	// Do not log the URL or return upstream messages/HTML.
	path := "/index.php/ajax/verify_check?" + url.Values{"type": {"search"}, "verify": {answer}}.Encode()
	body, _, err := f.request(ctx, http.MethodPost, path, strings.NewReader(""), 16<<10)
	if err != nil {
		return false, err
	}
	var result struct {
		Code *int `json:"code"`
	}
	if json.Unmarshal(body, &result) != nil || result.Code == nil {
		return false, source.ErrUnavailable
	}
	if *result.Code != 1 && *result.Code != 1002 {
		return false, source.ErrUnavailable
	}
	f.accepted = *result.Code == 1
	return f.accepted, nil
}

func (f *imageVerification) Resume(ctx context.Context) (source.Page, error) {
	if !f.accepted || f.resumed {
		return source.Page{}, verification.ErrSession
	}
	f.resumed = true
	ctx, done := f.requestContext(ctx)
	defer done()
	return f.adapter.Search(ctx, f.definition, f.query)
}

func (f *imageVerification) request(ctx context.Context, method, path string, body io.Reader, limit int64) ([]byte, string, error) {
	target := f.origin.ResolveReference(mustRelativeURL(path))
	request, err := http.NewRequestWithContext(ctx, method, target.String(), body)
	if err != nil {
		return nil, "", source.ErrInvalid
	}
	request.Header.Set("User-Agent", first(f.definition.Rule.UserAgent, "YTV/1.0"))
	request.Header.Set("Referer", f.definition.Endpoint)
	request.Header.Set("Cache-Control", "no-cache")
	if method == http.MethodPost {
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		request.Header.Set("X-Requested-With", "XMLHttpRequest")
	}
	response, err := f.client.Do(request)
	if err != nil {
		return nil, "", fmt.Errorf("%w: verification request failed", source.ErrUnavailable)
	}
	defer response.Body.Close()
	if required := (verification.Cloudflare{}).Detect(verification.Response{Header: response.Header}); required != nil {
		return nil, "", required
	}
	if response.StatusCode != http.StatusOK {
		return nil, "", source.ErrUnavailable
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, limit+1))
	if err != nil || int64(len(data)) > limit {
		return nil, "", source.ErrUnavailable
	}
	return data, response.Header.Get("Content-Type"), nil
}

func mustRelativeURL(path string) *url.URL { value, _ := url.Parse(path); return value }
