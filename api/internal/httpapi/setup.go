package httpapi

import (
	"context"
	"github.com/yueli-official/foundation/go/authorization"
	v1 "github.com/yueli-official/ytv/api/api/v1"
	"github.com/yueli-official/ytv/api/internal/apierr"
)

func (controller *Public) GetAdministratorSetup(ctx context.Context, _ *v1.GetAdministratorSetupReq) (*v1.GetAdministratorSetupRes, error) {
	status, err := controller.core.authz.AdministratorClaimStatus(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.GetAdministratorSetupRes{Claimed: status.Claimed, CanClaim: !status.Claimed}, nil
}
func (controller *Account) ClaimAdministrator(ctx context.Context, _ *v1.ClaimAdministratorReq) (*v1.ClaimAdministratorRes, error) {
	if controller.core.authz.Subject(ctx).Kind != authorization.SubjectUser {
		return nil, apierr.Public(apierr.CodeForbidden)
	}
	result, err := controller.core.authz.ClaimInitialAdministrator(ctx)
	if authorization.Is(err, authorization.ErrorConflict) {
		return nil, apierr.Wrap(apierr.CodeConflict, err)
	}
	if err != nil {
		return nil, err
	}
	return &v1.ClaimAdministratorRes{Claimed: result.Status.Claimed, Created: result.Created}, nil
}
