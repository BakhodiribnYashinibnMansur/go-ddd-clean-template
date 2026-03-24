package announcement_test

import (
	"context"
	"testing"

	"gct/config"
	"gct/internal/domain"
	"gct/internal/usecase/announcement"
	"gct/internal/shared/infrastructure/logger"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockRepo implements announcement.Repository
type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) Create(ctx context.Context, a *domain.Announcement) error {
	args := m.Called(ctx, a)
	return args.Error(0)
}

func (m *MockRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Announcement, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Announcement), args.Error(1)
}

func (m *MockRepo) List(ctx context.Context, filter domain.AnnouncementFilter) ([]domain.Announcement, int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]domain.Announcement), args.Get(1).(int64), args.Error(2)
}

func (m *MockRepo) Update(ctx context.Context, a *domain.Announcement) error {
	args := m.Called(ctx, a)
	return args.Error(0)
}

func (m *MockRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func setup(t *testing.T) (announcement.UseCaseI, *MockRepo) {
	t.Helper()
	repo := new(MockRepo)
	log := logger.New("debug")
	cfg := &config.Config{}
	uc := announcement.New(repo, log, cfg)
	return uc, repo
}
