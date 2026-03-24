package file_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteFile_Success(t *testing.T) {
	ctx := t.Context()
	uc, repo := setup(t)

	id := "file-123"

	repo.On("Delete", ctx, id).Return(nil)

	err := uc.DeleteFile(ctx, id)

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestDeleteFile_RepoError(t *testing.T) {
	ctx := t.Context()
	uc, repo := setup(t)

	id := "file-456"

	repo.On("Delete", ctx, id).Return(errors.New("delete failed"))

	err := uc.DeleteFile(ctx, id)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "delete failed")
	repo.AssertExpectations(t)
}
