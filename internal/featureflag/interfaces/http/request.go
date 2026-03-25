package http

// CreateRequest represents the request body for creating a feature flag.
type CreateRequest struct {
	Name              string `json:"name" binding:"required"`
	Description       string `json:"description"`
	Enabled           bool   `json:"enabled"`
	RolloutPercentage int    `json:"rollout_percentage"`
}

// UpdateRequest represents the request body for updating a feature flag.
type UpdateRequest struct {
	Name              *string `json:"name,omitempty"`
	Description       *string `json:"description,omitempty"`
	Enabled           *bool   `json:"enabled,omitempty"`
	RolloutPercentage *int    `json:"rollout_percentage,omitempty"`
}
