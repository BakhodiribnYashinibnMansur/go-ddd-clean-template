package setting_test

import (
	"context"

	"gct/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockUseCase implements usersetting.UseCaseI for testing.
type MockUseCase struct {
	mock.Mock
}

func (m *MockUseCase) Gets(ctx context.Context, userID uuid.UUID) ([]domain.UserSetting, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.UserSetting), args.Error(1)
}

func (m *MockUseCase) Set(ctx context.Context, userID uuid.UUID, key, value string) error {
	args := m.Called(ctx, userID, key, value)
	return args.Error(0)
}

func (m *MockUseCase) Delete(ctx context.Context, userID uuid.UUID, key string) error {
	args := m.Called(ctx, userID, key)
	return args.Error(0)
}

func (m *MockUseCase) SetPasscode(ctx context.Context, userID uuid.UUID, passcode string) error {
	args := m.Called(ctx, userID, passcode)
	return args.Error(0)
}

func (m *MockUseCase) VerifyPasscode(ctx context.Context, userID uuid.UUID, passcode string) (bool, error) {
	args := m.Called(ctx, userID, passcode)
	return args.Bool(0), args.Error(1)
}

func (m *MockUseCase) RemovePasscode(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}
