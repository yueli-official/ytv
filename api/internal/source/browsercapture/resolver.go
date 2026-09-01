// Package browsercapture resolves script-generated media URLs in a bounded,
// reusable headless Chrome pool. It never downloads or proxies media bytes.
package browsercapture

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

var ErrMediaNotFound = errors.New("browser capture found no media URL")

type Request struct {
	PageURL    string
	LegacyMode bool
}

type Result struct {
	URL string
}

type Resolver interface {
	Resolve(context.Context, Request) (Result, error)
}

type Options struct {
	Executable    string
	MaxConcurrent int
	Timeout       time.Duration
	NoSandbox     bool
}

type ChromeResolver struct {
	allocatorContext context.Context
	allocatorCancel  context.CancelFunc
	browserContext   context.Context
	browserCancel    context.CancelFunc
	slots            chan struct{}
	timeout          time.Duration
	closeOnce        sync.Once
}

func New(options Options) (*ChromeResolver, error) {
	executable, err := resolveExecutable(options.Executable)
	if err != nil {
		return nil, err
	}
	maxConcurrent := options.MaxConcurrent
	if maxConcurrent < 1 {
		maxConcurrent = 2
	}
	timeout := options.Timeout
	if timeout <= 0 {
		timeout = 20 * time.Second
	}
	allocatorOptions := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(executable),
		chromedp.Flag("headless", "new"),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("mute-audio", true),
		chromedp.Flag("autoplay-policy", "no-user-gesture-required"),
		chromedp.Flag("no-first-run", true),
		chromedp.WindowSize(1440, 1000),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/149 Safari/537.36"),
	)
	if options.NoSandbox {
		allocatorOptions = append(allocatorOptions, chromedp.NoSandbox)
	}
	allocatorContext, allocatorCancel := chromedp.NewExecAllocator(context.Background(), allocatorOptions...)
	browserContext, browserCancel := chromedp.NewContext(allocatorContext)
	if err := chromedp.Run(browserContext); err != nil {
		browserCancel()
		allocatorCancel()
		return nil, fmt.Errorf("start browser capture Chrome: %w", err)
	}
	return &ChromeResolver{
		allocatorContext: allocatorContext, allocatorCancel: allocatorCancel,
		browserContext: browserContext, browserCancel: browserCancel,
		slots: make(chan struct{}, maxConcurrent), timeout: timeout,
	}, nil
}

func (resolver *ChromeResolver) Resolve(ctx context.Context, request Request) (Result, error) {
	parsed, err := url.Parse(strings.TrimSpace(request.PageURL))
	if err != nil || parsed.Scheme != "https" || parsed.Hostname() == "" {
		return Result{}, fmt.Errorf("browser capture page URL is invalid")
	}
	select {
	case resolver.slots <- struct{}{}:
		defer func() { <-resolver.slots }()
	case <-ctx.Done():
		return Result{}, ctx.Err()
	}
	tabContext, tabCancel := chromedp.NewContext(resolver.browserContext)
	defer tabCancel()
	captureContext, cancel := context.WithTimeout(tabContext, resolver.timeout)
	defer cancel()

	result := make(chan string, 1)
	var completeOnce sync.Once
	complete := func(candidate string) {
		candidate = strings.TrimSpace(candidate)
		if !acceptableCandidate(candidate) {
			return
		}
		completeOnce.Do(func() { result <- candidate })
	}
	responses := sync.Map{}
	chromedp.ListenTarget(captureContext, func(event any) {
		switch value := event.(type) {
		case *network.EventRequestWillBeSent:
			if directMediaURL(value.Request.URL) || rangeMediaRequest(value.Request.URL, value.Request.Headers) {
				complete(value.Request.URL)
			}
		case *network.EventResponseReceived:
			if value.Type == network.ResourceTypeXHR || value.Type == network.ResourceTypeFetch {
				responses.Store(value.RequestID, value.Response.URL)
			}
			if mediaMIME(value.Response.MimeType) {
				complete(value.Response.URL)
			}
		case *network.EventLoadingFinished:
			rawURL, exists := responses.Load(value.RequestID)
			if !exists {
				return
			}
			go func(requestID network.RequestID, responseURL string) {
				body, bodyErr := network.GetResponseBody(requestID).Do(captureContext)
				if bodyErr == nil && strings.HasPrefix(strings.TrimSpace(string(body)), "#EXTM3U") {
					complete(responseURL)
				}
			}(value.RequestID, rawURL.(string))
		case *runtime.EventConsoleAPICalled:
			for _, argument := range value.Args {
				if argument.Type == runtime.TypeString {
					var text string
					if json.Unmarshal(argument.Value, &text) == nil && strings.HasPrefix(text, capturePrefix) {
						complete(strings.TrimPrefix(text, capturePrefix))
					}
				}
			}
		}
	})

	if err := chromedp.Run(captureContext,
		network.Enable(),
		runtime.Enable(),
		chromedp.ActionFunc(func(ctx context.Context) error {
			_, err := page.AddScriptToEvaluateOnNewDocument(captureScript).Do(ctx)
			return err
		}),
	); err != nil {
		return Result{}, fmt.Errorf("browser capture setup: %w", err)
	}
	navigation := make(chan error, 1)
	go func() {
		navigation <- chromedp.Run(captureContext,
			chromedp.ActionFunc(func(ctx context.Context) error {
				_, _, errorText, _, err := page.Navigate(request.PageURL).Do(ctx)
				if err != nil {
					return err
				}
				if errorText != "" {
					return errors.New(errorText)
				}
				return nil
			}),
		)
	}()
	select {
	case mediaURL := <-result:
		return Result{URL: mediaURL}, nil
	case err := <-navigation:
		if err != nil {
			return Result{}, fmt.Errorf("browser capture navigation: %w", err)
		}
		select {
		case mediaURL := <-result:
			return Result{URL: mediaURL}, nil
		case <-captureContext.Done():
			return Result{}, ErrMediaNotFound
		}
	case <-captureContext.Done():
		if errors.Is(captureContext.Err(), context.DeadlineExceeded) {
			return Result{}, ErrMediaNotFound
		}
		return Result{}, captureContext.Err()
	}
}

func (resolver *ChromeResolver) Close() {
	resolver.closeOnce.Do(func() {
		resolver.browserCancel()
		resolver.allocatorCancel()
	})
}

func resolveExecutable(configured string) (string, error) {
	if value := strings.TrimSpace(configured); value != "" {
		absolute, err := filepath.Abs(value)
		if err != nil {
			return "", err
		}
		if info, statErr := os.Stat(absolute); statErr != nil || info.IsDir() {
			return "", fmt.Errorf("browser capture executable is unavailable: %s", absolute)
		}
		return absolute, nil
	}
	for _, name := range []string{"google-chrome", "chromium", "chromium-browser", "chrome", "msedge"} {
		if value, err := exec.LookPath(name); err == nil {
			return value, nil
		}
	}
	return "", fmt.Errorf("YTV_BROWSER_EXECUTABLE is required for browser capture")
}

func directMediaURL(raw string) bool {
	parsed, err := url.Parse(raw)
	if err != nil {
		return false
	}
	path := strings.ToLower(parsed.Path)
	return strings.HasSuffix(path, ".m3u8") || strings.HasSuffix(path, ".mp4") || strings.HasSuffix(path, ".webm")
}

func rangeMediaRequest(raw string, headers network.Headers) bool {
	if !acceptableCandidate(raw) {
		return false
	}
	for name, value := range headers {
		if strings.EqualFold(name, "range") && strings.HasPrefix(strings.ToLower(fmt.Sprint(value)), "bytes=") {
			return true
		}
	}
	return false
}

func mediaMIME(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return strings.HasPrefix(value, "video/") || strings.Contains(value, "mpegurl")
}

func acceptableCandidate(raw string) bool {
	parsed, err := url.Parse(raw)
	if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Hostname() == "" {
		return false
	}
	lower := strings.ToLower(raw)
	for _, blocked := range []string{"googleads", "googlesyndication", "doubleclick", "adtrafficquality", "adkwai"} {
		if strings.Contains(lower, blocked) {
			return false
		}
	}
	return true
}

const capturePrefix = "YTV_MEDIA:"

const captureScript = `(() => {
  if (window.__ytvCaptureInstalled) return;
  window.__ytvCaptureInstalled = true;
  const emit = value => {
    if (typeof value === "string" && /^https?:\/\//i.test(value) && !value.startsWith("blob:")) {
      console.log("YTV_MEDIA:" + value);
    }
  };
  const inspectText = (response, text) => {
    if (typeof text === "string" && text.trimStart().startsWith("#EXTM3U")) emit(response.url);
  };
  const originalText = Response.prototype.text;
  Response.prototype.text = function() {
    return originalText.call(this).then(text => { inspectText(this, text); return text; });
  };
  const originalOpen = XMLHttpRequest.prototype.open;
  XMLHttpRequest.prototype.open = function(method, url) {
    this.addEventListener("load", () => {
      try { if (String(this.responseText).trimStart().startsWith("#EXTM3U")) emit(new URL(url, location.href).href); } catch {}
    });
    return originalOpen.apply(this, arguments);
  };
  const inspectVideo = video => {
    emit(video.currentSrc || video.src);
    for (const source of video.querySelectorAll("source")) emit(source.src);
    try { video.muted = true; video.play().catch(() => {}); } catch {}
  };
  const scan = root => {
    if (root?.matches?.("video")) inspectVideo(root);
    root?.querySelectorAll?.("video").forEach(inspectVideo);
  };
  new MutationObserver(records => records.forEach(record => {
    if (record.type === "attributes") scan(record.target);
    record.addedNodes.forEach(scan);
  })).observe(document.documentElement, { subtree: true, childList: true, attributes: true, attributeFilter: ["src"] });
  document.addEventListener("DOMContentLoaded", () => scan(document));
})();`
