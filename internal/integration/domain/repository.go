package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// IntegrationFilter carries filtering parameters for listing integrations.
type IntegrationFilter struct {
	Search  *string
	Type    *string
	Enabled *bool
	Limit   int64
	Offset  int64
}

// IntegrationView is a read-model DTO for integrations.
type IntegrationView struct {
	ID         uuid.UUID      `json:"id"`
	Name       string         `json:"name"`
	Type       string         `json:"type"`
	APIKey     string         `json:"api_key"`
	WebhookURL string         `json:"webhook_url"`
	Enabled    bool           `json:"enabled"`
	Config     map[string]any `json:"config"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

// IntegrationRepository is the write-side repository for the Integration aggregate.
type IntegrationRepository interface {
	Save(ctx context.Context, entity *Integration) error
	FindByID(ctx context.Context, id uuid.UUID) (*Integration, error)
	Update(ctx context.Context, entity *Integration) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// IntegrationReadRepository is the read-side repository returning projected views.
type IntegrationReadRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*IntegrationView, error)
	List(ctx context.Context, filter IntegrationFilter) ([]*IntegrationView, int64, error)
}
