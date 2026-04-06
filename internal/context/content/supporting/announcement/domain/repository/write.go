package repository

import (
	"context"

	"gct/internal/context/content/supporting/announcement/domain/entity"
)

// AnnouncementRepository is the write-side repository for the Announcement aggregate.
// Implementations must return ErrAnnouncementNotFound from FindByID when no row matches.
// Save and Update should persist the full aggregate state including any pending domain events.
type AnnouncementRepository interface {
	Save(ctx context.Context, e *entity.Announcement) error
	FindByID(ctx context.Context, id entity.AnnouncementID) (*entity.Announcement, error)
	Update(ctx context.Context, e *entity.Announcement) error
	Delete(ctx context.Context, id entity.AnnouncementID) error
	List(ctx context.Context, filter AnnouncementFilter) ([]*entity.Announcement, int64, error)
}
