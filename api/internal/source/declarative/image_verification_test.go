package declarative

import (
	"context"
	"errors"
	"fmt"
	"github.com/yueli-official/ytv/api/internal/source"
	"github.com/yueli-official/ytv/api/internal/source/rules"
	"github.com/yueli-official/ytv/api/internal/source/verification"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestImageVerificationKeepsCookiesIsolatedAndResumesOriginalSearch(t *testing.T) {
	var mu sync.Mutex
	next := 0
	verified := map[string]bool{}
	images := map[string]int{}
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		cookie, err := r.Cookie("sid")
		id := ""
		if err == nil {
			id = cookie.Value
		}
		if id == "" {
			next++
			id = fmt.Sprint(next)
			http.SetCookie(w, &http.Cookie{Name: "sid", Value: id, Path: "/", Secure: true, HttpOnly: true})
		}
		switch r.URL.Path {
		case "/search":
			if r.URL.Query().Get("q") != "original" {
				t.Error("changed captured query")
			}
			if verified[id] {
				fmt.Fprint(w, `<li><a href="/title/1">Result</a></li>`)
			} else {
				fmt.Fprint(w, `<title>系统安全验证</title><input name="verify">`)
			}
		case "/index.php/verify/index.html":
			images[id]++
			w.Header().Set("Content-Type", "image/png")
			img := image.NewRGBA(image.Rect(0, 0, 80, 24))
			img.Set(0, 0, color.White)
			png.Encode(w, img)
		case "/index.php/ajax/verify_check":
			if r.Method != "POST" || r.URL.Query().Get("type") != "search" || images[id] == 0 {
				t.Error("invalid session protocol")
			}
			w.Header().Set("Content-Type", "application/json")
			if r.URL.Query().Get("verify") == "right" {
				verified[id] = true
				fmt.Fprint(w, `{"code":1}`)
			} else {
				fmt.Fprint(w, `{"code":1002,"msg":"错误"}`)
			}
		}
	}))
	defer server.Close()
	def := source.Definition{Key: "fixture", Endpoint: server.URL, Rule: &rules.Plan{BaseURL: server.URL, VerificationXPath: []string{"//title[text()='系统安全验证']", "//input[@name='verify']"}, Search: rules.SearchPlan{Request: rules.RequestPlan{URL: "/search?q=@keyword"}, ListXPath: "//li", NameXPath: "/a/text()", LinkXPath: "/a/@href"}}}
	ctx := context.Background()
	sessions := verification.NewSessions()
	defer sessions.Close()
	a, _ := newImageVerification(server.Client(), def, source.Query{Text: "original"})
	b, _ := newImageVerification(server.Client(), def, source.Query{Text: "original"})
	first, err := sessions.Start(ctx, "alice", a)
	if err != nil {
		t.Fatal(err)
	}
	second, err := sessions.Start(ctx, "bob", b)
	if err != nil {
		t.Fatal(err)
	}
	if first.Image.MIME != "image/png" || len(first.Image.Bytes) == 0 {
		t.Fatal("no image")
	}
	if _, err = sessions.Submit(ctx, "alice", first.ID, "wrong"); err != nil {
		t.Fatal(err)
	}
	if _, err = sessions.Refresh(ctx, "alice", first.ID); err != nil {
		t.Fatal(err)
	}
	got, err := sessions.Submit(ctx, "alice", first.ID, "right")
	if err != nil || !got.Accepted || len(got.Page.Items) != 1 {
		t.Fatalf("result=%#v err=%v", got, err)
	}
	if b.accepted {
		t.Fatal("shared accepted state")
	}
	if _, err = sessions.Refresh(ctx, "bob", second.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = sessions.Submit(ctx, "alice", first.ID, "right"); !errors.Is(err, verification.ErrSession) {
		t.Fatal("replayed twice")
	}
	mu.Lock()
	defer mu.Unlock()
	if next != 2 || len(verified) != 1 || images["1"] != 2 || images["2"] != 2 {
		t.Fatalf("cookies leaked: users=%d verified=%v images=%v", next, verified, images)
	}
}

func TestImageVerificationRejectsWrongSourceAndHTMLImage(t *testing.T) {
	if _, err := NewMutefunSearchVerification(http.DefaultClient, source.Definition{Key: "untrusted"}, source.Query{Text: "test"}); err == nil {
		t.Fatal("unapproved source accepted")
	}
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/search" {
			fmt.Fprint(w, `<input name="verify">`)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		fmt.Fprint(w, `<script>not an image</script>`)
	}))
	defer server.Close()
	def := source.Definition{Endpoint: server.URL, Rule: &rules.Plan{BaseURL: server.URL, VerificationXPath: []string{"//input[@name='verify']"}, Search: rules.SearchPlan{Request: rules.RequestPlan{URL: "/search"}}}}
	flow, _ := newImageVerification(server.Client(), def, source.Query{Text: "test"})
	defer flow.Close()
	if _, err := flow.Refresh(context.Background()); err == nil {
		t.Fatal("HTML accepted as image")
	}
}

func TestImageVerificationRejectsCrossOriginRedirectAndAmbiguousAnswer(t *testing.T) {
	destinationCalled := false
	destination := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { destinationCalled = true }))
	defer destination.Close()
	mode := "redirect"
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if mode == "redirect" {
			http.Redirect(w, r, destination.URL, http.StatusFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"message":"ok"}`)
	}))
	defer server.Close()
	definition := source.Definition{Endpoint: server.URL, Rule: &rules.Plan{BaseURL: server.URL}}
	flow, _ := newImageVerification(server.Client(), definition, source.Query{Text: "test"})
	defer flow.Close()
	flow.primed = true
	if _, err := flow.Refresh(context.Background()); err == nil || destinationCalled {
		t.Fatal("cross-origin request followed")
	}
	mode = "ambiguous"
	accepted, err := flow.Submit(context.Background(), "test")
	if err == nil || accepted {
		t.Fatal("HTTP 200 without explicit success accepted")
	}
	if _, err := flow.Resume(context.Background()); !errors.Is(err, verification.ErrSession) {
		t.Fatal("unverified flow replayed")
	}
	flow.Close()
	if _, err := flow.Refresh(context.Background()); err == nil {
		t.Fatal("closed flow still active")
	}
}
