package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// IntegrationFilter carries optional filtering parameters for listing integrations.
// Nil pointer fields are treated as "no filter" by the repository implementation.
type IntegrationFilter struct {
	Search  *string
	Type    *string
	Enabled *bool
	Limit   int64
	Offset  int64
}

// IntegrationView is a read-model projection optimized for query responses.
// Note: APIKey is included in the view — callers should mask or redact it before exposing to non-admin clients.
type IntegrationView struct {
	ID         IntegrationID     `json:"id"`
	Name       string            `json:"name"`
	Type       string            `json:"type"`
	APIKey     string            `json:"api_key"`
	WebhookURL string            `json:"webhook_url"`
	Enabled    bool              `json:"enabled"`
	Config     map[string]string `json:"config"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

// IntegrationRepository is the write-side repository for the Integration aggregate.
// Delete performs a hard delete — callers should ensure authorization before invoking.
type IntegrationRepository interface {
	Save(ctx context.Context, entity *Integration) error
	FindByID(ctx context.Context, id IntegrationID) (*Integration, error)
	Update(ctx context.Context, entity *Integration) error
	Delete(ctx context.Context, id IntegrationID) error
}

// IntegrationAPIKeyView is a read-model projection for API key validation.
type IntegrationAPIKeyView struct {
	ID            uuid.UUID
	IntegrationID IntegrationID
	Key           string
	Active        bool
}

// IntegrationReadRepository is the read-side repository returning projected views.
// Implementations must return ErrIntegrationNotFound when FindByID yields no result.
type IntegrationReadRepository interface {
	FindByID(ctx context.Context, id IntegrationID) (*IntegrationView, error)
	List(ctx context.Context, filter IntegrationFilter) ([]*IntegrationView, int64, error)
	FindByAPIKey(ctx context.Context, apiKey string) (*IntegrationAPIKeyView, error)
}
