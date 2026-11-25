package httpapi

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	v1 "github.com/yueli-official/ytv/api/api/v1"
	"github.com/yueli-official/ytv/api/internal/siteprofile"
)

func (controller *Public) GetSiteBranding(ctx context.Context, _ *v1.GetSiteBrandingReq) (*v1.GetSiteBrandingRes, error) {
	profile, err := controller.core.profile.Get(ctx)
	if err != nil {
		return nil, mapError(err)
	}
	header := g.RequestFromCtx(ctx).Response.Header()
	header.Set("Access-Control-Allow-Origin", "*")
	header.Set("Cache-Control", "public, no-cache")
	return &v1.GetSiteBrandingRes{BrandImageURLs: profile.BrandImageURLs.WithDefaults(siteprofile.BrandingDefaults)}, nil
}

func (controller *Administrator) PatchSiteBranding(ctx context.Context, request *v1.PatchSiteBrandingReq) (*v1.PatchSiteBrandingRes, error) {
	if _, err := controller.core.requireAdministrator(ctx); err != nil {
		return nil, err
	}
	profile, err := controller.core.profile.PatchBranding(ctx, request.BrandImagePatch)
	if err != nil {
		return nil, mapError(err)
	}
	g.RequestFromCtx(ctx).Response.Header().Set("Cache-Control", "private, no-store")
	return &v1.PatchSiteBrandingRes{BrandImageURLs: profile.BrandImageURLs.WithDefaults(siteprofile.BrandingDefaults)}, nil
}
