package authz

import "github.com/yueli-official/foundation/go/authorization"

const (
	RootScopeID authorization.ScopeID   = "ytv"
	ScopeSite   authorization.ScopeType = "site"

	RoleAdministrator authorization.RoleKey = "administrator"

	CapabilityPublicRead     authorization.CapabilityKey = "ytv.public.read"
	CapabilityProgressSync   authorization.CapabilityKey = "ytv.progress.sync"
	CapabilitySourceManage   authorization.CapabilityKey = "ytv.source.manage"
	CapabilityAnalyticsRead  authorization.CapabilityKey = "ytv.analytics.read"
	CapabilitySettingsManage authorization.CapabilityKey = "ytv.settings.manage"
)

func Definition() authorization.Definition {
	return authorization.Definition{
		Consumer: "ytv",
		Version:  1,
		Capabilities: []authorization.CapabilityDefinition{
			accessCapability(CapabilityPublicRead),
			accessCapability(CapabilityProgressSync),
			protectedCapability(CapabilitySourceManage),
			protectedCapability(CapabilityAnalyticsRead),
			protectedCapability(CapabilitySettingsManage),
		},
		Scopes: authorization.ScopeSchema{Types: []authorization.ScopeTypeDefinition{{Key: ScopeSite, Root: true}}},
		AccessLayers: []authorization.AccessLayerDefinition{
			{Key: authorization.AccessLayerVisitor, Capabilities: []authorization.CapabilityKey{CapabilityPublicRead}},
			{Key: authorization.AccessLayerAuthenticated, Capabilities: []authorization.CapabilityKey{
				authorization.CapabilityApplicationCreate,
				authorization.CapabilityApplicationReadOwn,
				authorization.CapabilityApplicationWithdraw,
				authorization.CapabilityInvitationAccept,
				CapabilityProgressSync,
			}},
		},
		Roles: []authorization.RoleDefinition{{
			Key: RoleAdministrator, DisplayName: "管理员", Protected: true,
			Capabilities: []authorization.CapabilityKey{
				authorization.CapabilityManage,
				authorization.CapabilityAuditRead,
				CapabilitySourceManage,
				CapabilityAnalyticsRead,
				CapabilitySettingsManage,
			},
		}},
	}
}

func accessCapability(key authorization.CapabilityKey) authorization.CapabilityDefinition {
	return authorization.CapabilityDefinition{
		Key: key, Version: 1, Binding: authorization.BindingAccessLayerEligible,
		AllowedScopes: []authorization.ScopeType{ScopeSite},
	}
}

func protectedCapability(key authorization.CapabilityKey) authorization.CapabilityDefinition {
	return authorization.CapabilityDefinition{
		Key: key, Version: 1, Binding: authorization.BindingProtectedOnly,
		Risk: authorization.RiskHigh, Audit: authorization.AuditFull,
		AllowedScopes:    []authorization.ScopeType{ScopeSite},
		EligibleSubjects: []authorization.SubjectKind{authorization.SubjectUser, authorization.SubjectService},
	}
}
