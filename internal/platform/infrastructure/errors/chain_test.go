package errors_test

import (
	"context"
	"testing"
	"time"

	"gct/internal/platform/infrastructure/contextx"
	apperrors "gct/internal/platform/infrastructure/errors"
)

func TestErrorChain_RecordAndGet(t *testing.T) {
	chain := apperrors.NewErrorChain(time.Minute)
	ctx := contextx.WithRequestID(context.Background(), "req-123")

	err1 := apperrors.New(apperrors.ErrRepoNotFound, "")
	err2 := apperrors.New(apperrors.ErrServiceNotFound, "")

	chain.Record(ctx, err1)
	chain.Record(ctx, err2)

	entries := chain.Get("req-123")
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Code != apperrors.ErrRepoNotFound {
		t.Fatalf("expected %s, got %s", apperrors.ErrRepoNotFound, entries[0].Code)
	}
	if entries[1].Code != apperrors.ErrServiceNotFound {
		t.Fatalf("expected %s, got %s", apperrors.ErrServiceNotFound, entries[1].Code)
	}
}

func TestErrorChain_EmptyRequestID(t *testing.T) {
	chain := apperrors.NewErrorChain(time.Minute)
	ctx := context.Background() // no request_id

	err := apperrors.New(apperrors.ErrInternal, "")
	chain.Record(ctx, err)

	entries := chain.Get("")
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries for empty request_id, got %d", len(entries))
	}
}

func TestErrorChain_Cleanup(t *testing.T) {
	chain := apperrors.NewErrorChain(50 * time.Millisecond)
	ctx := contextx.WithRequestID(context.Background(), "req-old")

	chain.Record(ctx, apperrors.New(apperrors.ErrInternal, ""))
	time.Sleep(60 * time.Millisecond)
	chain.Cleanup()

	entries := chain.Get("req-old")
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries after cleanup, got %d", len(entries))
	}
}
