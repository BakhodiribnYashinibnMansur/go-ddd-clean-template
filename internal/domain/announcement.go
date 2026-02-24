package domain

import (
	"time"

	"github.com/google/uuid"
)

type Announcement struct {
	ID        uuid.UUID  `json:"id"`
	Title     string     `json:"title"`
	Content   string     `json:"content"`
	Type      string     `json:"type"` // info, warning, error, success
	IsActive  bool       `json:"is_active"`
	StartsAt  *time.Time `json:"starts_at,omitempty"`
	EndsAt    *time.Time `json:"ends_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type AnnouncementFilter struct {
	Search   string
	Type     string
	IsActive *bool
	Limit    int
	Offset   int
}

type CreateAnnouncementRequest struct {
	Title    string     `json:"title" binding:"required,min=2,max=200"`
	Content  string     `json:"content" binding:"required"`
	Type     string     `json:"type" binding:"required,oneof=info warning error success"`
	IsActive bool       `json:"is_active"`
	StartsAt *time.Time `json:"starts_at"`
	EndsAt   *time.Time `json:"ends_at"`
}

type UpdateAnnouncementRequest struct {
	Title    *string    `json:"title" binding:"omitempty,min=2,max=200"`
	Content  *string    `json:"content"`
	Type     *string    `json:"type" binding:"omitempty,oneof=info warning error success"`
	IsActive *bool      `json:"is_active"`
	StartsAt *time.Time `json:"starts_at"`
	EndsAt   *time.Time `json:"ends_at"`
}
