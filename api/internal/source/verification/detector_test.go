package verification

import (
	"github.com/antchfx/htmlquery"
	"net/http"
	"strings"
	"testing"
)

func TestChallengeRequiresReliableEvidence(t *testing.T) {
	for _, tc := range []struct {
		name, body, header string
		want               Kind
	}{
		{"normal", `<p>captcha appears in a synopsis</p>`, "", ""},
		{"partial form", `<input name="verify">`, "", ""},
		{"image", `<title>系统安全验证</title><input name="verify">`, "", Image},
		{"browser", `<p>Blocked</p>`, "challenge", Browser},
		{"plain denied", `<p>403 Forbidden</p>`, "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			doc, err := htmlquery.Parse(strings.NewReader(tc.body))
			if err != nil {
				t.Fatal(err)
			}
			h := http.Header{}
			h.Set("cf-mitigated", tc.header)
			got := (Chain{Cloudflare{}, DOM{Kind: Image, All: []string{"//title[text()='系统安全验证']", "//input[@name='verify']"}}}).Detect(Response{Header: h, Document: doc})
			if tc.want == "" {
				if got != nil {
					t.Fatalf("false challenge: %v", got)
				}
				return
			}
			if got == nil || got.Kind != tc.want {
				t.Fatalf("got %#v want %s", got, tc.want)
			}
		})
	}
}
