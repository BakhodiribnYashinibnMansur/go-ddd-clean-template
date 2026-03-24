package sitesetting_test

import (
	"context"

	"gct/internal/domain"

	"github.com/stretchr/testify/mock"
)

// MockUseCase implements sitesetting.UseCaseI for testing.
type MockUseCase struct {
	mock.Mock
}

func (m *MockUseCase) Get(ctx context.Context, filter *domain.SiteSettingFilter) (*domain.SiteSetting, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SiteSetting), args.Error(1)
}

func (m *MockUseCase) Gets(ctx context.Context, filter *domain.SiteSettingsFilter) ([]*domain.SiteSetting, int, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int), args.Error(2)
	}
	return args.Get(0).([]*domain.SiteSetting), args.Get(1).(int), args.Error(2)
}

func (m *MockUseCase) Update(ctx context.Context, setting *domain.SiteSetting) error {
	args := m.Called(ctx, setting)
	return args.Error(0)
}

func (m *MockUseCase) UpdateByKey(ctx context.Context, key, value string) error {
	args := m.Called(ctx, key, value)
	return args.Error(0)
}

func (m *MockUseCase) GetByKey(ctx context.Context, key string) (*domain.SiteSetting, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SiteSetting), args.Error(1)
}
