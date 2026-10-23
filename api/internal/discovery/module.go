// Package discovery owns multi-source fan-out, partial success, stable ordering
// and reversible display grouping over source catalog Adapters.
package discovery

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/yueli-official/ytv/api/internal/source"
	"github.com/yueli-official/ytv/api/internal/source/verification"
)

var (
	ErrInvalid          = errors.New("discovery query is invalid")
	ErrAllSourcesFailed = errors.New("all discovery sources failed")
)

type Query struct {
	Text       string
	Page       int
	SourceKeys []string
	HomeOnly   bool
}

type Candidate struct {
	Title    source.Title
	Priority int
	Ordinal  int
}

type Group struct {
	Key        string
	Title      string
	Year       string
	Category   string
	Candidates []Candidate
}

type SourceOutcome string

const (
	OutcomeSuccess      SourceOutcome = "success"
	OutcomeTimeout      SourceOutcome = "timeout"
	OutcomeUnavailable  SourceOutcome = "unavailable"
	OutcomeVerification SourceOutcome = "verification_required"
)

type SourceResult struct {
	Key              string
	Name             string
	Outcome          SourceOutcome
	Duration         time.Duration
	ItemCount        int
	VerificationKind string
}

type Page struct {
	Query     Query
	Groups    []Group
	Sources   []SourceResult
	Attempted int
	Succeeded int
	Failed    int
	Complete  bool
}

type Options struct {
	Timeout     time.Duration
	MaxParallel int
	Clock       func() time.Time
}

type Module struct {
	catalog     source.Catalog
	timeout     time.Duration
	maxParallel int
	clock       func() time.Time
}

func New(catalog source.Catalog, options Options) (*Module, error) {
	if catalog == nil {
		return nil, fmt.Errorf("discovery catalog is required")
	}
	if options.Timeout <= 0 {
		options.Timeout = 6 * time.Second
	}
	if options.MaxParallel <= 0 {
		options.MaxParallel = 4
	}
	if options.Clock == nil {
		options.Clock = time.Now
	}
	return &Module{catalog: catalog, timeout: options.Timeout, maxParallel: options.MaxParallel, clock: options.Clock}, nil
}

func (module *Module) Search(ctx context.Context, query Query) (Page, error) {
	query.Text = strings.Join(strings.Fields(query.Text), " ")
	if query.Text == "" || utf8.RuneCountInString(query.Text) > 100 {
		return Page{}, fmt.Errorf("%w: query text must contain 1 to 100 characters", ErrInvalid)
	}
	return module.discover(ctx, query, true)
}

func (module *Module) Browse(ctx context.Context, page int) (Page, error) {
	return module.discover(ctx, Query{Page: page, HomeOnly: true}, false)
}

func (module *Module) discover(ctx context.Context, query Query, search bool) (Page, error) {
	if query.Page < 1 {
		query.Page = 1
	}
	selected := make(map[string]bool, len(query.SourceKeys))
	for _, key := range query.SourceKeys {
		key = strings.TrimSpace(key)
		if key != "" {
			selected[key] = true
		}
	}
	definitions := make([]source.Definition, 0)
	for _, definition := range module.catalog.Definitions() {
		if definition.Enabled && (!query.HomeOnly || definition.HomeVisible) && (len(selected) == 0 || selected[definition.Key]) {
			definitions = append(definitions, definition)
		}
	}
	if len(definitions) == 0 {
		return Page{}, fmt.Errorf("%w: no enabled sources", ErrAllSourcesFailed)
	}

	searchContext, cancel := context.WithTimeout(ctx, module.timeout)
	defer cancel()
	semaphore := make(chan struct{}, module.maxParallel)
	results := make(chan sourceSearchResult, len(definitions))
	for _, definition := range definitions {
		definition := definition
		go func() {
			started := module.clock()
			select {
			case semaphore <- struct{}{}:
				defer func() { <-semaphore }()
			case <-searchContext.Done():
				results <- sourceSearchResult{definition: definition, err: searchContext.Err(), duration: module.clock().Sub(started)}
				return
			}
			var page source.Page
			var err error
			if search {
				page, err = module.catalog.Search(searchContext, definition.Key, source.Query{Text: query.Text, Page: query.Page})
			} else {
				page, err = module.catalog.Browse(searchContext, definition.Key, query.Page)
			}
			results <- sourceSearchResult{definition: definition, page: page, err: err, duration: module.clock().Sub(started)}
		}()
	}

	page := Page{Query: query, Attempted: len(definitions), Sources: make([]SourceResult, 0, len(definitions))}
	grouped := make(map[string]*Group)
	seenLocators := make(map[source.TitleLocator]bool)
	verificationRequired := false
	for range definitions {
		result := <-results
		outcome := OutcomeSuccess
		verificationKind := ""
		if result.err != nil {
			page.Failed++
			outcome = OutcomeUnavailable
			var required *verification.Required
			if errors.As(result.err, &required) {
				outcome = OutcomeVerification
				verificationKind = string(required.Kind)
				verificationRequired = true
			} else if errors.Is(result.err, context.DeadlineExceeded) || errors.Is(result.err, context.Canceled) {
				outcome = OutcomeTimeout
			}
		} else {
			page.Succeeded++
			for ordinal, title := range result.page.Items {
				if seenLocators[title.Locator] {
					continue
				}
				seenLocators[title.Locator] = true
				key := displayGroupKey(title)
				group := grouped[key]
				if group == nil {
					group = &Group{Key: key, Title: title.Name, Year: title.Year, Category: title.Category}
					grouped[key] = group
				}
				group.Candidates = append(group.Candidates, Candidate{Title: title, Priority: result.definition.Priority, Ordinal: ordinal})
			}
		}
		page.Sources = append(page.Sources, SourceResult{
			Key: result.definition.Key, Name: result.definition.Name, Outcome: outcome,
			Duration: result.duration, ItemCount: len(result.page.Items),
			VerificationKind: verificationKind,
		})
	}
	page.Complete = page.Failed == 0
	if page.Succeeded == 0 && !verificationRequired {
		return page, ErrAllSourcesFailed
	}

	grouped = mergeIncompleteGroups(grouped)
	for _, group := range grouped {
		slices.SortFunc(group.Candidates, func(left, right Candidate) int {
			if left.Priority != right.Priority {
				return left.Priority - right.Priority
			}
			return strings.Compare(left.Title.Locator.SourceKey, right.Title.Locator.SourceKey)
		})
		page.Groups = append(page.Groups, *group)
	}
	slices.SortFunc(page.Groups, func(left, right Group) int {
		leftPriority := left.Candidates[0].Priority
		rightPriority := right.Candidates[0].Priority
		if leftPriority != rightPriority {
			return leftPriority - rightPriority
		}
		leftOrdinal := left.Candidates[0].Ordinal
		rightOrdinal := right.Candidates[0].Ordinal
		if leftOrdinal != rightOrdinal {
			return leftOrdinal - rightOrdinal
		}
		return strings.Compare(left.Key, right.Key)
	})
	slices.SortFunc(page.Sources, func(left, right SourceResult) int {
		return strings.Compare(left.Key, right.Key)
	})
	return page, nil
}

type sourceSearchResult struct {
	definition source.Definition
	page       source.Page
	err        error
	duration   time.Duration
}

func displayGroupKey(title source.Title) string {
	parts := []string{normalized(title.Name), normalized(title.Year), normalized(title.Category)}
	return strings.Join(parts, ":")
}

func normalized(value string) string {
	var builder strings.Builder
	for _, current := range strings.ToLower(strings.TrimSpace(value)) {
		if unicode.IsLetter(current) || unicode.IsNumber(current) {
			builder.WriteRune(current)
		}
	}
	return builder.String()
}

// Missing metadata is compatible only when it identifies one unambiguous group.
// Complete conflicting years/categories stay separate, regardless of arrival order.
func mergeIncompleteGroups(groups map[string]*Group) map[string]*Group {
	ordered := make([]*Group, 0, len(groups))
	for _, g := range groups {
		ordered = append(ordered, g)
	}
	specificity := func(g *Group) int {
		n := 0
		if normalized(g.Year) != "" {
			n++
		}
		if normalized(g.Category) != "" {
			n++
		}
		return n
	}
	slices.SortFunc(ordered, func(a, b *Group) int {
		if d := specificity(b) - specificity(a); d != 0 {
			return d
		}
		return strings.Compare(a.Key, b.Key)
	})
	result := make(map[string]*Group)
	for _, g := range ordered {
		var match *Group
		matches := 0
		for _, candidate := range result {
			if normalized(candidate.Title) != normalized(g.Title) || specificity(candidate) <= specificity(g) {
				continue
			}
			if y := normalized(g.Year); y != "" && y != normalized(candidate.Year) {
				continue
			}
			if c := normalized(g.Category); c != "" && c != normalized(candidate.Category) {
				continue
			}
			match = candidate
			matches++
		}
		if matches == 1 {
			match.Candidates = append(match.Candidates, g.Candidates...)
		} else {
			result[g.Key] = g
		}
	}
	return result
}
