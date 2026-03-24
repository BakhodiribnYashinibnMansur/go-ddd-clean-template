package usersetting_test

import (
	"context"
	"testing"

	"gct/internal/domain"
	"gct/internal/usecase/usersetting"
	"gct/internal/shared/infrastructure/logger"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockRepo implements setting.RepoI for usecase tests.
type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) Gets(ctx context.Context, userID uuid.UUID) ([]domain.UserSetting, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.UserSetting), args.Error(1)
}

func (m *MockRepo) Upsert(ctx context.Context, s *domain.UserSetting) error {
	args := m.Called(ctx, s)
	return args.Error(0)
}

func (m *MockRepo) Delete(ctx context.Context, userID uuid.UUID, key string) error {
	args := m.Called(ctx, userID, key)
	return args.Error(0)
}

func setup(_ *testing.T) (usersetting.UseCaseI, *MockRepo) {
	repo := new(MockRepo)
	log := logger.New("debug")
	uc := usersetting.New(repo, log)
	return uc, repo
}
