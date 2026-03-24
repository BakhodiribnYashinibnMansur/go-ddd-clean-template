package notification_test

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

	req := domain.CreateNotificationRequest{
		Title:      "Welcome",
		Body:       "Hello, welcome aboard!",
		Type:       "info",
		TargetType: "all",
		IsActive:   true,
	}

	repo.On("Create", ctx, mock.AnythingOfType("*domain.Notification")).
		Return(nil)

	result, err := uc.Create(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotEqual(t, result.ID.String(), "00000000-0000-0000-0000-000000000000")
	assert.Equal(t, req.Title, result.Title)
	assert.Equal(t, req.Body, result.Body)
	assert.Equal(t, req.Type, result.Type)
	assert.Equal(t, req.TargetType, result.TargetType)
	assert.Equal(t, req.IsActive, result.IsActive)
	repo.AssertExpectations(t)
}

func TestCreate_RepoError(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	uc, repo := setup(t)

	req := domain.CreateNotificationRequest{
		Title:      "Fail",
		Body:       "This should fail",
		Type:       "error",
		TargetType: "user",
		IsActive:   false,
	}

	repo.On("Create", ctx, mock.AnythingOfType("*domain.Notification")).
		Return(errors.New("database error"))

	result, err := uc.Create(ctx, req)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "database error")
	repo.AssertExpectations(t)
}
