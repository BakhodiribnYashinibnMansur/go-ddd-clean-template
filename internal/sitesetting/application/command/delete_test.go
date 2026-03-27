package command

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestDeleteSiteSettingHandler_Handle(t *testing.T) {
	repo := &mockRepo{}
	log := &mockLogger{}

	handler := NewDeleteSiteSettingHandler(repo, log)

	err := handler.Handle(context.Background(), DeleteSiteSettingCommand{
		ID: uuid.New(),
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestDeleteSiteSettingHandler_RepoError(t *testing.T) {
	repoErr := errors.New("repo delete failed")
	errR := &errorRepo{deleteErr: repoErr}
	log := &mockLogger{}

	handler := NewDeleteSiteSettingHandler(errR, log)

	err := handler.Handle(context.Background(), DeleteSiteSettingCommand{
		ID: uuid.New(),
	})
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo delete error, got: %v", err)
	}
}
