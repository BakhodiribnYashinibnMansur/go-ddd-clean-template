package repository

import (
	"context"

	"gct/internal/context/content/generic/file/domain/entity"
)

// FileRepository is the write-side repository for the File aggregate.
// It only exposes Save because files are immutable — updates and soft-deletes are not supported.
type FileRepository interface {
	Save(ctx context.Context, e *entity.File) error
}
