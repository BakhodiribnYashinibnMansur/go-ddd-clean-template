package iprule_test

import (
	"errors"
	"testing"
	"time"

	"gct/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUpdate_PartialSuccess(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	uc, repo := setup(t)

	id := uuid.New()
	existing := &domain.IPRule{
		ID:        id,
		IPAddress: "192.168.1.1",
		Type:      "whitelist",
		Reason:    "Old reason",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	newReason := "Updated reason"
	newActive := false
	req := domain.UpdateIPRuleRequest{
		Reason:   &newReason,
		IsActive: &newActive,
	}

	repo.On("GetByID", ctx, id).Return(existing, nil)
	repo.On("Update", ctx, mock.AnythingOfType("*domain.IPRule")).Return(nil)

	result, err := uc.Update(ctx, id, req)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "192.168.1.1", result.IPAddress)
	assert.Equal(t, "whitelist", result.Type)
	assert.Equal(t, newReason, result.Reason)
	assert.Equal(t, false, result.IsActive)
	repo.AssertExpectations(t)
}

func TestUpdate_NotFound(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	uc, repo := setup(t)

	id := uuid.New()
	newIP := "10.0.0.1"
	req := domain.UpdateIPRuleRequest{
		IPAddress: &newIP,
	}

	repo.On("GetByID", ctx, id).Return(nil, errors.New("not found"))

	result, err := uc.Update(ctx, id, req)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not found")
	repo.AssertExpectations(t)
}

func TestUpdate_RepoUpdateError(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	uc, repo := setup(t)

	id := uuid.New()
	existing := &domain.IPRule{
		ID:        id,
		IPAddress: "192.168.1.1",
		Type:      "whitelist",
		Reason:    "Office IP",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	newIP := "10.0.0.1"
	req := domain.UpdateIPRuleRequest{
		IPAddress: &newIP,
	}

	repo.On("GetByID", ctx, id).Return(existing, nil)
	repo.On("Update", ctx, mock.AnythingOfType("*domain.IPRule")).
		Return(errors.New("update failed"))

	result, err := uc.Update(ctx, id, req)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "update failed")
	repo.AssertExpectations(t)
}
