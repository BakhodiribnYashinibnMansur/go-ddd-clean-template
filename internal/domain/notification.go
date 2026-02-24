package domain

import (
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID         uuid.UUID `json:"id"`
	Title      string    `json:"title"`
	Body       string    `json:"body"`
	Type       string    `json:"type"`        // info, warning, error, success
	TargetType string    `json:"target_type"` // all, admin, user
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type NotificationFilter struct {
	Search   string
	Type     string
	IsActive *bool
	Limit    int
	Offset   int
}

type CreateNotificationRequest struct {
	Title      string `json:"title" binding:"required,min=2,max=200"`
	Body       string `json:"body" binding:"required"`
	Type       string `json:"type" binding:"required,oneof=info warning error success"`
	TargetType string `json:"target_type" binding:"required,oneof=all admin user"`
	IsActive   bool   `json:"is_active"`
}

type UpdateNotificationRequest struct {
	Title      *string `json:"title" binding:"omitempty,min=2,max=200"`
	Body       *string `json:"body"`
	Type       *string `json:"type" binding:"omitempty,oneof=info warning error success"`
	TargetType *string `json:"target_type" binding:"omitempty,oneof=all admin user"`
	IsActive   *bool   `json:"is_active"`
}
