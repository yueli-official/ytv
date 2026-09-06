package sourcecontrol

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/yueli-official/ytv/api/internal/source"
)

var ErrInvalid = errors.New("invalid sourcecontrol input")

var ErrRevisionConflict = errors.New("source setting revision conflict")
var ErrAlreadyInstalled = errors.New("source is already installed")
var ErrNotInstalled = errors.New("source is not installed")

type Setting struct {
	SourceKey   string
	DisplayName string
	Installed   bool
	Enabled     bool
	HomeVisible bool
	Priority    int
	Revision    int64
	UpdatedAt   time.Time
}

type SourceView struct {
	Key            string
	Name           string
	Kind           source.Kind
	Catalog        string
	CatalogVersion string
	Enabled        bool
	HomeVisible    bool
	Priority       int
	Revision       int64
	UpdatedAt      time.Time
	Compatibility  []source.CompatibilityNotice
}

type AvailableSource struct {
	Key            string
	Name           string
	Kind           source.Kind
	Catalog        string
	CatalogVersion string
	Revision       int64
	Compatibility  []source.CompatibilityNotice
}

type ProbeResult struct {
	Source SourceView
	Health source.Health
}

type Store interface {
	LoadAll(context.Context) ([]Setting, error)
	Save(context.Context, Setting, int64) (Setting, error)
}

type Service struct {
	mu       sync.Mutex
	registry *source.Registry
	store    Store
	settings map[string]Setting
	clock    func() time.Time
}

func New(ctx context.Context, registry *source.Registry, store Store, clock func() time.Time) (*Service, error) {
	if registry == nil || store == nil {
		return nil, errors.New("ytv/sourcecontrol: registry and store are required")
	}
	if clock == nil {
		clock = time.Now
	}
	service := &Service{registry: registry, store: store, settings: map[string]Setting{}, clock: clock}
	known := map[string]bool{}
	for _, definition := range registry.Definitions() {
		known[definition.Key] = true
		service.settings[definition.Key] = Setting{
			SourceKey: definition.Key, DisplayName: definition.Name, Installed: !definition.InitiallyUninstalled,
			Enabled: definition.Enabled, HomeVisible: definition.HomeVisible, Priority: definition.Priority,
		}
	}
	stored, err := store.LoadAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("load source settings: %w", err)
	}
	for _, setting := range stored {
		if !known[setting.SourceKey] {
			continue
		}
		if strings.TrimSpace(setting.DisplayName) == "" {
			setting.DisplayName = service.settings[setting.SourceKey].DisplayName
		}
		if err := service.applySetting(setting); err != nil {
			return nil, fmt.Errorf("apply source setting %s: %w", setting.SourceKey, err)
		}
		service.settings[setting.SourceKey] = setting
	}
	return service, nil
}

func (service *Service) List() []SourceView {
	service.mu.Lock()
	defer service.mu.Unlock()
	definitions := service.registry.Definitions()
	result := make([]SourceView, 0, len(definitions))
	for _, definition := range definitions {
		setting := service.settings[definition.Key]
		if !setting.Installed {
			continue
		}
		result = append(result, SourceView{
			Key: definition.Key, Name: setting.DisplayName, Kind: definition.Kind,
			Catalog: definition.Catalog, CatalogVersion: definition.CatalogVersion,
			Enabled: setting.Enabled, HomeVisible: setting.HomeVisible, Priority: setting.Priority,
			Revision: setting.Revision, UpdatedAt: setting.UpdatedAt,
			Compatibility: definition.Compatibility,
		})
	}
	return result
}

func (service *Service) Available() []AvailableSource {
	service.mu.Lock()
	defer service.mu.Unlock()
	result := make([]AvailableSource, 0)
	for _, definition := range service.registry.Definitions() {
		setting := service.settings[definition.Key]
		if setting.Installed {
			continue
		}
		result = append(result, AvailableSource{Key: definition.Key, Name: setting.DisplayName, Kind: definition.Kind, Catalog: definition.Catalog, CatalogVersion: definition.CatalogVersion, Revision: setting.Revision, Compatibility: definition.Compatibility})
	}
	return result
}

func (service *Service) Update(ctx context.Context, sourceKey string, displayName string, enabled bool, homeVisible bool, priority int, expectedRevision int64) (SourceView, error) {
	sourceKey = strings.TrimSpace(sourceKey)
	displayName = strings.TrimSpace(displayName)
	if sourceKey == "" || displayName == "" || len([]rune(displayName)) > 80 || priority < 0 || expectedRevision < 0 {
		return SourceView{}, fmt.Errorf("%w: source setting input is invalid", ErrInvalid)
	}
	if !enabled {
		homeVisible = false
	}
	service.mu.Lock()
	defer service.mu.Unlock()
	current, exists := service.settings[sourceKey]
	if !exists || !current.Installed {
		return SourceView{}, fmt.Errorf("%w: %s", source.ErrNotFound, sourceKey)
	}
	if current.Revision != expectedRevision {
		return SourceView{}, ErrRevisionConflict
	}
	if current.DisplayName == displayName && current.Enabled == enabled && current.HomeVisible == homeVisible && current.Priority == priority {
		return service.view(sourceKey, current)
	}
	saved, err := service.store.Save(ctx, Setting{
		SourceKey: sourceKey, DisplayName: displayName, Installed: true,
		Enabled: enabled, HomeVisible: homeVisible, Priority: priority, UpdatedAt: service.clock().UTC(),
	}, expectedRevision)
	if err != nil {
		return SourceView{}, err
	}
	if err := service.applySetting(saved); err != nil {
		return SourceView{}, err
	}
	service.settings[sourceKey] = saved
	return service.view(sourceKey, saved)
}

func (service *Service) Install(ctx context.Context, sourceKey string, expectedRevision int64) (SourceView, error) {
	service.mu.Lock()
	defer service.mu.Unlock()
	sourceKey = strings.TrimSpace(sourceKey)
	current, exists := service.settings[sourceKey]
	if !exists {
		return SourceView{}, fmt.Errorf("%w: %s", source.ErrNotFound, sourceKey)
	}
	if current.Installed {
		return SourceView{}, ErrAlreadyInstalled
	}
	if current.Revision != expectedRevision {
		return SourceView{}, ErrRevisionConflict
	}
	current.Installed = true
	current.UpdatedAt = service.clock().UTC()
	saved, err := service.store.Save(ctx, current, expectedRevision)
	if err != nil {
		return SourceView{}, err
	}
	if err := service.applySetting(saved); err != nil {
		return SourceView{}, err
	}
	service.settings[sourceKey] = saved
	return service.view(sourceKey, saved)
}

func (service *Service) Remove(ctx context.Context, sourceKey string, expectedRevision int64) (AvailableSource, error) {
	service.mu.Lock()
	defer service.mu.Unlock()
	sourceKey = strings.TrimSpace(sourceKey)
	current, exists := service.settings[sourceKey]
	if !exists {
		return AvailableSource{}, fmt.Errorf("%w: %s", source.ErrNotFound, sourceKey)
	}
	if !current.Installed {
		return AvailableSource{}, ErrNotInstalled
	}
	if current.Revision != expectedRevision {
		return AvailableSource{}, ErrRevisionConflict
	}
	current.Installed = false
	current.UpdatedAt = service.clock().UTC()
	saved, err := service.store.Save(ctx, current, expectedRevision)
	if err != nil {
		return AvailableSource{}, err
	}
	if err := service.applySetting(saved); err != nil {
		return AvailableSource{}, err
	}
	service.settings[sourceKey] = saved
	for _, definition := range service.registry.Definitions() {
		if definition.Key == sourceKey {
			return AvailableSource{Key: sourceKey, Name: saved.DisplayName, Kind: definition.Kind, Catalog: definition.Catalog, CatalogVersion: definition.CatalogVersion, Revision: saved.Revision, Compatibility: definition.Compatibility}, nil
		}
	}
	return AvailableSource{}, fmt.Errorf("%w: %s", source.ErrNotFound, sourceKey)
}

func (service *Service) applySetting(setting Setting) error {
	return service.registry.UpdateDefinition(
		setting.SourceKey, setting.DisplayName,
		setting.Installed && setting.Enabled,
		setting.Installed && setting.Enabled && setting.HomeVisible,
		setting.Priority,
	)
}

func (service *Service) view(sourceKey string, setting Setting) (SourceView, error) {
	for _, definition := range service.registry.Definitions() {
		if definition.Key == sourceKey {
			return SourceView{
				Key: definition.Key, Name: setting.DisplayName, Kind: definition.Kind,
				Catalog: definition.Catalog, CatalogVersion: definition.CatalogVersion,
				Enabled: setting.Enabled, HomeVisible: setting.HomeVisible, Priority: setting.Priority,
				Revision: setting.Revision, UpdatedAt: setting.UpdatedAt,
				Compatibility: definition.Compatibility,
			}, nil
		}
	}
	return SourceView{}, fmt.Errorf("%w: %s", source.ErrNotFound, sourceKey)
}

func (service *Service) Probe(ctx context.Context, sourceKey string) (ProbeResult, error) {
	probeContext, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	startedAt := service.clock()
	health, err := service.registry.Probe(probeContext, sourceKey)
	if health.CheckedAt.IsZero() {
		health.CheckedAt = service.clock()
	}
	if health.Latency == 0 {
		health.Latency = service.clock().Sub(startedAt)
	}
	var current SourceView
	for _, item := range service.List() {
		if item.Key == strings.TrimSpace(sourceKey) {
			current = item
			break
		}
	}
	if current.Key == "" {
		return ProbeResult{}, fmt.Errorf("%w: %s", source.ErrNotFound, sourceKey)
	}
	if err != nil {
		health.Available = false
		return ProbeResult{Source: current, Health: health}, err
	}
	return ProbeResult{Source: current, Health: health}, nil
}

func (service *Service) ProbeAll(ctx context.Context) []ProbeResult {
	items := service.List()
	results := make([]ProbeResult, len(items))
	semaphore := make(chan struct{}, 8)
	var wait sync.WaitGroup
	for index, item := range items {
		wait.Add(1)
		go func() {
			defer wait.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			probeContext, cancel := context.WithTimeout(ctx, 4*time.Second)
			defer cancel()
			result, err := service.Probe(probeContext, item.Key)
			if result.Source.Key == "" {
				result.Source = item
			}
			if err != nil && strings.TrimSpace(result.Health.Detail) == "" {
				result.Health.Detail = "检测请求失败"
			}
			results[index] = result
		}()
	}
	wait.Wait()
	return results
}
