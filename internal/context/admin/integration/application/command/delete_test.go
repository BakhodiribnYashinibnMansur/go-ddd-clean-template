package command

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestDeleteHandler_Handle(t *testing.T) {
	repo := &mockIntegrationRepo{}
	handler := NewDeleteHandler(repo, &mockEventBus{}, &mockLogger{})

	id := uuid.New()
	err := handler.Handle(context.Background(), DeleteCommand{ID: id})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if repo.deleted != id {
		t.Errorf("expected deleted ID %s, got %s", id, repo.deleted)
	}
}
