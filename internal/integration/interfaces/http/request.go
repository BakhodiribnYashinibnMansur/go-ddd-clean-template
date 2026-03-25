package http

// CreateRequest represents the request body for creating an integration.
type CreateRequest struct {
	Name       string         `json:"name" binding:"required"`
	Type       string         `json:"type" binding:"required"`
	APIKey     string         `json:"api_key" binding:"required"`
	WebhookURL string         `json:"webhook_url" binding:"required"`
	Enabled    bool           `json:"enabled"`
	Config     map[string]any `json:"config,omitempty"`
}

// UpdateRequest represents the request body for updating an integration.
type UpdateRequest struct {
	Name       *string         `json:"name,omitempty"`
	Type       *string         `json:"type,omitempty"`
	APIKey     *string         `json:"api_key,omitempty"`
	WebhookURL *string         `json:"webhook_url,omitempty"`
	Enabled    *bool           `json:"enabled,omitempty"`
	Config     *map[string]any `json:"config,omitempty"`
}
