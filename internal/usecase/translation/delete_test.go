package translation_test

import (
	"errors"
	"testing"

	"gct/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDelete_Success_AllLanguages(t *testing.T) {
	ctx := t.Context()
	uc, repo := setup(t)

	filter := domain.TranslationFilter{
		EntityType: "role",
		EntityID:   uuid.New(),
	}

	repo.On("Delete", ctx, filter).Return(nil)

	err := uc.Delete(ctx, filter)

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestDelete_Success_SingleLanguage(t *testing.T) {
	ctx := t.Context()
	uc, repo := setup(t)

	lang := "uz"
	filter := domain.TranslationFilter{
		EntityType: "role",
		EntityID:   uuid.New(),
		LangCode:   &lang,
	}

	repo.On("Delete", ctx, filter).Return(nil)

	err := uc.Delete(ctx, filter)

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestDelete_RepoError(t *testing.T) {
	ctx := t.Context()
	uc, repo := setup(t)

	filter := domain.TranslationFilter{
		EntityType: "role",
		EntityID:   uuid.New(),
	}

	repo.On("Delete", ctx, filter).Return(errors.New("delete failed"))

	err := uc.Delete(ctx, filter)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "delete failed")
	repo.AssertExpectations(t)
}
