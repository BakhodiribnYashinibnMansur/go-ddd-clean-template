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
	Config     map[string]any `json:"config"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}
