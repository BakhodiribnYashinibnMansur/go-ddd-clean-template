package translation_test

import (
	"context"

	"gct/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockUseCase implements translation.UseCaseI for testing.
type MockUseCase struct {
	mock.Mock
}

func (m *MockUseCase) Upsert(ctx context.Context, entityType string, entityID uuid.UUID, req domain.UpsertTranslationsRequest) error {
	args := m.Called(ctx, entityType, entityID, req)
	return args.Error(0)
}

func (m *MockUseCase) Gets(ctx context.Context, filter domain.TranslationFilter) (domain.EntityTranslations, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(domain.EntityTranslations), args.Error(1)
}

func (m *MockUseCase) Delete(ctx context.Context, filter domain.TranslationFilter) error {
	args := m.Called(ctx, filter)
	return args.Error(0)
}
