package http

// CreateRequest represents the request body for creating a webhook.
type CreateRequest struct {
	Name    string   `json:"name" binding:"required"`
	URL     string   `json:"url" binding:"required"`
	Secret  string   `json:"secret" binding:"required"`
	Events  []string `json:"events" binding:"required"`
	Enabled bool     `json:"enabled"`
}

// UpdateRequest represents the request body for updating a webhook.
type UpdateRequest struct {
	Name    *string  `json:"name,omitempty"`
	URL     *string  `json:"url,omitempty"`
	Secret  *string  `json:"secret,omitempty"`
	Events  []string `json:"events,omitempty"`
	Enabled *bool    `json:"enabled,omitempty"`
}
