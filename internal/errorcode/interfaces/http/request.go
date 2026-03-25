package http

// CreateRequest represents the request body for creating an error code.
type CreateRequest struct {
	Code       string `json:"code" binding:"required"`
	Message    string `json:"message" binding:"required"`
	HTTPStatus int    `json:"http_status" binding:"required"`
	Category   string `json:"category" binding:"required"`
	Severity   string `json:"severity" binding:"required"`
	Retryable  bool   `json:"retryable"`
	RetryAfter int    `json:"retry_after"`
	Suggestion string `json:"suggestion"`
}

// UpdateRequest represents the request body for updating an error code.
type UpdateRequest struct {
	Message    string `json:"message" binding:"required"`
	HTTPStatus int    `json:"http_status" binding:"required"`
	Category   string `json:"category" binding:"required"`
	Severity   string `json:"severity" binding:"required"`
	Retryable  bool   `json:"retryable"`
	RetryAfter int    `json:"retry_after"`
	Suggestion string `json:"suggestion"`
}
