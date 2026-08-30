package source

import (
	"context"
	"testing"
)

type fakeAdapter struct{}

func (fakeAdapter) Browse(context.Context, Definition, int) (Page, error)   { return Page{}, nil }
func (fakeAdapter) Search(context.Context, Definition, Query) (Page, error) { return Page{}, nil }
func (fakeAdapter) ResolveTitle(context.Context, Definition, TitleLocator) (Title, error) {
	return Title{}, nil
}
func (fakeAdapter) ResolveEpisode(context.Context, Definition, EpisodeLocator) (Title, Episode, error) {
	return Title{}, Episode{}, nil
}
func (fakeAdapter) Probe(context.Context, Definition) (Health, error) {
	return Health{Available: true}, nil
}

func TestRegistryOrdersDefinitionsAndRejectsUnknownSources(t *testing.T) {
	registry, err := NewRegistry([]Definition{
		{Key: "later", Name: "Later", Kind: KindJSONCatalog, Endpoint: "https://later.example/api", Enabled: true, Priority: 20},
		{Key: "first", Name: "First", Kind: KindJSONCatalog, Endpoint: "https://first.example/api", Enabled: true, Priority: 10},
	}, map[Kind]Adapter{KindJSONCatalog: fakeAdapter{}})
	if err != nil {
		t.Fatal(err)
	}
	definitions := registry.Definitions()
	if definitions[0].Key != "first" || definitions[1].Key != "later" {
		t.Fatalf("definitions = %#v", definitions)
	}
	if _, err := registry.Search(context.Background(), "missing", Query{Text: "query"}); err == nil {
		t.Fatal("missing source unexpectedly searched")
	}
}

func TestRegistrySettingReordersAndDisabledProbeStillWorks(t *testing.T) {
	registry, err := NewRegistry([]Definition{
		{Key: "a", Name: "A", Kind: KindJSONCatalog, Endpoint: "https://a.example/api", Enabled: true, Priority: 10},
		{Key: "b", Name: "B", Kind: KindJSONCatalog, Endpoint: "https://b.example/api", Enabled: true, Priority: 20},
	}, map[Kind]Adapter{KindJSONCatalog: fakeAdapter{}})
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.UpdateSetting("b", false, true, 5); err != nil {
		t.Fatal(err)
	}
	definitions := registry.Definitions()
	if definitions[0].Key != "b" || definitions[0].Enabled || !definitions[0].HomeVisible {
		t.Fatalf("definitions = %#v", definitions)
	}
	if _, err := registry.Search(context.Background(), "b", Query{Text: "query"}); err == nil {
		t.Fatal("disabled source unexpectedly searched")
	}
	if health, err := registry.Probe(context.Background(), "b"); err != nil || !health.Available {
		t.Fatalf("disabled source probe = %#v, %v", health, err)
	}
}
