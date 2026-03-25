package http

import "time"

// CreateRequest represents the request body for creating an IP rule.
type CreateRequest struct {
	IPAddress string     `json:"ip_address" binding:"required"`
	Action    string     `json:"action" binding:"required"`
	Reason    string     `json:"reason" binding:"required"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// UpdateRequest represents the request body for updating an IP rule.
type UpdateRequest struct {
	IPAddress *string    `json:"ip_address,omitempty"`
	Action    *string    `json:"action,omitempty"`
	Reason    *string    `json:"reason,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}
