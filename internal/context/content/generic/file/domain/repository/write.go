package repository

import (
	"context"

	"gct/internal/context/content/generic/file/domain/entity"
	shareddomain "gct/internal/kernel/domain"
)

// FileRepository is the write-side repository for the File aggregate.
// It only exposes Save because files are immutable — updates and soft-deletes are not supported.
type FileRepository interface {
	Save(ctx context.Context, q shareddomain.Querier, e *entity.File) error
}
