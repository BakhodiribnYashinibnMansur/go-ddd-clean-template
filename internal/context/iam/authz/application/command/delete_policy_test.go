package command

import (
	"context"
	"testing"

	"gct/internal/context/iam/authz/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestDeletePolicyHandler_Success(t *testing.T) {
	t.Parallel()

	repo := &mockPolicyRepository{}
	log := &mockLogger{}

	handler := NewDeletePolicyHandler(repo, log)

	cmd := DeletePolicyCommand{ID: domain.PolicyID(uuid.New())}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)
}
