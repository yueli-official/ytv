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
	TotalPages int
	HasNext    bool
	Query      Query
	Groups     []Group
	Sources    []SourceResult
	Attempted  int
	Succeeded  int
	Failed     int
	Complete   bool
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
			page.TotalPages = max(page.TotalPages, result.page.TotalPages)
			page.HasNext = page.HasNext || result.page.NextPage > query.Page
			for ordinal, title := range result.page.Items {
				if seenLocators[title.Locator] {
					continue
				}
				seenLocators[title.Locator] = true
				key := displayGroupKey(title)
				group := grouped[key]
				if group == nil {
					group = &Group{Key: key, Title: identityTitle(title.Name), Year: title.Year, Category: mediaCategory(title.Category)}
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
	page.TotalPages = max(1, page.TotalPages)
	if page.Succeeded == 0 && !verificationRequired {
		return page, ErrAllSourcesFailed
	}

	grouped = mergeConfirmedCredits(grouped)
	grouped = mergeIncompleteGroups(grouped)
	grouped = mergeSameTitles(grouped)
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
		if query.Text != "" {
			l, r := normalized(left.Title) == normalized(identityTitle(query.Text)), normalized(right.Title) == normalized(identityTitle(query.Text))
			if l != r {
				if l {
					return -1
				}
				return 1
			}
		}
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
	parts := []string{normalized(identityTitle(title.Name)), normalized(title.Year), normalized(mediaCategory(title.Category))}
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

// Language editions retain their individual locators inside the film group.
func identityTitle(name string) string {
	name = strings.TrimSpace(name)
	for _, suffix := range []string{"(国语)", "（国语）", "(粤语)", "（粤语）", "(国)", "（国）", "(粤)", "（粤）", "国语版", "粤语版", "国语", "粤语"} {
		if strings.HasSuffix(name, suffix) {
			return strings.TrimSpace(strings.TrimSuffix(name, suffix))
		}
	}
	return name
}
func mediaCategory(category string) string {
	category = strings.TrimSpace(category)
	switch {
	case strings.Contains(category, "动漫"), strings.Contains(category, "动画"):
		return "动漫"
	case strings.Contains(category, "短剧"):
		return "短剧"
	case strings.Contains(category, "综艺"):
		return "综艺"
	case strings.HasSuffix(category, "剧"), category == "电视剧":
		return "电视剧"
	case strings.HasSuffix(category, "片"), category == "电影":
		return "电影"
	default:
		return category
	}
}
func matchingCredits(a, b source.Title) bool {
	names := func(raw string) map[string]bool {
		result := map[string]bool{}
		for _, name := range strings.FieldsFunc(raw, func(r rune) bool { return strings.ContainsRune(",，/、;；|", r) }) {
			name = normalized(name)
			if name != "" && name != "未知" && name != "内详" {
				result[name] = true
			}
		}
		return result
	}
	overlap := func(a, b string) int {
		n := 0
		right := names(b)
		for name := range names(a) {
			if right[name] {
				n++
			}
		}
		return n
	}
	return overlap(a.Directors, b.Directors) > 0 && overlap(a.Actors, b.Actors) >= 2
}

// A contradictory year alone is not enough to merge remakes. Require matching
// directors and at least two cast members for the same movie title.
func mergeConfirmedCredits(groups map[string]*Group) map[string]*Group {
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	for i, k := range keys {
		a := groups[k]
		if a == nil || a.Category != "电影" {
			continue
		}
		for _, other := range keys[i+1:] {
			b := groups[other]
			if b == nil || b.Category != "电影" || normalized(a.Title) != normalized(b.Title) {
				continue
			}
			matched := false
			for _, ca := range a.Candidates {
				for _, cb := range b.Candidates {
					if matchingCredits(ca.Title, cb.Title) {
						matched = true
					}
				}
			}
			if matched {
				a.Candidates = append(a.Candidates, b.Candidates...)
				delete(groups, other)
			}
		}
	}
	return groups
}

// Display identity follows the viewer-facing title. Missing descriptions/posters
// remain candidates of the same card instead of creating metadata-only duplicates.
func mergeSameTitles(groups map[string]*Group) map[string]*Group {
	keys := make([]string, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	result := make(map[string]*Group)
	for _, key := range keys {
		group := groups[key]
		identity := normalized(group.Title)
		if target := result[identity]; target != nil {
			target.Candidates = append(target.Candidates, group.Candidates...)
			if target.Year == "" {
				target.Year = group.Year
			}
			if target.Category == "" {
				target.Category = group.Category
			}
		} else {
			group.Key = identity
			result[identity] = group
		}
	}
	return result
}
