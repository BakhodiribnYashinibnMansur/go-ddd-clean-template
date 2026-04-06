package repository

import (
	"context"

	"gct/internal/context/admin/supporting/errorcode/domain/entity"
)

// ErrorCodeRepository is the write-side repository for the ErrorCode aggregate.
// Implementations must return ErrErrorCodeNotFound from FindByID when no row matches.
type ErrorCodeRepository interface {
	Save(ctx context.Context, e *entity.ErrorCode) error
	Update(ctx context.Context, e *entity.ErrorCode) error
	FindByID(ctx context.Context, id entity.ErrorCodeID) (*entity.ErrorCode, error)
	Delete(ctx context.Context, id entity.ErrorCodeID) error
}
