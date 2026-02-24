package file

import (
	"context"

	"gct/internal/domain"
)

// Repository defines the persistence interface for file_metadata operations.
type Repository interface {
	List(ctx context.Context, filter domain.FileMetadataFilter) ([]domain.FileMetadata, int64, error)
	Update(ctx context.Context, id string, req domain.UpdateFileMetadataRequest) (*domain.FileMetadata, error)
	Delete(ctx context.Context, id string) error
}

// UseCaseI is the public interface consumed by the controller layer.
type UseCaseI interface {
	ListFiles(ctx context.Context, filter domain.FileMetadataFilter) ([]domain.FileMetadata, int64, error)
	UpdateFile(ctx context.Context, id string, req domain.UpdateFileMetadataRequest) (*domain.FileMetadata, error)
	DeleteFile(ctx context.Context, id string) error
}
