package sourceconfig

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/yueli-official/ytv/api/internal/source"
	"github.com/yueli-official/ytv/api/internal/source/rulecompiler"
)

type definition struct {
	Key           string          `json:"key"`
	Name          string          `json:"name"`
	Kind          source.Kind     `json:"kind"`
	Endpoint      string          `json:"endpoint"`
	PlaybackHosts []string        `json:"playbackHosts"`
	ImageHosts    []string        `json:"imageHosts"`
	Installed     *bool           `json:"installed"`
	Enabled       bool            `json:"enabled"`
	HomeVisible   *bool           `json:"homeVisible"`
	Priority      int             `json:"priority"`
	RuleFormat    string          `json:"ruleFormat"`
	Rule          json.RawMessage `json:"rule"`
}

func Load(raw string, includeFixture bool) ([]source.Definition, error) {
	compiler, err := rulecompiler.New(map[string]rulecompiler.Adapter{"kazumi": rulecompiler.Kazumi{}})
	if err != nil {
		return nil, err
	}
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
			definition := source.Definition{
				Key: value.Key, Name: value.Name, Kind: value.Kind, Endpoint: value.Endpoint,
				PlaybackHosts: value.PlaybackHosts, ImageHosts: value.ImageHosts,
				InitiallyUninstalled: value.Installed != nil && !*value.Installed,
				Enabled:              value.Enabled, HomeVisible: homeVisible, Priority: value.Priority,
			}
			if value.Kind == source.KindDeclarative {
				if strings.TrimSpace(value.RuleFormat) == "" || len(value.Rule) == 0 {
					return nil, fmt.Errorf("source %q declarative rule format and document are required", value.Key)
				}
				compiled, compileErr := compiler.Compile(value.RuleFormat, value.Rule)
				if compileErr != nil {
					return nil, fmt.Errorf("source %q: %w", value.Key, compileErr)
				}
				definition.Rule = &compiled.Plan
				for _, diagnostic := range compiled.Diagnostics {
					definition.Compatibility = append(definition.Compatibility, source.CompatibilityNotice{
						Level: string(diagnostic.Severity), Code: diagnostic.Code, Message: diagnostic.Message,
					})
				}
				if strings.TrimSpace(definition.Endpoint) == "" {
					definition.Endpoint = compiled.Plan.BaseURL
				}
			}
			definitions = append(definitions, definition)
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
