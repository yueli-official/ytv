package source

import (
	"context"
	"fmt"
	"net/url"
	"slices"
	"strings"
	"sync"
)

type Registry struct {
	mu          sync.RWMutex
	definitions map[string]Definition
	ordered     []Definition
	adapters    map[Kind]Adapter
}

func NewRegistry(definitions []Definition, adapters map[Kind]Adapter) (*Registry, error) {
	if len(definitions) == 0 {
		return nil, fmt.Errorf("%w: at least one source definition is required", ErrInvalid)
	}
	registry := &Registry{
		definitions: make(map[string]Definition, len(definitions)),
		ordered:     make([]Definition, 0, len(definitions)),
		adapters:    make(map[Kind]Adapter, len(adapters)),
	}
	for kind, adapter := range adapters {
		if adapter == nil {
			return nil, fmt.Errorf("%w: adapter %q is nil", ErrInvalid, kind)
		}
		registry.adapters[kind] = adapter
	}
	for _, definition := range definitions {
		definition.Key = strings.TrimSpace(definition.Key)
		definition.Name = strings.TrimSpace(definition.Name)
		definition.Endpoint = strings.TrimSpace(definition.Endpoint)
		if definition.Key == "" || definition.Name == "" || definition.Endpoint == "" {
			return nil, fmt.Errorf("%w: source key, name and endpoint are required", ErrInvalid)
		}
		if _, exists := registry.definitions[definition.Key]; exists {
			return nil, fmt.Errorf("%w: duplicate source key %q", ErrInvalid, definition.Key)
		}
		if _, exists := registry.adapters[definition.Kind]; !exists {
			return nil, fmt.Errorf("%w: source %q has no %q adapter", ErrInvalid, definition.Key, definition.Kind)
		}
		endpoint, err := url.Parse(definition.Endpoint)
		if err != nil || endpoint.Hostname() == "" || (endpoint.Scheme != "https" && !(definition.AllowHTTP && endpoint.Scheme == "http")) {
			return nil, fmt.Errorf("%w: source %q endpoint must be an allowed http(s) URL", ErrInvalid, definition.Key)
		}
		registry.definitions[definition.Key] = definition
		registry.ordered = append(registry.ordered, definition)
	}
	slices.SortFunc(registry.ordered, compareDefinitions)
	return registry, nil
}

func (registry *Registry) Definitions() []Definition {
	registry.mu.RLock()
	defer registry.mu.RUnlock()
	result := make([]Definition, len(registry.ordered))
	copy(result, registry.ordered)
	return result
}

func (registry *Registry) Search(ctx context.Context, key string, query Query) (Page, error) {
	definition, adapter, err := registry.resolve(key)
	if err != nil {
		return Page{}, err
	}
	return adapter.Search(ctx, definition, query)
}

func (registry *Registry) Browse(ctx context.Context, key string, page int) (Page, error) {
	definition, adapter, err := registry.resolve(key)
	if err != nil {
		return Page{}, err
	}
	return adapter.Browse(ctx, definition, page)
}

func (registry *Registry) ResolveTitle(ctx context.Context, locator TitleLocator) (Title, error) {
	definition, adapter, err := registry.resolve(locator.SourceKey)
	if err != nil {
		return Title{}, err
	}
	return adapter.ResolveTitle(ctx, definition, locator)
}

func (registry *Registry) ResolveEpisode(ctx context.Context, locator EpisodeLocator) (Title, Episode, error) {
	definition, adapter, err := registry.resolve(locator.Title.SourceKey)
	if err != nil {
		return Title{}, Episode{}, err
	}
	return adapter.ResolveEpisode(ctx, definition, locator)
}

func (registry *Registry) Probe(ctx context.Context, key string) (Health, error) {
	definition, adapter, err := registry.resolveConfigured(key)
	if err != nil {
		return Health{}, err
	}
	return adapter.Probe(ctx, definition)
}

func (registry *Registry) UpdateSetting(key string, enabled bool, homeVisible bool, priority int) error {
	registry.mu.RLock()
	definition, exists := registry.definitions[strings.TrimSpace(key)]
	registry.mu.RUnlock()
	if !exists {
		return fmt.Errorf("%w: source %q", ErrNotFound, key)
	}
	return registry.UpdateDefinition(key, definition.Name, enabled, homeVisible, priority)
}

func (registry *Registry) UpdateDefinition(key string, name string, enabled bool, homeVisible bool, priority int) error {
	if priority < 0 {
		return fmt.Errorf("%w: source priority must not be negative", ErrInvalid)
	}
	registry.mu.Lock()
	defer registry.mu.Unlock()
	key = strings.TrimSpace(key)
	name = strings.TrimSpace(name)
	if name == "" || len([]rune(name)) > 80 {
		return fmt.Errorf("%w: source name is required and must not exceed 80 characters", ErrInvalid)
	}
	definition, exists := registry.definitions[key]
	if !exists {
		return fmt.Errorf("%w: source %q", ErrNotFound, key)
	}
	definition.Name = name
	definition.Enabled = enabled
	definition.HomeVisible = homeVisible
	definition.Priority = priority
	registry.definitions[key] = definition
	for index := range registry.ordered {
		if registry.ordered[index].Key == key {
			registry.ordered[index] = definition
			break
		}
	}
	slices.SortFunc(registry.ordered, compareDefinitions)
	return nil
}

func (registry *Registry) resolve(key string) (Definition, Adapter, error) {
	registry.mu.RLock()
	defer registry.mu.RUnlock()
	definition, exists := registry.definitions[strings.TrimSpace(key)]
	if !exists || !definition.Enabled {
		return Definition{}, nil, fmt.Errorf("%w: source %q", ErrNotFound, key)
	}
	return definition, registry.adapters[definition.Kind], nil
}

func (registry *Registry) resolveConfigured(key string) (Definition, Adapter, error) {
	registry.mu.RLock()
	defer registry.mu.RUnlock()
	definition, exists := registry.definitions[strings.TrimSpace(key)]
	if !exists {
		return Definition{}, nil, fmt.Errorf("%w: source %q", ErrNotFound, key)
	}
	return definition, registry.adapters[definition.Kind], nil
}

func compareDefinitions(left, right Definition) int {
	if left.Priority != right.Priority {
		return left.Priority - right.Priority
	}
	return strings.Compare(left.Key, right.Key)
}
