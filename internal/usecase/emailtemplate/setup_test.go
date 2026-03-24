package emailtemplate_test

import (
	"context"
	"testing"

	"gct/config"
	"gct/internal/domain"
	"gct/internal/usecase/emailtemplate"
	"gct/internal/shared/infrastructure/logger"

	"github.com/stretchr/testify/mock"
)

// MockRepo implements emailtemplate.Repository
type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) Create(ctx context.Context, t *domain.EmailTemplate) error {
	args := m.Called(ctx, t)
	return args.Error(0)
}

func (m *MockRepo) GetByID(ctx context.Context, id string) (*domain.EmailTemplate, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.EmailTemplate), args.Error(1)
}

func (m *MockRepo) List(ctx context.Context, filter domain.EmailTemplateFilter) ([]domain.EmailTemplate, int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]domain.EmailTemplate), args.Get(1).(int64), args.Error(2)
}

func (m *MockRepo) Update(ctx context.Context, t *domain.EmailTemplate) error {
	args := m.Called(ctx, t)
	return args.Error(0)
}

func (m *MockRepo) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func setup(t *testing.T) (emailtemplate.UseCaseI, *MockRepo) {
	t.Helper()
	repo := new(MockRepo)
	log := logger.New("debug")
	cfg := &config.Config{}
	uc := emailtemplate.New(repo, log, cfg)
	return uc, repo
}
