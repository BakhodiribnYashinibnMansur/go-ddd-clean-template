package http

// CreateRequest represents the request body for creating a rate limit rule.
type CreateRequest struct {
	Name              string `json:"name" binding:"required"`
	Rule              string `json:"rule" binding:"required"`
	RequestsPerWindow int    `json:"requests_per_window" binding:"required"`
	WindowDuration    int    `json:"window_duration" binding:"required"`
	Enabled           bool   `json:"enabled"`
}

// UpdateRequest represents the request body for updating a rate limit rule.
type UpdateRequest struct {
	Name              *string `json:"name,omitempty"`
	Rule              *string `json:"rule,omitempty"`
	RequestsPerWindow *int    `json:"requests_per_window,omitempty"`
	WindowDuration    *int    `json:"window_duration,omitempty"`
	Enabled           *bool   `json:"enabled,omitempty"`
}
