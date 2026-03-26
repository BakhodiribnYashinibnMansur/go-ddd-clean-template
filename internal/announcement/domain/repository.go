package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// AnnouncementFilter carries optional filtering criteria for listing announcements.
// Zero-value fields (nil pointers, zero ints) mean "no filter" — implementations should ignore them.
type AnnouncementFilter struct {
	Published *bool
	Priority  *int
	Limit     int64
	Offset    int64
}

// AnnouncementRepository is the write-side repository for the Announcement aggregate.
// Implementations must return ErrAnnouncementNotFound from FindByID when no row matches.
// Save and Update should persist the full aggregate state including any pending domain events.
type AnnouncementRepository interface {
	Save(ctx context.Context, entity *Announcement) error
	FindByID(ctx context.Context, id uuid.UUID) (*Announcement, error)
	Update(ctx context.Context, entity *Announcement) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter AnnouncementFilter) ([]*Announcement, int64, error)
}

// AnnouncementView is a read-model DTO projected from the announcement aggregate.
// It flattens the Lang value object into per-locale string fields for direct JSON serialization.
type AnnouncementView struct {
	ID          uuid.UUID  `json:"id"`
	TitleUz     string     `json:"title_uz"`
	TitleRu     string     `json:"title_ru"`
	TitleEn     string     `json:"title_en"`
	ContentUz   string     `json:"content_uz"`
	ContentRu   string     `json:"content_ru"`
	ContentEn   string     `json:"content_en"`
	Published   bool       `json:"published"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	Priority    int        `json:"priority"`
	StartDate   *time.Time `json:"start_date,omitempty"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// AnnouncementReadRepository is the read-side (CQRS query) repository.
// It returns pre-projected AnnouncementView DTOs, bypassing aggregate reconstruction for read performance.
type AnnouncementReadRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*AnnouncementView, error)
	List(ctx context.Context, filter AnnouncementFilter) ([]*AnnouncementView, int64, error)
}
