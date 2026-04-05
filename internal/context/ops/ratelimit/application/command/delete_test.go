package command

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestDeleteRateLimitHandler_Handle(t *testing.T) {
	repo := &mockRateLimitRepo{}
	handler := NewDeleteRateLimitHandler(repo, &mockLogger{})

	id := uuid.New()
	err := handler.Handle(context.Background(), DeleteRateLimitCommand{ID: id})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if repo.deleted != id {
		t.Errorf("expected deleted ID %s, got %s", id, repo.deleted)
	}
}
