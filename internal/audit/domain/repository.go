package domain

import (
	"context"
	"time"

	shared "gct/internal/shared/domain"

	"github.com/google/uuid"
)

// AuditLogFilter carries filtering parameters for listing audit logs.
type AuditLogFilter struct {
	UserID       *uuid.UUID
	Action       *AuditAction
	ResourceType *string
	ResourceID   *uuid.UUID
	Success      *bool
	FromDate     *time.Time
	ToDate       *time.Time
	Pagination   *shared.Pagination
}

// EndpointHistoryFilter carries filtering parameters for listing endpoint history.
type EndpointHistoryFilter struct {
	UserID     *uuid.UUID
	Method     *string
	Endpoint   *string
	StatusCode *int
	FromDate   *time.Time
	ToDate     *time.Time
	Pagination *shared.Pagination
}

// AuditLogRepository is the write-side repository for audit logs (immutable — Save only).
type AuditLogRepository interface {
	Save(ctx context.Context, auditLog *AuditLog) error
}

// EndpointHistoryRepository is the write-side repository for endpoint history (immutable — Save only).
type EndpointHistoryRepository interface {
	Save(ctx context.Context, entry *EndpointHistory) error
}

// AuditReadRepository is the read-side repository for audit queries.
type AuditReadRepository interface {
	ListAuditLogs(ctx context.Context, filter AuditLogFilter) ([]*AuditLogView, int64, error)
	ListEndpointHistory(ctx context.Context, filter EndpointHistoryFilter) ([]*EndpointHistoryView, int64, error)
}

// AuditLogView is a read-model DTO for audit log queries.
type AuditLogView struct {
	ID           uuid.UUID      `json:"id"`
	UserID       *uuid.UUID     `json:"user_id,omitempty"`
	SessionID    *uuid.UUID     `json:"session_id,omitempty"`
	Action       AuditAction    `json:"action"`
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

// EndpointHistoryView is a read-model DTO for endpoint history queries.
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
