package repository

import (
	"context"
	"time"

	"gct/internal/context/content/supporting/announcement/domain/entity"
)

// AnnouncementFilter carries optional filtering criteria for listing announcements.
// Zero-value fields (nil pointers, zero ints) mean "no filter" — implementations should ignore them.
type AnnouncementFilter struct {
	Published *bool
	Priority  *int
	Limit     int64
	Offset    int64
}

// AnnouncementView is a read-model DTO projected from the announcement aggregate.
// It flattens the Lang value object into per-locale string fields for direct JSON serialization.
type AnnouncementView struct {
	ID          entity.AnnouncementID `json:"id"`
	TitleUz     string                `json:"title_uz"`
	TitleRu     string                `json:"title_ru"`
	TitleEn     string                `json:"title_en"`
	ContentUz   string                `json:"content_uz"`
	ContentRu   string                `json:"content_ru"`
	ContentEn   string                `json:"content_en"`
	Published   bool                  `json:"published"`
	PublishedAt *time.Time            `json:"published_at,omitempty"`
	Priority    int                   `json:"priority"`
	StartDate   *time.Time            `json:"start_date,omitempty"`
	EndDate     *time.Time            `json:"end_date,omitempty"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
}

// AnnouncementReadRepository is the read-side (CQRS query) repository.
// It returns pre-projected AnnouncementView DTOs, bypassing aggregate reconstruction for read performance.
type AnnouncementReadRepository interface {
	FindByID(ctx context.Context, id entity.AnnouncementID) (*AnnouncementView, error)
	List(ctx context.Context, filter AnnouncementFilter) ([]*AnnouncementView, int64, error)
}
