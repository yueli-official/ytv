package runtime

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/guid"
	"github.com/yueli-official/foundation/go/goframe/api"
	"github.com/yueli-official/foundation/go/goframe/ratelimit"
	"github.com/yueli-official/ytv/api/internal/apierr"
)

const TraceHeader = "X-Trace-Id"

func MustAPIMiddleware() *api.Middleware {
	limit := 300
	if raw := os.Getenv("YTV_RATE_LIMIT_PER_MINUTE"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 0 {
			panic(fmt.Errorf("YTV_RATE_LIMIT_PER_MINUTE must be a non-negative integer"))
		}
		limit = parsed
	}
	limiter := ratelimit.MustNew(ratelimit.Policy{Limit: limit, Window: time.Minute})
	middleware, err := api.New(api.Options{
		TraceHeader: TraceHeader,
		RateLimited: apierr.RateLimited,
		Validation:  apierr.Validation,
		Internal:    apierr.Internal,
		Limiter:     limiter,
		ClientKey:   func(request *ghttp.Request) string { return request.GetClientIp() },
	})
	if err != nil {
		panic(err)
	}
	return middleware
}

func TraceRoute(request *ghttp.Request) {
	if request.Header.Get(TraceHeader) == "" {
		request.Header.Set(TraceHeader, guid.S())
	}
	api.TraceRoute(request)
}
