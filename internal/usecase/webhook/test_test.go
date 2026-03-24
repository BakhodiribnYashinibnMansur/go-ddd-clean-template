package webhook_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTest_GetByIDError(t *testing.T) {
	ctx := t.Context()
	uc, repo := setup(t)

	id := uuid.New()

	repo.On("GetByID", ctx, id).Return(nil, errors.New("webhook not found"))

	err := uc.Test(ctx, id)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "webhook not found")
	repo.AssertExpectations(t)
}

func TestTest_HTTPCall_Skipped(t *testing.T) {
	// The Test method makes a real HTTP POST to the webhook URL after GetByID.
	// We skip this test because it requires a live HTTP endpoint.
	// The GetByID error path is covered above.
	t.Skip("Test method makes real HTTP calls; skipping in unit tests")
}
