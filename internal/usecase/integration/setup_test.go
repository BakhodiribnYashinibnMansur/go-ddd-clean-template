package integration_test

import (
	"context"
	"testing"

	"gct/config"
	"gct/internal/domain"
	"gct/internal/usecase/integration"
	"gct/internal/shared/infrastructure/logger"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockRepo implements integration.Repository
type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) CreateIntegration(ctx context.Context, i *domain.Integration) error {
	args := m.Called(ctx, i)
	return args.Error(0)
}

func (m *MockRepo) GetIntegrationByID(ctx context.Context, id uuid.UUID) (*domain.Integration, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Integration), args.Error(1)
}

func (m *MockRepo) GetIntegrationByName(ctx context.Context, name string) (*domain.Integration, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Integration), args.Error(1)
}

func (m *MockRepo) ListIntegrations(ctx context.Context, filter domain.IntegrationFilter) ([]domain.Integration, int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]domain.Integration), args.Get(1).(int64), args.Error(2)
}

func (m *MockRepo) UpdateIntegration(ctx context.Context, i *domain.Integration) error {
	args := m.Called(ctx, i)
	return args.Error(0)
}

func (m *MockRepo) DeleteIntegration(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepo) CreateAPIKey(ctx context.Context, k *domain.APIKey) error {
	args := m.Called(ctx, k)
	return args.Error(0)
}

func (m *MockRepo) GetAPIKeyByID(ctx context.Context, id uuid.UUID) (*domain.APIKey, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.APIKey), args.Error(1)
}

func (m *MockRepo) GetAPIKeyByKey(ctx context.Context, hashedKey string) (*domain.APIKey, error) {
	args := m.Called(ctx, hashedKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.APIKey), args.Error(1)
}

func (m *MockRepo) ListAPIKeysByIntegration(ctx context.Context, integrationID uuid.UUID) ([]domain.APIKey, error) {
	args := m.Called(ctx, integrationID)
	return args.Get(0).([]domain.APIKey), args.Error(1)
}

func (m *MockRepo) UpdateAPIKey(ctx context.Context, k *domain.APIKey) error {
	args := m.Called(ctx, k)
	return args.Error(0)
}

func (m *MockRepo) UpdateAPIKeyLastUsed(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepo) DeleteAPIKey(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func setup(t *testing.T) (integration.UseCaseI, *MockRepo) {
	t.Helper()
	repo := new(MockRepo)
	log := logger.New("debug")
	cfg := &config.Config{}
	uc := integration.New(repo, log, cfg)
	return uc, repo
}
