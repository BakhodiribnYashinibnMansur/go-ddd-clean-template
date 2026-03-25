package application

import (
	"time"

	"github.com/google/uuid"
)

// AuditLogView is a read-model DTO returned by query handlers.
type AuditLogView struct {
	ID           uuid.UUID      `json:"id"`
	UserID       *uuid.UUID     `json:"user_id,omitempty"`
	SessionID    *uuid.UUID     `json:"session_id,omitempty"`
	Action       string         `json:"action"`
	ResourceType *string        `json:"resource_type,omitempty"`
	ResourceID   *uuid.UUID     `json:"resource_id,omitempty"`
	Platform     *string        `json:"platform,omitempty"`
	IPAddress    *string        `json:"ip_address,omitempty"`
	UserAgent    *string        `json:"user_agent,omitempty"`
	Permission   *string        `json:"permission,omitempty"`
	PolicyID     *uuid.UUID     `json:"policy_id,omitempty"`
	Decision     *string        `json:"decision,omitempty"`
	Success      bool           `json:"success"`
	ErrorMessage *string        `json:"error_message,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
}

// EndpointHistoryView is a read-model DTO returned by query handlers.
type EndpointHistoryView struct {
	ID         uuid.UUID  `json:"id"`
	UserID     *uuid.UUID `json:"user_id,omitempty"`
	Endpoint   string     `json:"endpoint"`
	Method     string     `json:"method"`
	StatusCode int        `json:"status_code"`
	Latency    int        `json:"latency"`
	IPAddress  *string    `json:"ip_address,omitempty"`
	UserAgent  *string    `json:"user_agent,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}
