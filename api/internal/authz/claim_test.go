package authz

import (
	"context"
	foundationauth "github.com/yueli-official/foundation/go/auth"
	"github.com/yueli-official/foundation/go/authorization"
	"testing"
)

func TestInitialClaimUsesAuthenticatedUserAndCannotBeReclaimed(t *testing.T) {
	runtime, err := authorization.NewMemory(authorization.MustCompile(Definition()), authorization.MemoryOptions{RootScopeID: RootScopeID, AllowUnclaimed: true})
	if err != nil {
		t.Fatal(err)
	}
	service := New(runtime)
	if _, err = service.ClaimInitialAdministrator(context.Background()); err == nil {
		t.Fatal("anonymous claim accepted")
	}
	user := func(id string) context.Context {
		return foundationauth.NewContext(context.Background(), &foundationauth.Principal{Subject: id, SubjectKind: foundationauth.SubjectUser})
	}
	first := user("first")
	if service.IsAdministrator(first) {
		t.Fatal("login alone granted administrator")
	}
	result, err := service.ClaimInitialAdministrator(first)
	if err != nil || !result.Created || !service.IsAdministrator(first) {
		t.Fatalf("claim = %#v, %v", result, err)
	}
	if _, err = service.ClaimInitialAdministrator(user("other")); !authorization.Is(err, authorization.ErrorConflict) {
		t.Fatalf("second claim = %v", err)
	}
	result, err = service.ClaimInitialAdministrator(first)
	if err != nil || result.Created {
		t.Fatalf("retry = %#v, %v", result, err)
	}
}
