package iprule_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDelete_Success(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	uc, repo := setup(t)

	id := uuid.New()

	repo.On("Delete", ctx, id).Return(nil)

	err := uc.Delete(ctx, id)

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestDelete_Error(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	uc, repo := setup(t)

	id := uuid.New()

	repo.On("Delete", ctx, id).Return(errors.New("delete failed"))

	err := uc.Delete(ctx, id)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "delete failed")
	repo.AssertExpectations(t)
}
