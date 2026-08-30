package sourceconfig

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/yueli-official/ytv/api/internal/source"
)

type definition struct {
	Key           string      `json:"key"`
	Name          string      `json:"name"`
	Kind          source.Kind `json:"kind"`
	Endpoint      string      `json:"endpoint"`
	PlaybackHosts []string    `json:"playbackHosts"`
	ImageHosts    []string    `json:"imageHosts"`
	Installed     *bool       `json:"installed"`
	Enabled       bool        `json:"enabled"`
	HomeVisible   *bool       `json:"homeVisible"`
	Priority      int         `json:"priority"`
}

func Load(raw string, includeFixture bool) ([]source.Definition, error) {
	definitions := make([]source.Definition, 0)
	if strings.TrimSpace(raw) != "" {
		decoder := json.NewDecoder(io.LimitReader(strings.NewReader(raw), 64<<10))
		decoder.DisallowUnknownFields()
		var values []definition
		if err := decoder.Decode(&values); err != nil {
			return nil, fmt.Errorf("decode YTV source definitions: %w", err)
		}
		for _, value := range values {
			homeVisible := true
			if value.HomeVisible != nil {
				homeVisible = *value.HomeVisible
			}
			definitions = append(definitions, source.Definition{
				Key: value.Key, Name: value.Name, Kind: value.Kind, Endpoint: value.Endpoint,
				PlaybackHosts: value.PlaybackHosts, ImageHosts: value.ImageHosts,
				InitiallyUninstalled: value.Installed != nil && !*value.Installed,
				Enabled:              value.Enabled, HomeVisible: homeVisible, Priority: value.Priority,
			})
		}
	}
	if includeFixture {
		definitions = append(definitions, source.Definition{
			Key: "fixture", Name: "本地演示源", Kind: source.KindFixture,
			Endpoint: "https://fixture.invalid/catalog", PlaybackHosts: []string{"*.fixture.invalid"},
			Enabled: true, HomeVisible: true, Priority: 1000,
		})
	}
	if len(definitions) == 0 {
		return nil, fmt.Errorf("at least one deployment source definition is required")
	}
	return definitions, nil
}
