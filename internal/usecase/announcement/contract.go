package announcement

import (
	"context"

	"gct/internal/domain"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, a *domain.Announcement) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Announcement, error)
	List(ctx context.Context, filter domain.AnnouncementFilter) ([]domain.Announcement, int64, error)
	Update(ctx context.Context, a *domain.Announcement) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type UseCaseI interface {
	Create(ctx context.Context, req domain.CreateAnnouncementRequest) (*domain.Announcement, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Announcement, error)
	List(ctx context.Context, filter domain.AnnouncementFilter) ([]domain.Announcement, int64, error)
	Update(ctx context.Context, id uuid.UUID, req domain.UpdateAnnouncementRequest) (*domain.Announcement, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Toggle(ctx context.Context, id uuid.UUID) (*domain.Announcement, error)
}
