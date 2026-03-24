package errorcode_test

import (
	"context"
	"testing"

	"gct/internal/domain"
	"gct/internal/usecase/errorcode"
	repo "gct/internal/repo/persistent/postgres/errorcode"
	"gct/internal/shared/infrastructure/logger"

	"github.com/stretchr/testify/mock"
)

// MockRepo implements errorcode.Repository
type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) Create(ctx context.Context, input repo.CreateErrorCodeInput) (*domain.ErrorCode, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ErrorCode), args.Error(1)
}

func (m *MockRepo) Update(ctx context.Context, code string, input repo.UpdateErrorCodeInput) (*domain.ErrorCode, error) {
	args := m.Called(ctx, code, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ErrorCode), args.Error(1)
}

func (m *MockRepo) GetByCode(ctx context.Context, code string) (*domain.ErrorCode, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ErrorCode), args.Error(1)
}

func (m *MockRepo) List(ctx context.Context) ([]*domain.ErrorCode, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.ErrorCode), args.Error(1)
}

func (m *MockRepo) Delete(ctx context.Context, code string) error {
	args := m.Called(ctx, code)
	return args.Error(0)
}

func setup(t *testing.T) (errorcode.UseCaseI, *MockRepo) {
	t.Helper()
	mockRepo := new(MockRepo)
	log := logger.New("debug")
	uc := errorcode.NewWithRepo(mockRepo, log)
	return uc, mockRepo
}
