package translation_test

import (
	"errors"
	"testing"
	"time"

	"gct/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGets_Success(t *testing.T) {
	ctx := t.Context()
	uc, repo := setup(t)

	entityID := uuid.New()
	filter := domain.TranslationFilter{
		EntityType: "role",
		EntityID:   entityID,
	}

	now := time.Now()
	rows := []*domain.Translation{
		{ID: uuid.New(), EntityType: "role", EntityID: entityID, LangCode: "uz", Data: map[string]string{"title": "Sarlavha"}, CreatedAt: now, UpdatedAt: now},
		{ID: uuid.New(), EntityType: "role", EntityID: entityID, LangCode: "en", Data: map[string]string{"title": "Title"}, CreatedAt: now, UpdatedAt: now},
	}

	repo.On("Gets", ctx, filter).Return(rows, nil)

	result, err := uc.Gets(ctx, filter)

	require.NoError(t, err)
	require.Len(t, result, 2)
	assert.Equal(t, "Sarlavha", result["uz"]["title"])
	assert.Equal(t, "Title", result["en"]["title"])
	repo.AssertExpectations(t)
}

func TestGets_Empty(t *testing.T) {
	ctx := t.Context()
	uc, repo := setup(t)

	entityID := uuid.New()
	filter := domain.TranslationFilter{
		EntityType: "role",
		EntityID:   entityID,
	}

	repo.On("Gets", ctx, filter).Return([]*domain.Translation{}, nil)

	result, err := uc.Gets(ctx, filter)

	require.NoError(t, err)
	assert.Empty(t, result)
	repo.AssertExpectations(t)
}

func TestGets_WithLangCode(t *testing.T) {
	ctx := t.Context()
	uc, repo := setup(t)

	entityID := uuid.New()
	lang := "uz"
	filter := domain.TranslationFilter{
		EntityType: "role",
		EntityID:   entityID,
		LangCode:   &lang,
	}

	now := time.Now()
	rows := []*domain.Translation{
		{ID: uuid.New(), EntityType: "role", EntityID: entityID, LangCode: "uz", Data: map[string]string{"title": "Sarlavha"}, CreatedAt: now, UpdatedAt: now},
	}

	repo.On("Gets", ctx, filter).Return(rows, nil)

	result, err := uc.Gets(ctx, filter)

	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, "Sarlavha", result["uz"]["title"])
	repo.AssertExpectations(t)
}

func TestGets_RepoError(t *testing.T) {
	ctx := t.Context()
	uc, repo := setup(t)

	entityID := uuid.New()
	filter := domain.TranslationFilter{
		EntityType: "role",
		EntityID:   entityID,
	}

	repo.On("Gets", ctx, filter).Return(nil, errors.New("database error"))

	result, err := uc.Gets(ctx, filter)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "database error")
	repo.AssertExpectations(t)
}
