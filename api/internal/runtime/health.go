package runtime

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gogf/gf/v2/net/ghttp"
	goframehealth "github.com/yueli-official/foundation/go/goframe/health"
	"github.com/yueli-official/foundation/go/health"
	"github.com/yueli-official/foundation/go/problem"
)

var notReady = problem.MustDescriptor(
	problem.MustKind("common.not_ready", http.StatusServiceUnavailable),
	"https://errors.yuelili.com/problems/common.not_ready",
)

func ReadinessHandler(database *sql.DB) func(*ghttp.Request) {
	runner := health.MustRunner(map[string]health.Check{
		"database": func(ctx context.Context) error { return database.PingContext(ctx) },
	}, health.RunnerOptions{Timeout: 3 * time.Second})
	handler, err := goframehealth.Handler(goframehealth.HandlerOptions{
		Runner:   runner,
		NotReady: notReady,
		TraceID:  func(request *ghttp.Request) string { return request.Header.Get(TraceHeader) },
	})
	if err != nil {
		panic(err)
	}
	return handler
}
