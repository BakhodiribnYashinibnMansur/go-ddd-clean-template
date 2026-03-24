package webhook_test

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

func TestUpdate_SuccessPartial(t *testing.T) {
	ctx := t.Context()
	uc, repo := setup(t)

	id := uuid.New()
	existing := &domain.Webhook{
		ID:        id,
		Name:      "old-name",
		URL:       "https://old.example.com/hook",
		Secret:    "old-secret",
		Events:    []string{"order.created"},
		Headers:   map[string]any{"X-Old": "val"},
		IsActive:  false,
		CreatedAt: time.Now().Add(-time.Hour),
		UpdatedAt: time.Now().Add(-time.Hour),
	}

	newName := "new-name"
	newActive := true
	req := domain.UpdateWebhookRequest{
		Name:     &newName,
		IsActive: &newActive,
		// URL, Secret, Events, Headers left nil => should keep old values
	}

	repo.On("GetByID", ctx, id).Return(existing, nil)
	repo.On("Update", ctx, mock.MatchedBy(func(w *domain.Webhook) bool {
		return w.Name == "new-name" &&
			w.URL == "https://old.example.com/hook" && // unchanged
			w.Secret == "old-secret" && // unchanged
			w.IsActive == true
	})).Return(nil)

	result, err := uc.Update(ctx, id, req)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "new-name", result.Name)
	assert.Equal(t, "https://old.example.com/hook", result.URL)
	assert.Equal(t, "old-secret", result.Secret)
	assert.True(t, result.IsActive)
	repo.AssertExpectations(t)
}

func TestUpdate_NotFoundOnGet(t *testing.T) {
	ctx := t.Context()
	uc, repo := setup(t)

	id := uuid.New()
	newName := "updated"
	req := domain.UpdateWebhookRequest{Name: &newName}

	repo.On("GetByID", ctx, id).Return(nil, errors.New("not found"))

	result, err := uc.Update(ctx, id, req)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not found")
	repo.AssertNotCalled(t, "Update")
	repo.AssertExpectations(t)
}

func TestUpdate_RepoUpdateError(t *testing.T) {
	ctx := t.Context()
	uc, repo := setup(t)

	id := uuid.New()
	existing := &domain.Webhook{
		ID:       id,
		Name:     "hook",
		URL:      "https://example.com/hook",
		Events:   []string{},
		Headers:  map[string]any{},
		IsActive: true,
	}

	newURL := "https://new.example.com/hook"
	req := domain.UpdateWebhookRequest{URL: &newURL}

	repo.On("GetByID", ctx, id).Return(existing, nil)
	repo.On("Update", ctx, mock.AnythingOfType("*domain.Webhook")).
		Return(errors.New("update failed"))

	result, err := uc.Update(ctx, id, req)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "update failed")
	repo.AssertExpectations(t)
}
