package sourcecontrol

import (
	"context"
	"errors"
	"github.com/yueli-official/ytv/api/internal/source"
	"testing"
)

func TestRecommendationPersistsAndPreservesOnOrdinaryUpdate(t *testing.T) {
	ctx := context.Background()
	registry, err := source.NewRegistry([]source.Definition{{Key: "anime", Name: "Anime", Kind: source.KindJSONCatalog, Endpoint: "https://example.com/api", Enabled: true}}, map[source.Kind]source.Adapter{source.KindJSONCatalog: availableAdapter{}})
	if err != nil {
		t.Fatal(err)
	}
	store := NewMemoryStore()
	service, err := New(ctx, registry, store, nil)
	if err != nil {
		t.Fatal(err)
	}
	if service.List()[0].Recommended {
		t.Fatal("unexpected default recommendation")
	}
	yes := true
	saved, err := service.Update(ctx, "anime", "Anime", true, false, 0, 0, PresentationUpdate{Recommended: &yes})
	if err != nil || !saved.Recommended {
		t.Fatalf("save: %#v %v", saved, err)
	}
	renamed, err := service.Update(ctx, "anime", "Renamed", true, false, 0, saved.Revision)
	if err != nil || !renamed.Recommended {
		t.Fatalf("preserve: %#v %v", renamed, err)
	}
	restored, err := New(ctx, registry, store, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !restored.List()[0].Recommended {
		t.Fatal("lost recommendation")
	}
	no := false
	cleared, err := restored.Update(ctx, "anime", "Renamed", true, false, 0, renamed.Revision, PresentationUpdate{Recommended: &no})
	if err != nil || cleared.Recommended {
		t.Fatalf("clear: %#v %v", cleared, err)
	}
	again, err := New(ctx, registry, store, nil)
	if err != nil {
		t.Fatal(err)
	}
	if again.List()[0].Recommended {
		t.Fatal("clearing not persisted")
	}
	if _, err = again.Update(ctx, "anime", "Renamed", true, false, 0, renamed.Revision, PresentationUpdate{Recommended: &yes}); !errors.Is(err, ErrRevisionConflict) {
		t.Fatal(err)
	}
}
