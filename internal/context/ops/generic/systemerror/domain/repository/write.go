package repository

import (
	"context"

	"gct/internal/context/ops/generic/systemerror/domain/entity"
)

// SystemErrorRepository is the write-side persistence contract. Note the absence of Delete —
// system errors are never deleted, only resolved, to maintain a complete audit trail.
type SystemErrorRepository interface {
	Save(ctx context.Context, entity *entity.SystemError) error
	FindByID(ctx context.Context, id entity.SystemErrorID) (*entity.SystemError, error)
	Update(ctx context.Context, entity *entity.SystemError) error
	List(ctx context.Context, filter SystemErrorFilter) ([]*entity.SystemError, int64, error)
}
