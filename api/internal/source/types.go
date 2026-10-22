package source

import (
	"context"
	"errors"
	"time"

	"github.com/yueli-official/ytv/api/internal/source/rules"
)

var (
	ErrInvalid     = errors.New("source input is invalid")
	ErrNotFound    = errors.New("source title is not found")
	ErrUnavailable = errors.New("source is unavailable")
)

type Kind string

const (
	KindJSONCatalog Kind = "json_catalog"
	KindHTMLCatalog Kind = "html_catalog"
	KindDeclarative Kind = "declarative"
	KindFixture     Kind = "fixture"
)

type Definition struct {
	Key                  string
	Name                 string
	Kind                 Kind
	Catalog              string
	CatalogVersion       string
	Endpoint             string
	PlaybackHosts        []string
	ImageHosts           []string
	InitiallyUninstalled bool
	Enabled              bool
	HomeVisible          bool
	Priority             int
	AllowHTTP            bool
	Rule                 *rules.Plan
	Compatibility        []CompatibilityNotice
}

type CompatibilityNotice struct {
	Level   string
	Code    string
	Message string
}

type Query struct {
	Text string
	Page int
}

type TitleLocator struct {
	SourceKey  string
	ExternalID string
}

type EpisodeLocator struct {
	Title      TitleLocator
	EpisodeKey string
}

type Episode struct {
	Key         string
	Title       string
	PlaybackURL string
	LineKey     string
	LineName    string
}

type Title struct {
	Locator      TitleLocator
	SourceName   string
	Name         string
	Year         string
	Category     string
	Area         string
	Language     string
	Score        string
	Remarks      string
	Summary      string
	ImageURL     string
	EpisodeCount int
	Episodes     []Episode
}

type Page struct {
	Items    []Title
	Total    int
	NextPage int
}

type Health struct {
	Available bool
	CheckedAt time.Time
	Latency   time.Duration
	Detail    string
}

type Adapter interface {
	Browse(context.Context, Definition, int) (Page, error)
	Search(context.Context, Definition, Query) (Page, error)
	ResolveTitle(context.Context, Definition, TitleLocator) (Title, error)
	ResolveEpisode(context.Context, Definition, EpisodeLocator) (Title, Episode, error)
	Probe(context.Context, Definition) (Health, error)
}

type Catalog interface {
	Definitions() []Definition
	Browse(context.Context, string, int) (Page, error)
	Search(context.Context, string, Query) (Page, error)
	ResolveTitle(context.Context, TitleLocator) (Title, error)
	ResolveEpisode(context.Context, EpisodeLocator) (Title, Episode, error)
	Probe(context.Context, string) (Health, error)
}
