package file

import (
	"context"

	"gct/internal/domain"
)

// UpdateFile patches a file_metadata record and returns the updated record.
func (uc *UseCase) UpdateFile(ctx context.Context, id string, req domain.UpdateFileMetadataRequest) (*domain.FileMetadata, error) {
	return uc.repo.Update(ctx, id, req)
}
