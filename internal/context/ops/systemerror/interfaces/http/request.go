package http

import "github.com/google/uuid"

// CreateRequest represents the request body for creating a system error.
type CreateRequest struct {
	Code        string         `json:"code" binding:"required"`
	Message     string         `json:"message" binding:"required"`
	StackTrace  *string        `json:"stack_trace,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Severity    string         `json:"severity" binding:"required"`
	ServiceName *string        `json:"service_name,omitempty"`
	RequestID   *uuid.UUID     `json:"request_id,omitempty"`
	UserID      *uuid.UUID     `json:"user_id,omitempty"`
	IPAddress   *string        `json:"ip_address,omitempty"`
	Path        *string        `json:"path,omitempty"`
	Method      *string        `json:"method,omitempty"`
}

// ResolveRequest represents the request body for resolving a system error.
type ResolveRequest struct {
	ResolvedBy uuid.UUID `json:"resolved_by" binding:"required"`
}
