package dto

import (
	"time"

	"github.com/google/uuid"
)

// SystemErrorView is a read-model DTO returned by query handlers.
type SystemErrorView struct {
	ID          uuid.UUID         `json:"id"`
	Code        string            `json:"code"`
	Message     string            `json:"message"`
	StackTrace  *string           `json:"stack_trace,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Severity    string            `json:"severity"`
	ServiceName *string           `json:"service_name,omitempty"`
	RequestID   *uuid.UUID        `json:"request_id,omitempty"`
	UserID      *uuid.UUID        `json:"user_id,omitempty"`
	IPAddress   *string           `json:"ip_address,omitempty"`
	Path        *string           `json:"path,omitempty"`
	Method      *string           `json:"method,omitempty"`
	IsResolved  bool              `json:"is_resolved"`
	ResolvedAt  *time.Time        `json:"resolved_at,omitempty"`
	ResolvedBy  *uuid.UUID        `json:"resolved_by,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
}
