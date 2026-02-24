package file

import (
	"context"
)

// DeleteFile removes a file_metadata record from the database.
func (uc *UseCase) DeleteFile(ctx context.Context, id string) error {
	return uc.repo.Delete(ctx, id)
}
