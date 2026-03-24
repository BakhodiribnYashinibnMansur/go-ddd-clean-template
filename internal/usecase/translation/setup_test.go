package translation_test

import (
	"context"
	"testing"

	"gct/internal/domain"
	"gct/internal/usecase/translation"
	"gct/internal/shared/infrastructure/logger"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockRepo implements translation.Repository
type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) Upsert(ctx context.Context, entityType string, entityID uuid.UUID, langCode string, data map[string]string) error {
	args := m.Called(ctx, entityType, entityID, langCode, data)
	return args.Error(0)
}

func (m *MockRepo) Gets(ctx context.Context, filter domain.TranslationFilter) ([]*domain.Translation, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Translation), args.Error(1)
}

func (m *MockRepo) Delete(ctx context.Context, filter domain.TranslationFilter) error {
	args := m.Called(ctx, filter)
	return args.Error(0)
}

func setup(t *testing.T) (translation.UseCaseI, *MockRepo) {
	t.Helper()
	repo := new(MockRepo)
	log := logger.New("debug")
	uc := translation.New(repo, log)
	return uc, repo
}
