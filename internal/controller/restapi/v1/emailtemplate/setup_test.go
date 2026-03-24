package emailtemplate_test

import (
	"context"

	"gct/internal/domain"

	"github.com/stretchr/testify/mock"
)

// MockUseCase implements ucemailtemplate.UseCaseI for testing.
type MockUseCase struct {
	mock.Mock
}

func (m *MockUseCase) Create(ctx context.Context, req domain.CreateEmailTemplateRequest) (*domain.EmailTemplate, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.EmailTemplate), args.Error(1)
}

func (m *MockUseCase) GetByID(ctx context.Context, id string) (*domain.EmailTemplate, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.EmailTemplate), args.Error(1)
}

func (m *MockUseCase) List(ctx context.Context, filter domain.EmailTemplateFilter) ([]domain.EmailTemplate, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]domain.EmailTemplate), args.Get(1).(int64), args.Error(2)
}

func (m *MockUseCase) Update(ctx context.Context, id string, req domain.UpdateEmailTemplateRequest) (*domain.EmailTemplate, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.EmailTemplate), args.Error(1)
}

func (m *MockUseCase) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
