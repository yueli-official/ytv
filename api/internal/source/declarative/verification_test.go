package declarative

import (
	"context"
	"errors"
	"fmt"
	"github.com/yueli-official/ytv/api/internal/source"
	"github.com/yueli-official/ytv/api/internal/source/rules"
	"github.com/yueli-official/ytv/api/internal/source/verification"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSearchDetectsVerificationBeforeParsingResults(t *testing.T) {
	for _, tc := range []struct {
		name, body, header string
		status             int
		want               bool
	}{
		{"image 200", `<title>系统安全验证</title><input name="verify">`, "", 200, true},
		{"browser 403", `Forbidden`, "challenge", 403, true},
		{"ordinary 403", `Forbidden`, "", 403, false},
		{"empty results", `<p>没有结果</p>`, "", 200, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("cf-mitigated", tc.header)
				w.WriteHeader(tc.status)
				fmt.Fprint(w, tc.body)
			}))
			defer server.Close()
			adapter, _ := New(server.Client())
			def := source.Definition{Key: "fixture", Endpoint: server.URL, Rule: &rules.Plan{BaseURL: server.URL, VerificationXPath: []string{"//title[text()='系统安全验证']", "//input[@name='verify']"}, Search: rules.SearchPlan{Request: rules.RequestPlan{Method: "GET", URL: "/search?q=@keyword"}, ListXPath: "//li", NameXPath: "/a/text()", LinkXPath: "/a/@href"}}}
			_, err := adapter.Search(context.Background(), def, source.Query{Text: "test"})
			if errors.Is(err, verification.ErrRequired) != tc.want {
				t.Fatalf("error %v; challenge expected %v", err, tc.want)
			}
			if tc.status == 403 && !tc.want && err == nil {
				t.Fatal("403 must remain an error")
			}
			if tc.status == 403 {
				_, err = adapter.fetchText(context.Background(), server.URL, server.URL, server.URL)
				if errors.Is(err, verification.ErrRequired) != tc.want {
					t.Fatalf("special extractor error %v", err)
				}
			}
		})
	}
}
