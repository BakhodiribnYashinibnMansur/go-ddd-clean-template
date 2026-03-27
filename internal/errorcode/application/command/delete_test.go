package command

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestDeleteErrorCodeHandler_Handle(t *testing.T) {
	repo := &mockErrorCodeRepo{}
	handler := NewDeleteErrorCodeHandler(repo, &mockLogger{})

	id := uuid.New()
	err := handler.Handle(context.Background(), DeleteErrorCodeCommand{ID: id})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if repo.deleted != id {
		t.Errorf("expected deleted ID %s, got %s", id, repo.deleted)
	}
}
