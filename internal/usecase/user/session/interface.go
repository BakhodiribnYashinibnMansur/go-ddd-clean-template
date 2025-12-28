package session

import (
	"context"

	"github.com/evrone/go-clean-template/internal/domain"
)

type UseCaseI interface {
	Create(ctx context.Context, s *domain.Session) (*domain.Session, error)
	GetByID(ctx context.Context, filter *domain.SessionFilter) (*domain.Session, error)
	UpdateActivity(ctx context.Context, filter *domain.SessionFilter) error
	Revoke(ctx context.Context, filter *domain.SessionFilter) error
	Delete(ctx context.Context, filter *domain.SessionFilter) error
}
