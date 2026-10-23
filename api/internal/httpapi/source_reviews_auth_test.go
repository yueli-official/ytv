package httpapi

import (
	"context"
	"github.com/yueli-official/foundation/go/auth"
	"testing"
)

func TestReviewIdentity(t *testing.T) {
	for _, tc := range []struct {
		name          string
		kind          auth.SubjectKind
		subject, want string
	}{
		{"user", auth.SubjectUser, "viewer", "viewer"},
		{"guest isolated from user", auth.SubjectGuest, "viewer", "guest:viewer"},
		{"client cannot vote", auth.SubjectClient, "viewer", ""},
		{"empty guest", auth.SubjectGuest, " ", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := auth.NewContext(context.Background(), &auth.Principal{Subject: tc.subject, SubjectKind: tc.kind})
			got, err := requiredReviewer(ctx)
			if got != tc.want || (err != nil) != (tc.want == "") {
				t.Fatalf("got %q, %v", got, err)
			}
			if tc.kind == auth.SubjectGuest {
				if _, err := requiredUser(ctx); err == nil {
					t.Fatal("guest gained account-only access")
				}
			}
		})
	}
	if _, err := requiredReviewer(context.Background()); err == nil {
		t.Fatal("missing identity accepted")
	}
}
