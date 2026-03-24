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

func TestList_Success(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	uc, repo := setup(t)

	filter := domain.NotificationFilter{
		Search: "welcome",
		Limit:  10,
		Offset: 0,
	}

	expected := []domain.Notification{
		{
			ID:         uuid.New(),
			Title:      "Welcome",
			Body:       "Hello!",
			Type:       "info",
			TargetType: "all",
			IsActive:   true,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
		{
			ID:         uuid.New(),
			Title:      "Welcome Back",
			Body:       "Nice to see you again",
			Type:       "info",
			TargetType: "user",
			IsActive:   true,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
	}

	repo.On("List", ctx, filter).Return(expected, int64(2), nil)

	items, total, err := uc.List(ctx, filter)

	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, items, 2)
	assert.Equal(t, expected[0].Title, items[0].Title)
	repo.AssertExpectations(t)
}

func TestList_Empty(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	uc, repo := setup(t)

	filter := domain.NotificationFilter{
		Search: "nonexistent",
		Limit:  10,
	}

	repo.On("List", ctx, filter).Return([]domain.Notification{}, int64(0), nil)

	items, total, err := uc.List(ctx, filter)

	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, items)
	repo.AssertExpectations(t)
}

func TestList_Error(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	uc, repo := setup(t)

	filter := domain.NotificationFilter{Limit: 10}

	repo.On("List", ctx, filter).Return([]domain.Notification{}, int64(0), errors.New("database error"))

	items, total, err := uc.List(ctx, filter)

	require.Error(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, items)
	assert.Contains(t, err.Error(), "database error")
	repo.AssertExpectations(t)
}
