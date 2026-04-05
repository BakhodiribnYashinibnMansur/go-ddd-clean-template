package query

import (
	"context"
	"gct/internal/kernel/infrastructure/logger"
	"testing"
	"time"

	"gct/internal/context/content/generic/translation/domain"

	"github.com/stretchr/testify/require"
)

func TestListTranslationsHandler_Handle(t *testing.T) {
	t.Parallel()

	now := time.Now()
	readRepo := &mockReadRepo{
		views: []*domain.TranslationView{
			{ID: domain.NewTranslationID(), Key: "k1", Language: "en", Value: "v1", Group: "g1", CreatedAt: now, UpdatedAt: now},
			{ID: domain.NewTranslationID(), Key: "k2", Language: "fr", Value: "v2", Group: "g2", CreatedAt: now, UpdatedAt: now},
		},
		total: 2,
	}

	handler := NewListTranslationsHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), ListTranslationsQuery{
		Filter: domain.TranslationFilter{Limit: 10, Offset: 0},
	})
	require.NoError(t, err)
	if result.Total != 2 {
		t.Errorf("expected total 2, got %d", result.Total)
	}
	if len(result.Translations) != 2 {
		t.Fatalf("expected 2 translations, got %d", len(result.Translations))
	}
	if result.Translations[0].Key != "k1" {
		t.Errorf("expected k1, got %s", result.Translations[0].Key)
	}
}

func TestListTranslationsHandler_Empty(t *testing.T) {
	t.Parallel()

	readRepo := &mockReadRepo{views: []*domain.TranslationView{}, total: 0}

	handler := NewListTranslationsHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), ListTranslationsQuery{
		Filter: domain.TranslationFilter{},
	})
	require.NoError(t, err)
	if result.Total != 0 {
		t.Errorf("expected total 0, got %d", result.Total)
	}
	if len(result.Translations) != 0 {
		t.Errorf("expected 0 translations, got %d", len(result.Translations))
	}
}

func TestListTranslationsHandler_WithFilters(t *testing.T) {
	t.Parallel()

	now := time.Now()
	readRepo := &mockReadRepo{
		views: []*domain.TranslationView{
			{ID: domain.NewTranslationID(), Key: "welcome", Language: "en", Value: "Welcome", Group: "auth", CreatedAt: now, UpdatedAt: now},
		},
		total: 1,
	}

	handler := NewListTranslationsHandler(readRepo, logger.Noop())
	lang := "en"
	group := "auth"

	result, err := handler.Handle(context.Background(), ListTranslationsQuery{
		Filter: domain.TranslationFilter{
			Language: &lang,
			Group:    &group,
			Limit:    10,
		},
	})
	require.NoError(t, err)
	if result.Total != 1 {
		t.Errorf("expected total 1, got %d", result.Total)
	}
}

func TestListTranslationsHandler_RepoError(t *testing.T) {
	t.Parallel()

	readRepo := &errorReadRepo{err: errRepo}
	handler := NewListTranslationsHandler(readRepo, logger.Noop())
	_, err := handler.Handle(context.Background(), ListTranslationsQuery{Filter: domain.TranslationFilter{}})
	if err == nil {
		t.Fatal("expected error from repo")
	}
}
