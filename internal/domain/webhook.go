package domain

import (
	"time"

	"github.com/google/uuid"
)

type Webhook struct {
	ID        uuid.UUID      `json:"id"`
	Name      string         `json:"name"`
	URL       string         `json:"url"`
	Secret    string         `json:"secret"`
	Events    []string       `json:"events"`
	Headers   map[string]any `json:"headers"`
	IsActive  bool           `json:"is_active"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt *time.Time     `json:"deleted_at,omitempty"`
}

type WebhookFilter struct {
	Search   string
	IsActive *bool
	Limit    int
	Offset   int
}

type CreateWebhookRequest struct {
	Name     string         `json:"name" binding:"required,min=2,max=100"`
	URL      string         `json:"url" binding:"required,url"`
	Secret   string         `json:"secret"`
	Events   []string       `json:"events"`
	Headers  map[string]any `json:"headers"`
	IsActive bool           `json:"is_active"`
}

type UpdateWebhookRequest struct {
	Name     *string        `json:"name" binding:"omitempty,min=2,max=100"`
	URL      *string        `json:"url" binding:"omitempty,url"`
	Secret   *string        `json:"secret"`
	Events   []string       `json:"events"`
	Headers  map[string]any `json:"headers"`
	IsActive *bool          `json:"is_active"`
}
