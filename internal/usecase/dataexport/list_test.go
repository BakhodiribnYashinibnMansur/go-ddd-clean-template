package dataexport_test

import (
	"errors"
	"testing"
	"time"

	"gct/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestList_Success(t *testing.T) {
	ctx := t.Context()
	uc, repo := setup(t)

	filter := domain.DataExportFilter{
		Limit:  10,
		Offset: 0,
	}

	now := time.Now()
	userID := "user-1"
	items := []domain.DataExport{
		{ID: "exp-1", Type: "users", Status: "completed", CreatedBy: &userID, CreatedAt: now, CompletedAt: &now},
		{ID: "exp-2", Type: "orders", Status: "completed", CreatedBy: &userID, CreatedAt: now, CompletedAt: &now},
	}

	repo.On("List", ctx, filter).Return(items, int64(2), nil)

	result, total, err := uc.List(ctx, filter)

	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, result, 2)
	assert.Equal(t, "exp-1", result[0].ID)
	repo.AssertExpectations(t)
}

func TestList_Empty(t *testing.T) {
	ctx := t.Context()
	uc, repo := setup(t)

	filter := domain.DataExportFilter{Limit: 10}

	repo.On("List", ctx, filter).Return([]domain.DataExport{}, int64(0), nil)

	result, total, err := uc.List(ctx, filter)

	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, result)
	repo.AssertExpectations(t)
}

func TestList_WithTypeFilter(t *testing.T) {
	ctx := t.Context()
	uc, repo := setup(t)

	filter := domain.DataExportFilter{
		Type:  "users",
		Limit: 10,
	}

	repo.On("List", ctx, filter).Return([]domain.DataExport{}, int64(0), nil)

	_, _, err := uc.List(ctx, filter)

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestList_RepoError(t *testing.T) {
	ctx := t.Context()
	uc, repo := setup(t)

	filter := domain.DataExportFilter{Limit: 10}

	repo.On("List", ctx, filter).Return([]domain.DataExport{}, int64(0), errors.New("query failed"))

	_, _, err := uc.List(ctx, filter)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "query failed")
	repo.AssertExpectations(t)
}
