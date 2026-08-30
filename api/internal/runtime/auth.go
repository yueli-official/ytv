package runtime

import (
	"time"

	"github.com/gogf/gf/v2/net/ghttp"
	foundationauth "github.com/yueli-official/foundation/go/auth"
	goframeauth "github.com/yueli-official/foundation/go/goframe/auth"
	foundationhttp "github.com/yueli-official/foundation/go/goframe/http"
	"github.com/yueli-official/foundation/go/jwks"
	"github.com/yueli-official/ytv/api/internal/apierr"
)

type RemoteVerifierConfig struct {
	JWKSURL           string
	Issuer            string
	Audience          string
	AllowLoopbackHTTP bool
	Clock             func() time.Time
}

func NewRemoteVerifier(config RemoteVerifierConfig) (*foundationauth.Verifier, error) {
	keys, err := jwks.NewRemoteSource(config.JWKSURL, jwks.RemoteOptions{AllowLoopbackHTTP: config.AllowLoopbackHTTP})
	if err != nil {
		return nil, err
	}
	audiences := []string(nil)
	if config.Audience != "" {
		audiences = []string{config.Audience}
	}
	return foundationauth.NewVerifier(foundationauth.Config{
		Keys: keys, Issuer: config.Issuer, Audiences: audiences, Clock: config.Clock,
	})
}

func RequiredAuth(verifier goframeauth.TokenVerifier) func(*ghttp.Request) {
	return mustAuth(verifier).Required
}

func OptionalAuth(verifier goframeauth.TokenVerifier) func(*ghttp.Request) {
	return mustAuth(verifier).Optional
}

func mustAuth(verifier goframeauth.TokenVerifier) *goframeauth.Middleware {
	writer := foundationhttp.MustWriter(foundationhttp.WriterOptions{TraceHeader: TraceHeader})
	middleware, err := goframeauth.NewMiddleware(goframeauth.Options{
		Verifier:         verifier,
		Writer:           &writer,
		UnauthorizedKind: apierr.Unauthorized.Kind(),
		UnauthorizedType: apierr.Unauthorized.Type(),
		TraceID:          func(request *ghttp.Request) string { return request.Header.Get(TraceHeader) },
		Realm:            "ytv",
	})
	if err != nil {
		panic(err)
	}
	return middleware
}
