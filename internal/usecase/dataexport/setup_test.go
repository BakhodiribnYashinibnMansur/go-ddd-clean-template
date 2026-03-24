package dataexport_test

import (
	"context"
	"testing"

	"gct/config"
	"gct/internal/domain"
	"gct/internal/usecase/dataexport"
	"gct/internal/shared/infrastructure/logger"

	"github.com/stretchr/testify/mock"
)

// MockRepo implements dataexport.Repository
type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) Create(ctx context.Context, e *domain.DataExport) error {
	args := m.Called(ctx, e)
	return args.Error(0)
}

func (m *MockRepo) List(ctx context.Context, filter domain.DataExportFilter) ([]domain.DataExport, int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]domain.DataExport), args.Get(1).(int64), args.Error(2)
}

func (m *MockRepo) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func setup(t *testing.T) (dataexport.UseCaseI, *MockRepo) {
	t.Helper()
	repo := new(MockRepo)
	log := logger.New("debug")
	cfg := &config.Config{}
	uc := dataexport.New(repo, log, cfg)
	return uc, repo
}
