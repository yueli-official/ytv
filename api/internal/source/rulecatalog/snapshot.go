package rulecatalog

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/yueli-official/ytv/api/internal/source"
	"github.com/yueli-official/ytv/api/internal/source/rulecompiler"
)

//go:embed snapshot/kazumi.json
var kazumiSnapshotJSON []byte

type kazumiSnapshot struct {
	Revision string                     `json:"revision"`
	Index    []indexEntry               `json:"index"`
	Rules    map[string]json.RawMessage `json:"rules"`
}

// LoadKazumiSnapshot compiles the deployment-owned, commit-pinned catalog
// bundled with the YTV binary. Startup never depends on GitHub availability.
func LoadKazumiSnapshot(revision string) (LoadResult, error) {
	var snapshot kazumiSnapshot
	if err := json.Unmarshal(kazumiSnapshotJSON, &snapshot); err != nil {
		return LoadResult{}, fmt.Errorf("decode embedded Kazumi snapshot: %w", err)
	}
	if strings.ToLower(strings.TrimSpace(revision)) != snapshot.Revision {
		return LoadResult{}, fmt.Errorf("embedded Kazumi snapshot revision is %s, requested %s", snapshot.Revision, revision)
	}
	compiler, err := rulecompiler.New(map[string]rulecompiler.Adapter{"kazumi": rulecompiler.Kazumi{}})
	if err != nil {
		return LoadResult{}, err
	}
	loaded := LoadResult{
		Definitions: make([]source.Definition, 0, len(snapshot.Index)),
		Skipped:     make([]SkippedRule, 0),
	}
	seen := make(map[string]bool, len(snapshot.Index))
	for order, entry := range snapshot.Index {
		document, exists := snapshot.Rules[entry.Name]
		if !exists {
			loaded.Skipped = append(loaded.Skipped, SkippedRule{Name: entry.Name, Reason: "rule is missing from embedded snapshot"})
			continue
		}
		result := compileKazumiRule(compiler, entry, document, order)
		if result.skipped.Reason != "" {
			loaded.Skipped = append(loaded.Skipped, result.skipped)
			continue
		}
		if seen[result.definition.Key] {
			loaded.Skipped = append(loaded.Skipped, SkippedRule{Name: result.definition.Name, Reason: "duplicate source key"})
			continue
		}
		seen[result.definition.Key] = true
		loaded.Definitions = append(loaded.Definitions, result.definition)
	}
	sort.SliceStable(loaded.Definitions, func(left, right int) bool {
		return loaded.Definitions[left].Name < loaded.Definitions[right].Name
	})
	return loaded, nil
}
