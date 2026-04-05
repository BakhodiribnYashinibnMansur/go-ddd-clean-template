package command

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestDeleteIPRuleHandler_Handle(t *testing.T) {
	repo := &mockIPRuleRepo{}
	handler := NewDeleteIPRuleHandler(repo, &mockLogger{})

	id := uuid.New()
	err := handler.Handle(context.Background(), DeleteIPRuleCommand{ID: id})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if repo.deleted != id {
		t.Errorf("expected deleted ID %s, got %s", id, repo.deleted)
	}
}
