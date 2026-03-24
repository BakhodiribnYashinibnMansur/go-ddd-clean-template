package errorcode_test

import (
	"context"

	"gct/internal/domain"
	repo "gct/internal/repo/persistent/postgres/errorcode"

	"github.com/stretchr/testify/mock"
)

// MockUseCase implements errorcode.UseCaseI for testing.
type MockUseCase struct {
	mock.Mock
}

func (m *MockUseCase) Create(ctx context.Context, input repo.CreateErrorCodeInput) (*domain.ErrorCode, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ErrorCode), args.Error(1)
}

func (m *MockUseCase) Update(ctx context.Context, code string, input repo.UpdateErrorCodeInput) (*domain.ErrorCode, error) {
	args := m.Called(ctx, code, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ErrorCode), args.Error(1)
}

func (m *MockUseCase) GetByCode(ctx context.Context, code string) (*domain.ErrorCode, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ErrorCode), args.Error(1)
}

func (m *MockUseCase) List(ctx context.Context) ([]*domain.ErrorCode, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.ErrorCode), args.Error(1)
}

func (m *MockUseCase) Delete(ctx context.Context, code string) error {
	args := m.Called(ctx, code)
	return args.Error(0)
}
