package discovery

import (
	"context"
	"fmt"
	"github.com/yueli-official/ytv/api/internal/source"
	"github.com/yueli-official/ytv/api/internal/source/verification"
	"testing"
)

func TestSearchPreservesBlockedSourcesAndAvailableResults(t *testing.T) {
	for _, partial := range []bool{false, true} {
		catalog := &fakeCatalog{definitions: []source.Definition{{Key: "blocked", Enabled: true}}, errors: map[string]error{"blocked": fmt.Errorf("search: %w", &verification.Required{Kind: verification.Image})}, pages: map[string]source.Page{}}
		if partial {
			catalog.definitions = append(catalog.definitions, source.Definition{Key: "good", Enabled: true})
			catalog.pages["good"] = source.Page{Items: []source.Title{{Name: "Available", Locator: source.TitleLocator{SourceKey: "good", ExternalID: "one"}}}}
		}
		module, _ := New(catalog, Options{})
		page, err := module.Search(context.Background(), Query{Text: "test"})
		if err != nil || page.Complete || page.Failed != 1 {
			t.Fatalf("page/error %#v %v", page, err)
		}
		found := false
		for _, s := range page.Sources {
			if s.Key == "blocked" {
				found = s.Outcome == OutcomeVerification && s.VerificationKind == "image"
			}
		}
		if !found {
			t.Fatalf("lost verification: %#v", page)
		}
		if partial && (len(page.Groups) != 1 || page.Succeeded != 1) {
			t.Fatalf("lost valid results: %#v", page)
		}
	}
}
