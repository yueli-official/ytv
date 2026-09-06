package apierr

import (
	"net/http"

	"github.com/yueli-official/foundation/go/problem"
)

const typeRoot = "https://errors.yueli.dev/problems/"

var (
	RateLimited  = commonDescriptor("common.rate_limited", http.StatusTooManyRequests)
	Validation   = commonDescriptor("common.validation_failed", http.StatusBadRequest)
	Internal     = commonDescriptor("common.internal", http.StatusInternalServerError)
	Unauthorized = descriptors[CodeNotAuthenticated]
	Forbidden    = descriptors[CodeForbidden]
	NotFound     = descriptors[CodeNotFound]
	Conflict     = descriptors[CodeConflict]
)

func commonDescriptor(code string, status int) problem.Descriptor {
	return problem.MustDescriptor(problem.MustKind(code, status), typeRoot+code)
}

func Error(descriptor problem.Descriptor) error {
	value, err := problem.NewError(descriptor, nil)
	if err != nil {
		return err
	}
	return value
}

func Public(code string) error { return Error(descriptors[code]) }
func Wrap(code string, cause error) error {
	value, err := problem.WrapError(descriptors[code], cause, nil)
	if err != nil {
		return cause
	}
	return value
}
