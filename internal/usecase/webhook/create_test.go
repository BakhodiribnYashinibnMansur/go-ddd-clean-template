package webhook_test

import (
	"errors"
	"testing"

	"gct/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreate_Success(t *testing.T) {
	ctx := t.Context()
	uc, repo := setup(t)

	req := domain.CreateWebhookRequest{
		Name:     "order.created",
		URL:      "https://example.com/hook",
		Secret:   "s3cret",
		Events:   []string{"order.created", "order.updated"},
		Headers:  map[string]any{"X-Custom": "value"},
		IsActive: true,
	}

	repo.On("Create", ctx, mock.AnythingOfType("*domain.Webhook")).
		Return(nil)

	result, err := uc.Create(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotEqual(t, result.ID.String(), "00000000-0000-0000-0000-000000000000")
	assert.Equal(t, req.Name, result.Name)
	assert.Equal(t, req.URL, result.URL)
	assert.Equal(t, req.Secret, result.Secret)
	assert.Equal(t, req.Events, result.Events)
	assert.Equal(t, req.Headers, result.Headers)
	assert.Equal(t, req.IsActive, result.IsActive)
	repo.AssertExpectations(t)
}

func TestCreate_NilEventsAndHeaders_Initialized(t *testing.T) {
	ctx := t.Context()
	uc, repo := setup(t)

	req := domain.CreateWebhookRequest{
		Name:     "test-hook",
		URL:      "https://example.com/hook",
		Events:   nil,
		Headers:  nil,
		IsActive: false,
	}

	repo.On("Create", ctx, mock.MatchedBy(func(w *domain.Webhook) bool {
		return w.Events != nil && len(w.Events) == 0 &&
			w.Headers != nil && len(w.Headers) == 0
	})).Return(nil)

	result, err := uc.Create(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotNil(t, result.Events)
	assert.Empty(t, result.Events)
	assert.NotNil(t, result.Headers)
	assert.Empty(t, result.Headers)
	repo.AssertExpectations(t)
}

func TestCreate_RepoError(t *testing.T) {
	ctx := t.Context()
	uc, repo := setup(t)

	req := domain.CreateWebhookRequest{
		Name: "fail-hook",
		URL:  "https://example.com/hook",
	}

	repo.On("Create", ctx, mock.AnythingOfType("*domain.Webhook")).
		Return(errors.New("database error"))

	result, err := uc.Create(ctx, req)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "database error")
	repo.AssertExpectations(t)
}
