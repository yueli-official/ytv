package httpapi

import (
	"context"
	v1 "github.com/yueli-official/ytv/api/api/v1"
	"sort"
)

// Rank source choices after discovery caching, so edits and withdrawals apply
// to the next search without changing movie relevance or exposing a score.
func (controller *Public) rankSearchSources(ctx context.Context, response *v1.SearchRes) error {
	sources, err := controller.GetPublicSources(ctx, &v1.GetPublicSourcesReq{})
	if err != nil {
		return err
	}
	type ranking struct {
		score       int64
		recommended bool
	}
	ranks := make(map[string]ranking, len(sources.Items))
	for _, item := range sources.Items {
		r := item.Reviews
		ranks[item.Key] = ranking{
			score:       int64(r.HD)*10 + int64(r.AdsFew)*8 - int64(r.AdsMany)*8 - int64(r.Blurry)*10,
			recommended: item.Recommended,
		}
	}
	for i := range response.Groups {
		candidates := response.Groups[i].Candidates
		sort.SliceStable(candidates, func(i, j int) bool {
			a, b := ranks[candidates[i].SourceKey], ranks[candidates[j].SourceKey]
			if a.score != b.score {
				return a.score > b.score
			}
			return a.recommended && !b.recommended
		})
	}
	return nil
}
