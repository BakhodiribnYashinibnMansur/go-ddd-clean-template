package command_test

import (
	"context"
	"testing"

	"gct/internal/context/iam/generic/usersetting/application/command"
	settingentity "gct/internal/context/iam/generic/usersetting/domain/entity"

	"github.com/stretchr/testify/require"
)

func TestDeleteUserSettingHandler_Handle(t *testing.T) {
	t.Parallel()

	repo := &mockUserSettingRepo{}
	handler := command.NewDeleteUserSettingHandler(repo, &mockLogger{})

	id := settingentity.NewUserSettingID()
	cmd := command.DeleteUserSettingCommand{ID: id}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	if repo.deleted != id {
		t.Fatalf("expected deleted ID %s, got %s", id, repo.deleted)
	}
}
