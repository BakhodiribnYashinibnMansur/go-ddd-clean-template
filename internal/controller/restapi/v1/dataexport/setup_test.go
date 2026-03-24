package dataexport_test

import (
	"context"

	"gct/internal/domain"

	"github.com/stretchr/testify/mock"
)

// MockUseCase implements dataexport.UseCaseI for testing.
type MockUseCase struct {
	mock.Mock
}

func (m *MockUseCase) Create(ctx context.Context, req domain.CreateDataExportRequest, userID string) (*domain.DataExport, error) {
	args := m.Called(ctx, req, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DataExport), args.Error(1)
}

func (m *MockUseCase) List(ctx context.Context, filter domain.DataExportFilter) ([]domain.DataExport, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]domain.DataExport), args.Get(1).(int64), args.Error(2)
}

func (m *MockUseCase) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
