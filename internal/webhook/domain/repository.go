package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// WebhookFilter carries optional filtering parameters. Search performs a LIKE match on name/URL.
type WebhookFilter struct {
	Search  *string
	Enabled *bool
	Limit   int64
	Offset  int64
}

// WebhookView is a flat read-model projection for API responses. Note that the Secret field
// is included — the presentation layer should redact or omit it in list endpoints.
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

// WebhookRepository is the write-side persistence contract for the Webhook aggregate.
// Implementations must return ErrWebhookNotFound from FindByID when no row matches.
type WebhookRepository interface {
	Save(ctx context.Context, entity *Webhook) error
	FindByID(ctx context.Context, id uuid.UUID) (*Webhook, error)
	Update(ctx context.Context, entity *Webhook) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// WebhookReadRepository provides read-only access for listing and detail views.
type WebhookReadRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*WebhookView, error)
	List(ctx context.Context, filter WebhookFilter) ([]*WebhookView, int64, error)
}
