package session

import (
	"context"

	"gct/internal/domain"
)

type UseCaseI interface {
	Create(ctx context.Context, in *domain.Session) (*domain.Session, error)
	Get(ctx context.Context, in *domain.SessionFilter) (*domain.Session, error)
	Gets(ctx context.Context, in *domain.SessionsFilter) ([]*domain.Session, int, error)
	Update(ctx context.Context, in *domain.Session) error
	UpdateActivity(ctx context.Context, in *domain.SessionFilter) error
	Revoke(ctx context.Context, in *domain.SessionFilter) error
	Delete(ctx context.Context, in *domain.SessionFilter) error
}
