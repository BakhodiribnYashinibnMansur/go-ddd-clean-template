package pgxutil

import (
	"context"
	"errors"
	"testing"
)

func TestRepoSpan_Success(t *testing.T) {
	ctx := context.Background()
	ctx, end := RepoSpan(ctx, "UserRepo.Find")
	end(nil)
	_ = ctx
}

func TestRepoSpan_Error(t *testing.T) {
	ctx := context.Background()
	ctx, end := RepoSpan(ctx, "UserRepo.Save")
	end(errors.New("db connection lost"))
	_ = ctx
}

func TestRepoSpanSimple(t *testing.T) {
	ctx, span := RepoSpanSimple(context.Background(), "UserRepo.List")
	defer span.End()

	if ctx == nil {
		t.Fatal("expected non-nil context from RepoSpanSimple")
	}
	if span == nil {
		t.Fatal("expected non-nil span from RepoSpanSimple")
	}
}
