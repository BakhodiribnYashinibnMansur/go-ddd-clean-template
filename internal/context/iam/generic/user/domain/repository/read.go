package repository

import (
	"context"

	"gct/internal/context/iam/generic/user/domain/entity"
	shared "gct/internal/kernel/domain"
)

// UserReadRepository provides read-only access returning lightweight UserView projections.
// It should never be used for write operations or aggregate reconstruction.
type UserReadRepository interface {
	FindByID(ctx context.Context, id entity.UserID) (*entity.UserView, error)
	List(ctx context.Context, filter entity.UsersFilter) ([]*entity.UserView, int64, error)
	FindSessionByID(ctx context.Context, id entity.SessionID) (*shared.AuthSession, error)
	FindUserForAuth(ctx context.Context, id entity.UserID) (*shared.AuthUser, error)
}
