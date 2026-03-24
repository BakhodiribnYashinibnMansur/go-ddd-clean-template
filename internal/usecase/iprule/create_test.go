package iprule_test

import (
	"errors"
	"testing"

	"gct/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreate_Success(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	uc, repo := setup(t)

	req := domain.CreateIPRuleRequest{
		IPAddress: "192.168.1.100",
		Type:      "whitelist",
		Reason:    "Trusted office IP",
		IsActive:  true,
	}

	repo.On("Create", ctx, mock.AnythingOfType("*domain.IPRule")).
		Return(nil)

	result, err := uc.Create(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotEqual(t, result.ID.String(), "00000000-0000-0000-0000-000000000000")
	assert.Equal(t, req.IPAddress, result.IPAddress)
	assert.Equal(t, req.Type, result.Type)
	assert.Equal(t, req.Reason, result.Reason)
	assert.Equal(t, req.IsActive, result.IsActive)
	repo.AssertExpectations(t)
}

func TestCreate_RepoError(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	uc, repo := setup(t)

	req := domain.CreateIPRuleRequest{
		IPAddress: "10.0.0.1",
		Type:      "blacklist",
		Reason:    "Suspicious activity",
		IsActive:  true,
	}

	repo.On("Create", ctx, mock.AnythingOfType("*domain.IPRule")).
		Return(errors.New("database error"))

	result, err := uc.Create(ctx, req)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "database error")
	repo.AssertExpectations(t)
}
