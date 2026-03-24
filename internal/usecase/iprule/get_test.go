package iprule_test

import (
	"errors"
	"testing"
	"time"

	"gct/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetByID_Success(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	uc, repo := setup(t)

	id := uuid.New()
	expected := &domain.IPRule{
		ID:        id,
		IPAddress: "192.168.1.100",
		Type:      "whitelist",
		Reason:    "Trusted IP",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	repo.On("GetByID", ctx, id).Return(expected, nil)

	result, err := uc.GetByID(ctx, id)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, expected.ID, result.ID)
	assert.Equal(t, expected.IPAddress, result.IPAddress)
	assert.Equal(t, expected.Type, result.Type)
	repo.AssertExpectations(t)
}

func TestGetByID_NotFound(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	uc, repo := setup(t)

	id := uuid.New()

	repo.On("GetByID", ctx, id).Return(nil, errors.New("not found"))

	result, err := uc.GetByID(ctx, id)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not found")
	repo.AssertExpectations(t)
}
