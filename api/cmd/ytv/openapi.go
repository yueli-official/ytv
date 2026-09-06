package main

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	foundationopenapi "github.com/yueli-official/foundation/go/goframe/openapi"
	"github.com/yueli-official/ytv/api/internal/httpapi"
)

func exportOpenAPI(output string) error {
	s := g.Server("ytv-openapi")
	s.SetAddr("127.0.0.1:0")
	core := &httpapi.Core{}
	s.Group("/", func(group *ghttp.RouterGroup) {
		group.Bind(core.Public())
		group.Bind(core.Account())
		group.Bind(core.Administrator())
	})
	return foundationopenapi.Export(foundationopenapi.ExportConfig{Server: s, Output: output, Overwrite: true})
}
