package command

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestDeleteTranslationHandler_Handle(t *testing.T) {
	repo := &mockRepo{}
	log := &mockLogger{}

	handler := NewDeleteTranslationHandler(repo, log)

	err := handler.Handle(context.Background(), DeleteTranslationCommand{
		ID: uuid.New(),
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestDeleteTranslationHandler_RepoError(t *testing.T) {
	repoErr := errors.New("repo delete failed")
	errR := &errorRepo{deleteErr: repoErr}
	log := &mockLogger{}

	handler := NewDeleteTranslationHandler(errR, log)

	err := handler.Handle(context.Background(), DeleteTranslationCommand{
		ID: uuid.New(),
	})
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo delete error, got: %v", err)
	}
}
