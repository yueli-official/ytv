package authz

import (
	"context"

	foundationauth "github.com/yueli-official/foundation/go/auth"
	"github.com/yueli-official/foundation/go/authorization"
)

type Runtime interface {
	authorization.Authorizer
	authorization.AccessReader
	authorization.AdministratorClaimer
}

type Service struct {
	runtime       Runtime
	bootstrapMode bool
}

func New(runtime Runtime, bootstrap ...bool) *Service {
	enabled := true
	if len(bootstrap) > 0 {
		enabled = bootstrap[0]
	}
	return &Service{runtime: runtime, bootstrapMode: enabled}
}
func (service *Service) BootstrapMode() bool { return service.bootstrapMode }

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

func (service *Service) AdministratorClaimStatus(ctx context.Context) (authorization.AdministratorClaimStatus, error) {
	return service.runtime.AdministratorClaimStatus(ctx)
}
func (service *Service) ClaimInitialAdministrator(ctx context.Context) (authorization.ClaimInitialAdministratorResult, error) {
	return service.runtime.ClaimInitialAdministrator(ctx, authorization.ClaimInitialAdministratorCommand{Actor: service.Subject(ctx)})
}
