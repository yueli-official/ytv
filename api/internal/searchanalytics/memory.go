package searchanalytics

import (
	"context"
	"sort"
	"sync"
	"time"
)

type MemoryStore struct {
	mu     sync.Mutex
	events map[string]Event
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{events: map[string]Event{}}
}

func (store *MemoryStore) Record(_ context.Context, event Event) (bool, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	if _, exists := store.events[event.ID]; exists {
		return true, nil
	}
	store.events[event.ID] = event
	return false, nil
}

func (store *MemoryStore) Overview(_ context.Context, since time.Time, limit int) (Overview, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	type aggregate struct {
		QueryStat
		ResultTotal   int64
		DurationTotal int64
	}
	queries := map[string]*aggregate{}
	days := map[string]*DayStat{}
	var result Overview
	for _, event := range store.events {
		if event.OccurredAt.Before(since) {
			continue
		}
		result.Summary.Searches++
		result.Summary.AverageDurationMS += event.Duration.Milliseconds()
		if event.ResultCount == 0 {
			result.Summary.ZeroResultSearches++
		}
		if event.FailedSourceCount > 0 {
			result.Summary.IncompleteSearches++
		}
		item := queries[event.Query]
		if item == nil {
			item = &aggregate{QueryStat: QueryStat{Query: event.Query}}
			queries[event.Query] = item
		}
		item.Searches++
		item.ResultTotal += int64(event.ResultCount)
		item.DurationTotal += event.Duration.Milliseconds()
		if event.ResultCount == 0 {
			item.ZeroResultSearches++
		}
		if event.OccurredAt.After(item.LastSearchedAt) {
			item.LastSearchedAt = event.OccurredAt
		}
		day := event.OccurredAt.In(time.FixedZone("CST", 8*60*60)).Format("2006-01-02")
		point := days[day]
		if point == nil {
			point = &DayStat{Day: day}
			days[day] = point
		}
		point.Searches++
		if event.ResultCount == 0 {
			point.ZeroResultSearches++
		}
	}
	if result.Summary.Searches > 0 {
		result.Summary.AverageDurationMS /= result.Summary.Searches
	}
	for _, item := range queries {
		item.AverageResultCount = float64(item.ResultTotal) / float64(item.Searches)
		item.AverageDurationMS = item.DurationTotal / item.Searches
		result.TopQueries = append(result.TopQueries, item.QueryStat)
	}
	sort.Slice(result.TopQueries, func(i, j int) bool {
		if result.TopQueries[i].Searches != result.TopQueries[j].Searches {
			return result.TopQueries[i].Searches > result.TopQueries[j].Searches
		}
		return result.TopQueries[i].LastSearchedAt.After(result.TopQueries[j].LastSearchedAt)
	})
	if len(result.TopQueries) > limit {
		result.TopQueries = result.TopQueries[:limit]
	}
	for _, point := range days {
		result.Series = append(result.Series, *point)
	}
	sort.Slice(result.Series, func(i, j int) bool { return result.Series[i].Day < result.Series[j].Day })
	return result, nil
}
