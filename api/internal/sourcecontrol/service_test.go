package sourcecontrol

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yueli-official/ytv/api/internal/source"
)

type availableAdapter struct{}

func (availableAdapter) Browse(context.Context, source.Definition, int) (source.Page, error) {
	return source.Page{}, nil
}
func (availableAdapter) Search(context.Context, source.Definition, source.Query) (source.Page, error) {
	return source.Page{}, nil
}
func (availableAdapter) ResolveTitle(context.Context, source.Definition, source.TitleLocator) (source.Title, error) {
	return source.Title{}, nil
}
func (availableAdapter) ResolveEpisode(context.Context, source.Definition, source.EpisodeLocator) (source.Title, source.Episode, error) {
	return source.Title{}, source.Episode{}, nil
}
func (availableAdapter) Probe(context.Context, source.Definition) (source.Health, error) {
	return source.Health{Available: true}, nil
}

func TestStoredSettingHydratesAndUpdateUsesRevision(t *testing.T) {
	registry, err := source.NewRegistry([]source.Definition{{
		Key: "catalog", Name: "Catalog", Kind: source.KindJSONCatalog,
		Endpoint: "https://catalog.example/api", Enabled: true, HomeVisible: true, Priority: 10,
	}}, map[source.Kind]source.Adapter{source.KindJSONCatalog: availableAdapter{}})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC)
	service, err := New(context.Background(), registry, NewMemoryStore(Setting{
		SourceKey: "catalog", DisplayName: "Catalog", Installed: true,
		Enabled: false, HomeVisible: false, Priority: 30, Revision: 2, UpdatedAt: now,
	}), func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	if got := service.List()[0]; got.Enabled || got.HomeVisible || got.Priority != 30 || got.Revision != 2 {
		t.Fatalf("hydrated source = %#v", got)
	}
	unchanged, err := service.Update(context.Background(), "catalog", "Catalog", false, false, 30, 2)
	if err != nil || unchanged.Revision != 2 {
		t.Fatalf("unchanged source = %#v, %v", unchanged, err)
	}
	updated, err := service.Update(context.Background(), "catalog", "Catalog", true, true, 5, 2)
	if err != nil || !updated.Enabled || !updated.HomeVisible || updated.Priority != 5 || updated.Revision != 3 {
		t.Fatalf("updated source = %#v, %v", updated, err)
	}
	if _, err := service.Update(context.Background(), "catalog", "Catalog", false, false, 9, 2); !errors.Is(err, ErrRevisionConflict) {
		t.Fatalf("stale update error = %v", err)
	}
}

func TestApprovedSourceLifecycleInstallsRenamesAndRemoves(t *testing.T) {
	registry, err := source.NewRegistry([]source.Definition{
		{Key: "one", Name: "One", Kind: source.KindJSONCatalog, Endpoint: "https://one.example/api", Enabled: true, Priority: 10},
		{Key: "two", Name: "Two", Kind: source.KindJSONCatalog, Endpoint: "https://two.example/api", Enabled: false, Priority: 20},
	}, map[source.Kind]source.Adapter{source.KindJSONCatalog: availableAdapter{}})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC)
	service, err := New(context.Background(), registry, NewMemoryStore(), func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	removed, err := service.Remove(context.Background(), "two", 0)
	if err != nil || removed.Key != "two" || removed.Revision != 1 || len(service.List()) != 1 || len(service.Available()) != 1 {
		t.Fatalf("removed/installed/available = %#v / %#v / %#v / %v", removed, service.List(), service.Available(), err)
	}
	installed, err := service.Install(context.Background(), "two", removed.Revision)
	if err != nil || installed.Enabled || installed.Revision != 2 || len(service.List()) != 2 || len(service.Available()) != 0 {
		t.Fatalf("installed/list/available = %#v / %#v / %#v / %v", installed, service.List(), service.Available(), err)
	}
	updated, err := service.Update(context.Background(), "two", "Renamed Two", true, false, 5, installed.Revision)
	if err != nil || updated.Name != "Renamed Two" || updated.Revision != 3 {
		t.Fatalf("updated = %#v / %v", updated, err)
	}
}

func TestApprovedDefinitionCanStartAvailableWithoutJoiningRuntime(t *testing.T) {
	registry, err := source.NewRegistry([]source.Definition{
		{Key: "installed", Name: "Installed", Kind: source.KindJSONCatalog, Endpoint: "https://installed.example/api", Enabled: true},
		{Key: "candidate", Name: "Candidate", Kind: source.KindJSONCatalog, Endpoint: "https://candidate.example/api", InitiallyUninstalled: true},
	}, map[source.Kind]source.Adapter{source.KindJSONCatalog: availableAdapter{}})
	if err != nil {
		t.Fatal(err)
	}
	service, err := New(context.Background(), registry, NewMemoryStore(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := service.List(); len(got) != 1 || got[0].Key != "installed" {
		t.Fatalf("installed = %#v", got)
	}
	if got := service.Available(); len(got) != 1 || got[0].Key != "candidate" || got[0].Revision != 0 {
		t.Fatalf("available = %#v", got)
	}
}
