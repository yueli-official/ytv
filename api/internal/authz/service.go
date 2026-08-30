package authz

import (
	"context"

	foundationauth "github.com/yueli-official/foundation/go/auth"
	"github.com/yueli-official/foundation/go/authorization"
)

type Runtime interface {
	authorization.Authorizer
	authorization.AccessReader
}

type Service struct {
	runtime Runtime
}

func New(runtime Runtime) *Service {
	return &Service{runtime: runtime}
}

func (service *Service) Subject(ctx context.Context) authorization.SubjectRef {
	principal, ok := foundationauth.FromContext(ctx)
	if !ok {
		return authorization.SubjectRef{Kind: authorization.SubjectAnonymous}
	}
	if principal.SubjectKind == foundationauth.SubjectUser && principal.Subject != "" {
		return authorization.SubjectRef{Kind: authorization.SubjectUser, ID: principal.Subject}
	}
	if principal.SubjectKind == foundationauth.SubjectClient && principal.ClientID != "" {
		return authorization.SubjectRef{Kind: authorization.SubjectService, ID: principal.ClientID}
	}
	return authorization.SubjectRef{Kind: authorization.SubjectAnonymous}
}

func (service *Service) IsAdministrator(ctx context.Context) bool {
	decision, err := service.runtime.Decide(ctx, authorization.DecisionRequest{
		Subject: service.Subject(ctx), Capability: authorization.CapabilityManage, ScopeID: RootScopeID,
	})
	return err == nil && decision.Allowed
}
