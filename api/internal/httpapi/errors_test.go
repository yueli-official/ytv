package httpapi

import (
	"errors"
	"testing"

	"github.com/yueli-official/foundation/go/problem"
	"github.com/yueli-official/ytv/api/internal/apierr"
	"github.com/yueli-official/ytv/api/internal/playback"
	"github.com/yueli-official/ytv/api/internal/progress"
	"github.com/yueli-official/ytv/api/internal/source"
	"github.com/yueli-official/ytv/api/internal/sourcecontrol"
)

func TestPublicErrorClassificationPreservesCauses(t *testing.T) {
	for _, test := range []struct {
		cause  error
		code   string
		status int
	}{
		{source.ErrUnavailable, apierr.CodeSourceUnavailable, 503},
		{playback.ErrUnsafeURL, apierr.CodePlaybackUnsafe, 409},
		{sourcecontrol.ErrRevisionConflict, apierr.CodeRevisionConflict, 409},
		{sourcecontrol.ErrAlreadyInstalled, apierr.CodeSourceAlreadyInstalled, 409},
		{progress.ErrInvalid, apierr.CodeInvalidInput, 400},
	} {
		mapped := mapError(test.cause)
		if !errors.Is(mapped, test.cause) {
			t.Fatal("mapping discarded domain cause")
		}
		value, ok, err := problem.FromError(mapped, "ytv-classification-test")
		if err != nil || !ok || value.Code != test.code || value.Status != test.status {
			t.Fatal("incorrect public classification")
		}
	}
	internal := errors.New("storage-diagnostic-marker")
	if mapError(internal) != internal {
		t.Fatal("storage failure must remain internal, not become a validation error")
	}
}
