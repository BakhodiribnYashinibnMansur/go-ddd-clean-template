package application

import (
	"time"

	"github.com/google/uuid"
)

// IntegrationView is a read-model DTO returned by query handlers.
type IntegrationView struct {
	ID         uuid.UUID      `json:"id"`
	Name       string         `json:"name"`
	Type       string         `json:"type"`
	APIKey     string         `json:"api_key"`
	WebhookURL string         `json:"webhook_url"`
	Enabled    bool           `json:"enabled"`
	Config     map[string]string `json:"config"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

// APIKeyView is a read-model DTO for API key validation results.
type APIKeyView struct {
	ID            uuid.UUID `json:"id"`
	IntegrationID uuid.UUID `json:"integration_id"`
	Key           string    `json:"key"`
	Active        bool      `json:"active"`
}
