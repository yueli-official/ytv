package httpapi

import (
	"context"
	"errors"
	"github.com/yueli-official/foundation/go/auth"
	v1 "github.com/yueli-official/ytv/api/api/v1"
	"github.com/yueli-official/ytv/api/internal/apierr"
	"github.com/yueli-official/ytv/api/internal/source"
	"github.com/yueli-official/ytv/api/internal/sourcereview"
	"strings"
)

func (c *Core) reviewableSource(key string) bool {
	for _, v := range c.sources.List() {
		if v.Key == key && v.Enabled {
			return true
		}
	}
	return false
}
func (c *Account) GetSourceReview(ctx context.Context, r *v1.GetSourceReviewReq) (*v1.GetSourceReviewRes, error) {
	user, err := requiredReviewer(ctx)
	if err != nil {
		return nil, err
	}
	if !c.core.reviewableSource(r.SourceKey) {
		return nil, mapError(source.ErrNotFound)
	}
	if c.core.reviews == nil {
		return nil, errors.New("source reviews unavailable")
	}
	value, err := c.core.reviews.Get(ctx, user, r.SourceKey)
	if err != nil {
		return nil, mapError(err)
	}
	return &v1.GetSourceReviewRes{Review: reviewView(value)}, nil
}
func (c *Account) SaveSourceReview(ctx context.Context, r *v1.SaveSourceReviewReq) (*v1.SaveSourceReviewRes, error) {
	user, err := requiredReviewer(ctx)
	if err != nil {
		return nil, err
	}
	if !c.core.reviewableSource(r.SourceKey) {
		return nil, mapError(source.ErrNotFound)
	}
	if c.core.reviews == nil {
		return nil, errors.New("source reviews unavailable")
	}
	value, err := c.core.reviews.Save(ctx, user, r.SourceKey, r.Ads, r.Quality)
	if err != nil {
		return nil, mapError(err)
	}
	return &v1.SaveSourceReviewRes{Review: reviewView(value)}, nil
}
func reviewView(r sourcereview.Review) v1.SourceReviewView {
	return v1.SourceReviewView{Ads: r.Ads, Quality: r.Quality, UpdatedAt: r.UpdatedAt}
}
func reviewSummaryView(r sourcereview.Summary) v1.SourceReviewSummary {
	return v1.SourceReviewSummary{Total: r.Total, AdsFew: r.AdsFew, AdsMany: r.AdsMany, HD: r.HD, Normal: r.Normal, Blurry: r.Blurry}
}

// Reviews belong to the verified visitor or account, never a caller-supplied ID.
func requiredReviewer(ctx context.Context) (string, error) {
	principal, ok := auth.FromContext(ctx)
	if !ok {
		return "", apierr.Error(apierr.Unauthorized)
	}
	if strings.TrimSpace(principal.Subject) == "" {
		return "", apierr.Error(apierr.Forbidden)
	}
	switch principal.SubjectKind {
	case auth.SubjectUser:
		return principal.Subject, nil
	case auth.SubjectGuest:
		return "guest:" + principal.Subject, nil
	default:
		return "", apierr.Error(apierr.Forbidden)
	}
}
