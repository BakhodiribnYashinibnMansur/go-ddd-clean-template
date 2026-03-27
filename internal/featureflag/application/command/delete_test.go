package command

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestDeleteHandler_Handle(t *testing.T) {
	repo := &mockRepo{}
	eb := &mockEventBus{}
	log := &mockLogger{}

	handler := NewDeleteHandler(repo, eb, log)

	err := handler.Handle(context.Background(), DeleteCommand{
		ID: uuid.New(),
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestDeleteHandler_RepoError(t *testing.T) {
	repoErr := errors.New("repo delete failed")
	errR := &errorRepo{deleteErr: repoErr}
	eb := &mockEventBus{}
	log := &mockLogger{}

	handler := NewDeleteHandler(errR, eb, log)

	err := handler.Handle(context.Background(), DeleteCommand{
		ID: uuid.New(),
	})
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo delete error, got: %v", err)
	}
}
