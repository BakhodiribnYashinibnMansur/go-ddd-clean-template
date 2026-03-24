package file_test

import (
	"context"
	"testing"

	"gct/internal/domain"
	"gct/internal/usecase/file"
	"gct/internal/shared/infrastructure/logger"

	"github.com/stretchr/testify/mock"
)

// MockRepo implements file.Repository
type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) List(ctx context.Context, filter domain.FileMetadataFilter) ([]domain.FileMetadata, int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]domain.FileMetadata), args.Get(1).(int64), args.Error(2)
}

func (m *MockRepo) Update(ctx context.Context, id string, req domain.UpdateFileMetadataRequest) (*domain.FileMetadata, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.FileMetadata), args.Error(1)
}

func (m *MockRepo) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func setup(t *testing.T) (file.UseCaseI, *MockRepo) {
	t.Helper()
	repo := new(MockRepo)
	log := logger.New("debug")
	uc := file.New(repo, log)
	return uc, repo
}
