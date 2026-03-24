package minio_test

import (
	"context"

	"gct/internal/domain"

	"github.com/stretchr/testify/mock"
)

// MockFileUseCase implements file.UseCaseI for testing.
type MockFileUseCase struct {
	mock.Mock
}

func (m *MockFileUseCase) ListFiles(ctx context.Context, filter domain.FileMetadataFilter) ([]domain.FileMetadata, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]domain.FileMetadata), args.Get(1).(int64), args.Error(2)
}

func (m *MockFileUseCase) UpdateFile(ctx context.Context, id string, req domain.UpdateFileMetadataRequest) (*domain.FileMetadata, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.FileMetadata), args.Error(1)
}

func (m *MockFileUseCase) DeleteFile(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
