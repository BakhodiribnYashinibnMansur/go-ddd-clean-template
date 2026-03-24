package emailtemplate_test

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

	filter := domain.EmailTemplateFilter{
		Limit:  10,
		Offset: 0,
	}

	now := time.Now()
	expected := []domain.EmailTemplate{
		{ID: "id-1", Name: "Template 1", Subject: "Sub 1", Type: "transactional", IsActive: true, CreatedAt: now, UpdatedAt: now},
		{ID: "id-2", Name: "Template 2", Subject: "Sub 2", Type: "marketing", IsActive: false, CreatedAt: now, UpdatedAt: now},
	}

	repo.On("List", ctx, filter).Return(expected, int64(2), nil)

	items, total, err := uc.List(ctx, filter)

	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, items, 2)
	assert.Equal(t, "Template 1", items[0].Name)
	assert.Equal(t, "Template 2", items[1].Name)
	repo.AssertExpectations(t)
}

func TestList_Empty(t *testing.T) {
	ctx := t.Context()
	uc, repo := setup(t)

	filter := domain.EmailTemplateFilter{Limit: 10}

	repo.On("List", ctx, filter).Return([]domain.EmailTemplate{}, int64(0), nil)

	items, total, err := uc.List(ctx, filter)

	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, items)
	repo.AssertExpectations(t)
}

func TestList_WithSearchFilter(t *testing.T) {
	ctx := t.Context()
	uc, repo := setup(t)

	filter := domain.EmailTemplateFilter{
		Search: "welcome",
		Limit:  10,
	}

	now := time.Now()
	expected := []domain.EmailTemplate{
		{ID: "id-1", Name: "Welcome Email", Subject: "Welcome!", Type: "transactional", IsActive: true, CreatedAt: now, UpdatedAt: now},
	}

	repo.On("List", ctx, filter).Return(expected, int64(1), nil)

	items, total, err := uc.List(ctx, filter)

	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, items, 1)
	repo.AssertExpectations(t)
}

func TestList_RepoError(t *testing.T) {
	ctx := t.Context()
	uc, repo := setup(t)

	filter := domain.EmailTemplateFilter{Limit: 10}

	repo.On("List", ctx, filter).Return([]domain.EmailTemplate{}, int64(0), errors.New("query failed"))

	_, _, err := uc.List(ctx, filter)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "query failed")
	repo.AssertExpectations(t)
}
