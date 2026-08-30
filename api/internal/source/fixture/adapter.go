// Package fixture provides a deterministic local Source Adapter for tests and
// Workspace acceptance. Production definitions must not use it.
package fixture

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/yueli-official/ytv/api/internal/source"
)

type Adapter struct {
	titles  map[string]source.Title
	ordered []source.Title
}

func New(titles []source.Title) *Adapter {
	byID := make(map[string]source.Title, len(titles))
	for _, title := range titles {
		byID[title.Locator.ExternalID] = title
	}
	ordered := append([]source.Title(nil), titles...)
	slices.SortFunc(ordered, func(left, right source.Title) int { return strings.Compare(left.Name, right.Name) })
	return &Adapter{titles: byID, ordered: ordered}
}

func (adapter *Adapter) Browse(_ context.Context, definition source.Definition, page int) (source.Page, error) {
	items := make([]source.Title, 0, len(adapter.ordered))
	for _, title := range adapter.ordered {
		value := title
		value.Locator.SourceKey = definition.Key
		value.SourceName = definition.Name
		items = append(items, value)
	}
	return source.Page{Items: items, Total: len(items)}, nil
}

func (adapter *Adapter) Search(_ context.Context, definition source.Definition, query source.Query) (source.Page, error) {
	needle := strings.ToLower(strings.TrimSpace(query.Text))
	items := make([]source.Title, 0)
	for _, title := range adapter.titles {
		if needle != "" && !strings.Contains(strings.ToLower(title.Name), needle) {
			continue
		}
		value := title
		value.Locator.SourceKey = definition.Key
		value.SourceName = definition.Name
		items = append(items, value)
	}
	return source.Page{Items: items, Total: len(items)}, nil
}

func (adapter *Adapter) ResolveTitle(_ context.Context, definition source.Definition, locator source.TitleLocator) (source.Title, error) {
	title, exists := adapter.titles[locator.ExternalID]
	if !exists {
		return source.Title{}, fmt.Errorf("%w: %s", source.ErrNotFound, locator.ExternalID)
	}
	title.Locator.SourceKey = definition.Key
	title.SourceName = definition.Name
	return title, nil
}

func (adapter *Adapter) ResolveEpisode(ctx context.Context, definition source.Definition, locator source.EpisodeLocator) (source.Title, source.Episode, error) {
	title, err := adapter.ResolveTitle(ctx, definition, locator.Title)
	if err != nil {
		return source.Title{}, source.Episode{}, err
	}
	for _, episode := range title.Episodes {
		if episode.Key == locator.EpisodeKey {
			return title, episode, nil
		}
	}
	return source.Title{}, source.Episode{}, fmt.Errorf("%w: episode %s", source.ErrNotFound, locator.EpisodeKey)
}

func (*Adapter) Probe(context.Context, source.Definition) (source.Health, error) {
	return source.Health{Available: true, CheckedAt: time.Now(), Detail: "deterministic fixture"}, nil
}
