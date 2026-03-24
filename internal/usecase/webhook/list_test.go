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

func TestList_Success(t *testing.T) {
	ctx := t.Context()
	uc, repo := setup(t)

	filter := domain.WebhookFilter{
		Search: "order",
		Limit:  10,
		Offset: 0,
	}
	expected := []domain.Webhook{
		{
			ID:        uuid.New(),
			Name:      "order-hook",
			URL:       "https://example.com/hook1",
			Events:    []string{"order.created"},
			Headers:   map[string]any{},
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			Name:      "order-update-hook",
			URL:       "https://example.com/hook2",
			Events:    []string{"order.updated"},
			Headers:   map[string]any{},
			IsActive:  false,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	repo.On("List", ctx, filter).Return(expected, int64(2), nil)

	items, total, err := uc.List(ctx, filter)

	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, items, 2)
	assert.Equal(t, expected, items)
	repo.AssertExpectations(t)
}

func TestList_Empty(t *testing.T) {
	ctx := t.Context()
	uc, repo := setup(t)

	filter := domain.WebhookFilter{Limit: 10}

	repo.On("List", ctx, filter).Return([]domain.Webhook{}, int64(0), nil)

	items, total, err := uc.List(ctx, filter)

	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, items)
	repo.AssertExpectations(t)
}

func TestList_RepoError(t *testing.T) {
	ctx := t.Context()
	uc, repo := setup(t)

	filter := domain.WebhookFilter{Limit: 10}

	repo.On("List", ctx, filter).Return([]domain.Webhook{}, int64(0), errors.New("connection refused"))

	items, total, err := uc.List(ctx, filter)

	require.Error(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, items)
	assert.Contains(t, err.Error(), "connection refused")
	repo.AssertExpectations(t)
}
