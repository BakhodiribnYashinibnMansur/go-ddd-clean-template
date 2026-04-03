package pgxutil

import (
	"context"
	"errors"
	"testing"
)

func TestAppSpan_Success(t *testing.T) {
	ctx := context.Background()
	ctx, end := AppSpan(ctx, "TestOp.Success")
	end(nil)
	_ = ctx
}

func TestAppSpan_Error(t *testing.T) {
	ctx := context.Background()
	ctx, end := AppSpan(ctx, "TestOp.Error")
	end(errors.New("something went wrong"))
	_ = ctx
}

func TestAppSpan_ReturnsContext(t *testing.T) {
	ctx, end := AppSpan(context.Background(), "TestOp.Context")
	defer end(nil)

	if ctx == nil {
		t.Fatal("expected non-nil context from AppSpan")
	}
}
