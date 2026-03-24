package translation_test

import (
	"errors"
	"testing"

	"gct/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpsert_Success_SingleLanguage(t *testing.T) {
	ctx := t.Context()
	uc, repo := setup(t)

	entityType := "role"
	entityID := uuid.New()
	req := domain.UpsertTranslationsRequest{
		"uz": {"title": "Sarlavha", "description": "Tavsif"},
	}

	repo.On("Upsert", ctx, entityType, entityID, "uz", req["uz"]).Return(nil)

	err := uc.Upsert(ctx, entityType, entityID, req)

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestUpsert_Success_MultipleLanguages(t *testing.T) {
	ctx := t.Context()
	uc, repo := setup(t)

	entityType := "permission"
	entityID := uuid.New()
	req := domain.UpsertTranslationsRequest{
		"uz": {"title": "Sarlavha"},
		"en": {"title": "Title"},
	}

	repo.On("Upsert", ctx, entityType, entityID, "uz", req["uz"]).Return(nil)
	repo.On("Upsert", ctx, entityType, entityID, "en", req["en"]).Return(nil)

	err := uc.Upsert(ctx, entityType, entityID, req)

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestUpsert_EmptyRequest(t *testing.T) {
	ctx := t.Context()
	uc, repo := setup(t)

	entityType := "role"
	entityID := uuid.New()
	req := domain.UpsertTranslationsRequest{}

	err := uc.Upsert(ctx, entityType, entityID, req)

	require.NoError(t, err)
	repo.AssertNotCalled(t, "Upsert")
}

func TestUpsert_RepoError(t *testing.T) {
	ctx := t.Context()
	uc, repo := setup(t)

	entityType := "role"
	entityID := uuid.New()
	req := domain.UpsertTranslationsRequest{
		"uz": {"title": "Sarlavha"},
	}

	repo.On("Upsert", ctx, entityType, entityID, "uz", req["uz"]).
		Return(errors.New("database error"))

	err := uc.Upsert(ctx, entityType, entityID, req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "translation upsert [uz]")
	assert.Contains(t, err.Error(), "database error")
	repo.AssertExpectations(t)
}
