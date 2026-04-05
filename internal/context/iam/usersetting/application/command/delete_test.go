package command_test

import (
	"context"
	"testing"

	"gct/internal/context/iam/usersetting/application/command"
	"gct/internal/context/iam/usersetting/domain"
	"github.com/stretchr/testify/require"
)

func TestDeleteUserSettingHandler_Handle(t *testing.T) {
	t.Parallel()

	repo := &mockUserSettingRepo{}
	handler := command.NewDeleteUserSettingHandler(repo, &mockLogger{})

	id := domain.NewUserSettingID()
	cmd := command.DeleteUserSettingCommand{ID: id}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	if repo.deleted != id.UUID() {
		t.Fatalf("expected deleted ID %s, got %s", id, repo.deleted)
	}
}
