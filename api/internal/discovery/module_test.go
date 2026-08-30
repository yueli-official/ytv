package discovery

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yueli-official/ytv/api/internal/source"
)

type fakeCatalog struct {
	definitions []source.Definition
	pages       map[string]source.Page
	errors      map[string]error
}

func (catalog *fakeCatalog) Definitions() []source.Definition { return catalog.definitions }
func (catalog *fakeCatalog) Browse(_ context.Context, key string, _ int) (source.Page, error) {
	return catalog.pages[key], catalog.errors[key]
}
func (catalog *fakeCatalog) Search(_ context.Context, key string, _ source.Query) (source.Page, error) {
	return catalog.pages[key], catalog.errors[key]
}
func (*fakeCatalog) ResolveTitle(context.Context, source.TitleLocator) (source.Title, error) {
	return source.Title{}, nil
}
func (*fakeCatalog) ResolveEpisode(context.Context, source.EpisodeLocator) (source.Title, source.Episode, error) {
	return source.Title{}, source.Episode{}, nil
}
func (*fakeCatalog) Probe(context.Context, string) (source.Health, error) {
	return source.Health{}, nil
}

func TestSearchReturnsStableGroupsAndPartialFailure(t *testing.T) {
	catalog := &fakeCatalog{
		definitions: []source.Definition{
			{Key: "b", Name: "B", Enabled: true, Priority: 20},
			{Key: "a", Name: "A", Enabled: true, Priority: 10},
			{Key: "failed", Name: "Failed", Enabled: true, Priority: 30},
		},
		pages: map[string]source.Page{
			"a": {Items: []source.Title{{Locator: source.TitleLocator{SourceKey: "a", ExternalID: "1"}, Name: "环 海之门", Year: "2026", Category: "科幻"}}},
			"b": {Items: []source.Title{{Locator: source.TitleLocator{SourceKey: "b", ExternalID: "9"}, Name: "环海之门", Year: "2026", Category: "科幻"}}},
		},
		errors: map[string]error{"failed": source.ErrUnavailable},
	}
	module, err := New(catalog, Options{Timeout: time.Second, MaxParallel: 2})
	if err != nil {
		t.Fatal(err)
	}
	page, err := module.Search(context.Background(), Query{Text: "  环海之门  "})
	if err != nil {
		t.Fatal(err)
	}
	if page.Query.Text != "环海之门" || page.Succeeded != 2 || page.Failed != 1 || page.Complete {
		t.Fatalf("page = %#v", page)
	}
	if len(page.Groups) != 1 || len(page.Groups[0].Candidates) != 2 {
		t.Fatalf("groups = %#v", page.Groups)
	}
	if page.Groups[0].Candidates[0].Title.Locator.SourceKey != "a" {
		t.Fatalf("candidate order = %#v", page.Groups[0].Candidates)
	}
}

func TestSearchFailsOnlyWhenEverySourceFails(t *testing.T) {
	catalog := &fakeCatalog{
		definitions: []source.Definition{{Key: "a", Name: "A", Enabled: true}},
		errors:      map[string]error{"a": errors.New("down")},
		pages:       map[string]source.Page{},
	}
	module, _ := New(catalog, Options{Timeout: time.Second})
	page, err := module.Search(context.Background(), Query{Text: "test"})
	if !errors.Is(err, ErrAllSourcesFailed) || page.Failed != 1 {
		t.Fatalf("page/error = %#v / %v", page, err)
	}
}

func TestSearchRejectsBlankAndOversizedQueries(t *testing.T) {
	module, _ := New(&fakeCatalog{definitions: []source.Definition{{Key: "a", Enabled: true}}}, Options{})
	if _, err := module.Search(context.Background(), Query{}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("blank error = %v", err)
	}
}

func TestBrowseUsesEnabledSourcesWithoutInventingAQuery(t *testing.T) {
	catalog := &fakeCatalog{
		definitions: []source.Definition{{Key: "a", Name: "A", Enabled: true, HomeVisible: true, Priority: 10}},
		pages: map[string]source.Page{"a": {Items: []source.Title{{
			Locator: source.TitleLocator{SourceKey: "a", ExternalID: "1"}, Name: "来源更新", Year: "2026",
		}}}},
		errors: map[string]error{},
	}
	module, _ := New(catalog, Options{})
	page, err := module.Browse(context.Background(), 1)
	if err != nil || page.Query.Text != "" || len(page.Groups) != 1 || page.Groups[0].Title != "来源更新" {
		t.Fatalf("browse = %#v, %v", page, err)
	}
}

func TestSearchUsesOnlyRequestedEnabledSources(t *testing.T) {
	catalog := &fakeCatalog{
		definitions: []source.Definition{
			{Key: "a", Name: "A", Enabled: true, Priority: 10},
			{Key: "b", Name: "B", Enabled: true, Priority: 20},
		},
		pages: map[string]source.Page{
			"a": {Items: []source.Title{{Locator: source.TitleLocator{SourceKey: "a", ExternalID: "1"}, Name: "A 结果"}}},
			"b": {Items: []source.Title{{Locator: source.TitleLocator{SourceKey: "b", ExternalID: "2"}, Name: "B 结果"}}},
		},
		errors: map[string]error{},
	}
	module, _ := New(catalog, Options{})
	page, err := module.Search(context.Background(), Query{Text: "结果", SourceKeys: []string{"b"}})
	if err != nil || page.Attempted != 1 || len(page.Sources) != 1 || page.Sources[0].Key != "b" {
		t.Fatalf("selected source page = %#v, %v", page, err)
	}
}

func TestBrowseUsesOnlySourcesVisibleOnHome(t *testing.T) {
	catalog := &fakeCatalog{
		definitions: []source.Definition{
			{Key: "home", Name: "Home", Enabled: true, HomeVisible: true, Priority: 10},
			{Key: "search-only", Name: "Search", Enabled: true, HomeVisible: false, Priority: 20},
		},
		pages: map[string]source.Page{
			"home":        {Items: []source.Title{{Locator: source.TitleLocator{SourceKey: "home", ExternalID: "1"}, Name: "首页结果"}}},
			"search-only": {Items: []source.Title{{Locator: source.TitleLocator{SourceKey: "search-only", ExternalID: "2"}, Name: "搜索结果"}}},
		},
		errors: map[string]error{},
	}
	module, _ := New(catalog, Options{})
	page, err := module.Browse(context.Background(), 1)
	if err != nil || page.Attempted != 1 || len(page.Sources) != 1 || page.Sources[0].Key != "home" {
		t.Fatalf("home source page = %#v, %v", page, err)
	}
}
