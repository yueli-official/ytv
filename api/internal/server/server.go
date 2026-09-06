package server

import (
	"database/sql"
	"net/http"

	"github.com/gogf/gf/v2/net/ghttp"
	foundationauth "github.com/yueli-official/foundation/go/auth"
	foundationhttp "github.com/yueli-official/foundation/go/goframe/http"
	"github.com/yueli-official/ytv/api/internal/apierr"
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
	for status, code := range map[int]string{http.StatusNotFound: apierr.CodeNotFound, http.StatusMethodNotAllowed: apierr.CodeMethodNotAllowed} {
		httpServer.BindStatusHandler(status, func(request *ghttp.Request) {
			request.Response.ClearBuffer()
			writer := foundationhttp.MustWriter(foundationhttp.WriterOptions{TraceHeader: ytvruntime.TraceHeader})
			_, _ = writer.WriteError(request, apierr.Public(code), request.Header.Get(ytvruntime.TraceHeader))
		})
	}
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
