package authz

import (
	"context"
	"testing"

	"github.com/yueli-official/foundation/go/authorization"
)

func TestDefinitionKeepsPublicAndManagementCapabilitiesSeparate(t *testing.T) {
	definition, err := authorization.Compile(Definition())
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := authorization.NewMemory(definition, authorization.MemoryOptions{
		RootScopeID:       RootScopeID,
		ProtectedSubjects: []authorization.SubjectRef{{Kind: authorization.SubjectUser, ID: "admin"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	visitor, err := runtime.Decide(context.Background(), authorization.DecisionRequest{
		Subject: authorization.SubjectRef{Kind: authorization.SubjectAnonymous}, Capability: CapabilityPublicRead, ScopeID: RootScopeID,
	})
	if err != nil || !visitor.Allowed {
		t.Fatalf("visitor decision = %#v, %v", visitor, err)
	}
	management, err := runtime.Decide(context.Background(), authorization.DecisionRequest{
		Subject: authorization.SubjectRef{Kind: authorization.SubjectAnonymous}, Capability: CapabilitySourceManage, ScopeID: RootScopeID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if management.Allowed {
		t.Fatal("anonymous subject unexpectedly manages sources")
	}
}
