package command

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestDeleteAnnouncementHandler_Handle(t *testing.T) {
	repo := &mockAnnouncementRepo{}
	handler := NewDeleteAnnouncementHandler(repo, &mockLogger{})

	id := uuid.New()
	err := handler.Handle(context.Background(), DeleteAnnouncementCommand{ID: id})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if repo.deleted != id {
		t.Errorf("expected deleted ID %s, got %s", id, repo.deleted)
	}
}
