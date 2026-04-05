package http

// CreateRequest represents the request body for creating a translation.
type CreateRequest struct {
	Key      string `json:"key" binding:"required"`
	Language string `json:"language" binding:"required"`
	Value    string `json:"value" binding:"required"`
	Group    string `json:"group" binding:"required"`
}

// UpdateRequest represents the request body for updating a translation.
type UpdateRequest struct {
	Key      *string `json:"key,omitempty"`
	Language *string `json:"language,omitempty"`
	Value    *string `json:"value,omitempty"`
	Group    *string `json:"group,omitempty"`
}
