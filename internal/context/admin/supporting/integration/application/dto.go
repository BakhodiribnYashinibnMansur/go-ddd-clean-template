package application

import (
	"time"

	"gct/internal/context/admin/supporting/integration/domain"

	"github.com/google/uuid"
)

// IntegrationView is a read-model DTO returned by query handlers.
type IntegrationView struct {
	ID         domain.IntegrationID `json:"id"`
	Name       string               `json:"name"`
	Type       string               `json:"type"`
	APIKey     string               `json:"api_key"`
	WebhookURL string               `json:"webhook_url"`
	Enabled    bool                 `json:"enabled"`
	Config     map[string]string    `json:"config"`
	CreatedAt  time.Time            `json:"created_at"`
	UpdatedAt  time.Time            `json:"updated_at"`
}

// APIKeyView is a read-model DTO for API key validation results.
type APIKeyView struct {
	ID            uuid.UUID            `json:"id"`
	IntegrationID domain.IntegrationID `json:"integration_id"`
	Key           string               `json:"key"`
	Active        bool                 `json:"active"`
}
