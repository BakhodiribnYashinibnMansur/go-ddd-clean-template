package command_test

import (
	"context"
	"testing"

	"gct/internal/context/iam/usersetting/application/command"

	"github.com/google/uuid"
)

func TestDeleteUserSettingHandler_Handle(t *testing.T) {
	repo := &mockUserSettingRepo{}
	handler := command.NewDeleteUserSettingHandler(repo, &mockLogger{})

	id := uuid.New()
	cmd := command.DeleteUserSettingCommand{ID: id}

	err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	if repo.deleted != id {
		t.Fatalf("expected deleted ID %s, got %s", id, repo.deleted)
	}
}
