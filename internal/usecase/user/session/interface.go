package session

import (
	"context"
	"time"

	"github.com/evrone/go-clean-template/internal/domain"
	"github.com/google/uuid"
)

type UseCaseI interface {
	Create(ctx context.Context, s domain.Session, duration time.Duration) (domain.Session, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.Session, error)
	UpdateActivity(ctx context.Context, id uuid.UUID) error
	Revoke(ctx context.Context, id uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}
