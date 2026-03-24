package dataexport_test

import (
	"encoding/json"
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

	req := domain.CreateDataExportRequest{
		Type:    "users",
		Filters: json.RawMessage(`{"status":"active"}`),
	}
	userID := "user-123"

	repo.On("Create", ctx, mock.AnythingOfType("*domain.DataExport")).Return(nil)

	result, err := uc.Create(ctx, req, userID)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotEmpty(t, result.ID)
	assert.Equal(t, "users", result.Type)
	assert.Equal(t, "completed", result.Status)
	assert.Equal(t, &userID, result.CreatedBy)
	assert.NotNil(t, result.CompletedAt)
	assert.JSONEq(t, `{"status":"active"}`, string(result.Filters))
	repo.AssertExpectations(t)
}

func TestCreate_NilFilters_DefaultsToEmptyJSON(t *testing.T) {
	ctx := t.Context()
	uc, repo := setup(t)

	req := domain.CreateDataExportRequest{
		Type:    "orders",
		Filters: nil,
	}
	userID := "user-456"

	repo.On("Create", ctx, mock.MatchedBy(func(e *domain.DataExport) bool {
		return string(e.Filters) == "{}"
	})).Return(nil)

	result, err := uc.Create(ctx, req, userID)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "{}", string(result.Filters))
	repo.AssertExpectations(t)
}

func TestCreate_RepoError(t *testing.T) {
	ctx := t.Context()
	uc, repo := setup(t)

	req := domain.CreateDataExportRequest{
		Type: "users",
	}
	userID := "user-789"

	repo.On("Create", ctx, mock.AnythingOfType("*domain.DataExport")).
		Return(errors.New("insert failed"))

	result, err := uc.Create(ctx, req, userID)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "insert failed")
	repo.AssertExpectations(t)
}
