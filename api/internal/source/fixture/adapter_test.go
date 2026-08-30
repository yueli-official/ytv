package fixture

import (
	"context"
	"testing"

	"github.com/yueli-official/ytv/api/internal/source"
)

func TestAdapterProvidesDeterministicSearchAndResolve(t *testing.T) {
	adapter := New(Titles())
	definition := source.Definition{Key: "fixture", Name: "Fixture"}
	page, err := adapter.Search(context.Background(), definition, source.Query{Text: "霜"})
	if err != nil || len(page.Items) != 1 || page.Items[0].Locator.SourceKey != "fixture" {
		t.Fatalf("page/error = %#v / %v", page, err)
	}
	resolved, err := adapter.ResolveTitle(context.Background(), definition, source.TitleLocator{SourceKey: "fixture", ExternalID: "frost-line"})
	if err != nil || resolved.Name != "霜线" {
		t.Fatalf("resolved/error = %#v / %v", resolved, err)
	}
}
