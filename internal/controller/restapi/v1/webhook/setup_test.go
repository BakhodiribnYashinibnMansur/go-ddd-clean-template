package webhook_test

import (
	"context"

	"gct/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockUseCase implements ucwebhook.UseCaseI for testing.
type MockUseCase struct {
	mock.Mock
}

func (m *MockUseCase) Create(ctx context.Context, req domain.CreateWebhookRequest) (*domain.Webhook, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Webhook), args.Error(1)
}

func (m *MockUseCase) GetByID(ctx context.Context, id uuid.UUID) (*domain.Webhook, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Webhook), args.Error(1)
}

func (m *MockUseCase) List(ctx context.Context, filter domain.WebhookFilter) ([]domain.Webhook, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]domain.Webhook), args.Get(1).(int64), args.Error(2)
}

func (m *MockUseCase) Update(ctx context.Context, id uuid.UUID, req domain.UpdateWebhookRequest) (*domain.Webhook, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Webhook), args.Error(1)
}

func (m *MockUseCase) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUseCase) Test(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
