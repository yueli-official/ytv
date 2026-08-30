package server

import (
	"database/sql"

	"github.com/gogf/gf/v2/net/ghttp"
	foundationauth "github.com/yueli-official/foundation/go/auth"
	"github.com/yueli-official/ytv/api/internal/httpapi"
	ytvruntime "github.com/yueli-official/ytv/api/internal/runtime"
)

type Dependencies struct {
	Controller *httpapi.Core
	Verifier   *foundationauth.Verifier
	Database   *sql.DB
}

func Configure(httpServer *ghttp.Server, dependencies Dependencies) {
	apiMiddleware := ytvruntime.MustAPIMiddleware().Handle
	httpServer.Use(ytvruntime.TraceRoute)
	httpServer.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(apiMiddleware)
		group.GET("/healthz", func(request *ghttp.Request) {
			request.Response.WriteJson(map[string]string{"status": "up"})
		})
		group.GET("/readyz", ytvruntime.ReadinessHandler(dependencies.Database))
	})
	if dependencies.Controller == nil || dependencies.Verifier == nil {
		return
	}
	httpServer.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(apiMiddleware, ytvruntime.OptionalAuth(dependencies.Verifier))
		group.Bind(dependencies.Controller.Public())
	})
	httpServer.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(apiMiddleware, ytvruntime.RequiredAuth(dependencies.Verifier))
		group.Bind(dependencies.Controller.Account())
		group.Bind(dependencies.Controller.Administrator())
	})
}
