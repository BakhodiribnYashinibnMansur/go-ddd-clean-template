package errorx_test

import (
	"context"
	"testing"

	apperrors "gct/internal/kernel/infrastructure/errorx"
)

func TestResolver_RegisterAndResolve(t *testing.T) {
	resolver := apperrors.NewResolver()
	resolved := false

	resolver.Register(apperrors.ErrRepoConnection, func(ctx context.Context, err *apperrors.AppError) bool {
		resolved = true
		return true
	})

	err := apperrors.New(apperrors.ErrRepoConnection, "")
	result := resolver.Resolve(context.Background(), err)

	if !result {
		t.Fatal("expected resolution to be applied")
	}
	if !resolved {
		t.Fatal("expected resolved flag to be true")
	}
}

func TestResolver_NoActionRegistered(t *testing.T) {
	resolver := apperrors.NewResolver()

	err := apperrors.New(apperrors.ErrBadRequest, "")
	result := resolver.Resolve(context.Background(), err)

	if result {
		t.Fatal("expected no resolution for unregistered code")
	}
}

func TestResolver_NilError(t *testing.T) {
	resolver := apperrors.NewResolver()
	result := resolver.Resolve(context.Background(), nil)

	if result {
		t.Fatal("expected false for nil error")
	}
}
