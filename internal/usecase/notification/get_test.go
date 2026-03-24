package notification_test

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
	expected := &domain.Notification{
		ID:         id,
		Title:      "Test Notification",
		Body:       "Test body",
		Type:       "info",
		TargetType: "all",
		IsActive:   true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	repo.On("GetByID", ctx, id).Return(expected, nil)

	result, err := uc.GetByID(ctx, id)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, expected.ID, result.ID)
	assert.Equal(t, expected.Title, result.Title)
	assert.Equal(t, expected.Body, result.Body)
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
