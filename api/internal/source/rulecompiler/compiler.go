// Package rulecompiler converts external rule formats into YTV's format-neutral plan.
package rulecompiler

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/yueli-official/ytv/api/internal/source/rules"
)

type Severity string

const (
	SeverityNotice      Severity = "notice"
	SeverityWarning     Severity = "warning"
	SeverityUnsupported Severity = "unsupported"
)

type Diagnostic struct {
	Severity Severity
	Code     string
	Message  string
}

type Result struct {
	Plan        rules.Plan
	Diagnostics []Diagnostic
}

func (result Result) Supported() bool {
	for _, diagnostic := range result.Diagnostics {
		if diagnostic.Severity == SeverityUnsupported {
			return false
		}
	}
	return true
}

type Adapter interface {
	Compile(json.RawMessage) (Result, error)
}

type Compiler struct {
	adapters map[string]Adapter
}

func New(adapters map[string]Adapter) (*Compiler, error) {
	if len(adapters) == 0 {
		return nil, fmt.Errorf("rule compiler requires at least one format adapter")
	}
	registered := make(map[string]Adapter, len(adapters))
	for format, adapter := range adapters {
		format = strings.ToLower(strings.TrimSpace(format))
		if format == "" || adapter == nil {
			return nil, fmt.Errorf("rule compiler format and adapter are required")
		}
		if _, exists := registered[format]; exists {
			return nil, fmt.Errorf("rule compiler format %q is duplicated", format)
		}
		registered[format] = adapter
	}
	return &Compiler{adapters: registered}, nil
}

func (compiler *Compiler) Compile(format string, document json.RawMessage) (Result, error) {
	format = strings.ToLower(strings.TrimSpace(format))
	adapter, exists := compiler.adapters[format]
	if !exists {
		return Result{}, fmt.Errorf("rule format %q is not registered", format)
	}
	result, err := adapter.Compile(document)
	if err != nil {
		return Result{}, fmt.Errorf("compile %s rule: %w", format, err)
	}
	result.Plan = result.Plan.Normalize()
	return result, nil
}

func (compiler *Compiler) Formats() []string {
	formats := make([]string, 0, len(compiler.adapters))
	for format := range compiler.adapters {
		formats = append(formats, format)
	}
	sort.Strings(formats)
	return formats
}
