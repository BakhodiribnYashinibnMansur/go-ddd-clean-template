package application

import (
	"time"

	shared "gct/internal/kernel/domain"

	"github.com/google/uuid"
)

// AnnouncementView is a read-model DTO returned by query handlers.
type AnnouncementView struct {
	ID          uuid.UUID   `json:"id"`
	Title       shared.Lang `json:"title"`
	Content     shared.Lang `json:"content"`
	Published   bool        `json:"published"`
	PublishedAt *time.Time  `json:"published_at,omitempty"`
	Priority    int         `json:"priority"`
	StartDate   *time.Time  `json:"start_date,omitempty"`
	EndDate     *time.Time  `json:"end_date,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}
