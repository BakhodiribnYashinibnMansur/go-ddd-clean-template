package http

// CreateRequest represents the request body for creating a site setting.
type CreateRequest struct {
	Key         string `json:"key" binding:"required"`
	Value       string `json:"value" binding:"required"`
	Type        string `json:"type" binding:"required"`
	Description string `json:"description"`
}

// UpdateRequest represents the request body for updating a site setting.
type UpdateRequest struct {
	Key         *string `json:"key,omitempty"`
	Value       *string `json:"value,omitempty"`
	Type        *string `json:"type,omitempty"`
	Description *string `json:"description,omitempty"`
}
