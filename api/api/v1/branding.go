package v1

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/yueli-official/foundation/go/siteprofile"
)

type GetSiteBrandingReq struct {
	g.Meta `path:"/api/v1/site/branding" method:"get" tags:"Public" summary:"Get public site branding"`
}

type GetSiteBrandingRes struct {
	siteprofile.BrandImageURLs
}

type PatchSiteBrandingReq struct {
	g.Meta `path:"/api/v1/site/branding" method:"patch" tags:"Administration" summary:"Update site branding"`
	siteprofile.BrandImagePatch
}

type PatchSiteBrandingRes struct {
	siteprofile.BrandImageURLs
}
