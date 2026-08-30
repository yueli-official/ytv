package apierr

import (
	"net/http"

	"github.com/yueli-official/foundation/go/problem"
)

const typeRoot = "https://errors.yuelili.com/problems/"

var (
	RateLimited  = descriptor("common.rate_limited", http.StatusTooManyRequests)
	Validation   = descriptor("validation.failed", http.StatusBadRequest)
	Internal     = descriptor("common.internal", http.StatusInternalServerError)
	Unauthorized = descriptor("ytv.not_authenticated", http.StatusUnauthorized)
	Forbidden    = descriptor("ytv.forbidden", http.StatusForbidden)
	NotFound     = descriptor("ytv.not_found", http.StatusNotFound)
	Conflict     = descriptor("ytv.conflict", http.StatusConflict)
)

func descriptor(code string, status int) problem.Descriptor {
	return problem.MustDescriptor(problem.MustKind(code, status), typeRoot+code)
}

func Error(descriptor problem.Descriptor) error {
	value, err := problem.NewError(descriptor, nil)
	if err != nil {
		return err
	}
	return value
}
