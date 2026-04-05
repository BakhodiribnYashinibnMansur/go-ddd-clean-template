package command

import (
	"context"
	"testing"

	"gct/internal/context/iam/authz/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestDeletePermissionHandler_Success(t *testing.T) {
	t.Parallel()

	repo := &mockPermissionRepository{}
	log := &mockLogger{}

	handler := NewDeletePermissionHandler(repo, log)

	cmd := DeletePermissionCommand{ID: domain.PermissionID(uuid.New())}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)
}
