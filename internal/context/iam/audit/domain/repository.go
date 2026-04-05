package domain

import (
	"context"
	"time"

	shared "gct/internal/kernel/domain"

	"github.com/google/uuid"
)

// AuditLogFilter carries optional criteria for querying audit logs.
// All pointer fields are treated as "no filter" when nil. FromDate/ToDate define an inclusive time range.
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

// EndpointHistoryFilter carries optional criteria for querying endpoint history records.
// Nil pointer fields are ignored. StatusCode filters to an exact HTTP status match.
type EndpointHistoryFilter struct {
	UserID     *uuid.UUID
	Method     *string
	Endpoint   *string
	StatusCode *int
	FromDate   *time.Time
	ToDate     *time.Time
	Pagination *shared.Pagination
}

// AuditLogRepository is the write-side repository for audit logs.
// It exposes only Save because audit logs are append-only — no update or delete is permitted.
// Implementations must dispatch any pending domain events (e.g., AuditLogCreated) after persistence.
type AuditLogRepository interface {
	Save(ctx context.Context, auditLog *AuditLog) error
}

// EndpointHistoryRepository is the write-side repository for endpoint history.
// Like AuditLogRepository, it is append-only — entries are never modified or deleted.
type EndpointHistoryRepository interface {
	Save(ctx context.Context, entry *EndpointHistory) error
}

// AuditReadRepository is the read-side (CQRS query) repository for the audit bounded context.
// It returns pre-projected view DTOs and supports paginated, filtered listing for both
// audit logs and endpoint history in a single interface.
type AuditReadRepository interface {
	ListAuditLogs(ctx context.Context, filter AuditLogFilter) ([]*AuditLogView, int64, error)
	ListEndpointHistory(ctx context.Context, filter EndpointHistoryFilter) ([]*EndpointHistoryView, int64, error)
}

// AuditLogView is a flat read-model DTO for audit log queries.
// It mirrors the AuditLog aggregate fields but is safe for direct JSON serialization without domain logic.
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
	Metadata     map[string]string `json:"metadata,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
}

// EndpointHistoryView is a flat read-model DTO for endpoint history queries.
// Latency is stored in milliseconds; StatusCode is the raw HTTP response code.
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
