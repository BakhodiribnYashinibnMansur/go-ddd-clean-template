package emailtemplate_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDelete_Success(t *testing.T) {
	ctx := t.Context()
	uc, repo := setup(t)

	repo.On("Delete", ctx, "tmpl-1").Return(nil)

	err := uc.Delete(ctx, "tmpl-1")

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestDelete_RepoError(t *testing.T) {
	ctx := t.Context()
	uc, repo := setup(t)

	repo.On("Delete", ctx, "tmpl-1").
		Return(errors.New("delete failed"))

	err := uc.Delete(ctx, "tmpl-1")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "delete failed")
	repo.AssertExpectations(t)
}
