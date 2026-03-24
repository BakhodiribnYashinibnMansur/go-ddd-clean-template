package notification_test

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
	existing := &domain.Notification{
		ID:         id,
		Title:      "Old Title",
		Body:       "Old Body",
		Type:       "info",
		TargetType: "all",
		IsActive:   true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	newTitle := "New Title"
	newActive := false
	req := domain.UpdateNotificationRequest{
		Title:    &newTitle,
		IsActive: &newActive,
	}

	repo.On("GetByID", ctx, id).Return(existing, nil)
	repo.On("Update", ctx, mock.AnythingOfType("*domain.Notification")).Return(nil)

	result, err := uc.Update(ctx, id, req)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, newTitle, result.Title)
	assert.Equal(t, "Old Body", result.Body)
	assert.Equal(t, "info", result.Type)
	assert.Equal(t, "all", result.TargetType)
	assert.Equal(t, false, result.IsActive)
	repo.AssertExpectations(t)
}

func TestUpdate_NotFound(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	uc, repo := setup(t)

	id := uuid.New()
	newTitle := "New Title"
	req := domain.UpdateNotificationRequest{
		Title: &newTitle,
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
	existing := &domain.Notification{
		ID:         id,
		Title:      "Old Title",
		Body:       "Old Body",
		Type:       "info",
		TargetType: "all",
		IsActive:   true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	newTitle := "New Title"
	req := domain.UpdateNotificationRequest{
		Title: &newTitle,
	}

	repo.On("GetByID", ctx, id).Return(existing, nil)
	repo.On("Update", ctx, mock.AnythingOfType("*domain.Notification")).
		Return(errors.New("update failed"))

	result, err := uc.Update(ctx, id, req)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "update failed")
	repo.AssertExpectations(t)
}
