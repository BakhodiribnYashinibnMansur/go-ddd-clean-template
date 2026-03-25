package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// WebhookFilter carries filtering parameters for listing webhooks.
type WebhookFilter struct {
	Search  *string
	Enabled *bool
	Limit   int64
	Offset  int64
}

// WebhookView is a read-model DTO for webhooks.
type WebhookView struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	Secret    string    `json:"secret"`
	Events    []string  `json:"events"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// WebhookRepository is the write-side repository for the Webhook aggregate.
type WebhookRepository interface {
	Save(ctx context.Context, entity *Webhook) error
	FindByID(ctx context.Context, id uuid.UUID) (*Webhook, error)
	Update(ctx context.Context, entity *Webhook) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// WebhookReadRepository is the read-side repository returning projected views.
type WebhookReadRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*WebhookView, error)
	List(ctx context.Context, filter WebhookFilter) ([]*WebhookView, int64, error)
}
