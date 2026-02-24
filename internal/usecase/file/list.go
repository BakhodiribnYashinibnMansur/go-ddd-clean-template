package file

import (
	"context"

	"gct/internal/domain"
)

// ListFiles returns a paginated list of file_metadata records.
func (uc *UseCase) ListFiles(ctx context.Context, filter domain.FileMetadataFilter) ([]domain.FileMetadata, int64, error) {
	return uc.repo.List(ctx, filter)
}
