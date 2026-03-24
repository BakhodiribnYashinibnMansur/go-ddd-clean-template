package webhook_test

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
	ctx := t.Context()
	uc, repo := setup(t)

	id := uuid.New()
	expected := &domain.Webhook{
		ID:        id,
		Name:      "my-hook",
		URL:       "https://example.com/hook",
		Secret:    "secret",
		Events:    []string{"order.created"},
		Headers:   map[string]any{"X-Key": "val"},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	repo.On("GetByID", ctx, id).Return(expected, nil)

	result, err := uc.GetByID(ctx, id)

	require.NoError(t, err)
	assert.Equal(t, expected, result)
	repo.AssertExpectations(t)
}

func TestGetByID_NotFound(t *testing.T) {
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
