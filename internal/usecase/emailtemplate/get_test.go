package emailtemplate_test

import (
	"errors"
	"testing"
	"time"

	"gct/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetByID_Success(t *testing.T) {
	ctx := t.Context()
	uc, repo := setup(t)

	expected := &domain.EmailTemplate{
		ID:        "abc-123",
		Name:      "Welcome",
		Subject:   "Hello!",
		HtmlBody:  "<h1>Hi</h1>",
		TextBody:  "Hi",
		Type:      "transactional",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	repo.On("GetByID", ctx, "abc-123").Return(expected, nil)

	result, err := uc.GetByID(ctx, "abc-123")

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, expected.ID, result.ID)
	assert.Equal(t, expected.Name, result.Name)
	assert.Equal(t, expected.Subject, result.Subject)
	repo.AssertExpectations(t)
}

func TestGetByID_NotFound(t *testing.T) {
	ctx := t.Context()
	uc, repo := setup(t)

	repo.On("GetByID", ctx, "nonexistent").
		Return(nil, errors.New("not found"))

	result, err := uc.GetByID(ctx, "nonexistent")

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not found")
	repo.AssertExpectations(t)
}
