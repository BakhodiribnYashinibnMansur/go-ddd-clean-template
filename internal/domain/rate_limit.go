package domain

import (
	"time"

	"github.com/google/uuid"
)

type RateLimit struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	PathPattern   string    `json:"path_pattern"`
	Method        string    `json:"method"`
	LimitCount    int       `json:"limit_count"`
	WindowSeconds int       `json:"window_seconds"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type RateLimitFilter struct {
	Search   string
	IsActive *bool
	Limit    int
	Offset   int
}

type CreateRateLimitRequest struct {
	Name          string `json:"name" binding:"required,min=2,max=100"`
	PathPattern   string `json:"path_pattern" binding:"required"`
	Method        string `json:"method" binding:"required"`
	LimitCount    int    `json:"limit_count" binding:"required,min=1"`
	WindowSeconds int    `json:"window_seconds" binding:"required,min=1"`
	IsActive      bool   `json:"is_active"`
}

type UpdateRateLimitRequest struct {
	Name          *string `json:"name" binding:"omitempty,min=2,max=100"`
	PathPattern   *string `json:"path_pattern"`
	Method        *string `json:"method"`
	LimitCount    *int    `json:"limit_count" binding:"omitempty,min=1"`
	WindowSeconds *int    `json:"window_seconds" binding:"omitempty,min=1"`
	IsActive      *bool   `json:"is_active"`
}
