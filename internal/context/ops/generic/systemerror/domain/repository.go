package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// SystemErrorFilter enables multi-dimensional querying of system errors.
// Date range filters (FromDate/ToDate) use inclusive bounds. Nil fields are ignored.
type SystemErrorFilter struct {
	Code       *string
	Severity   *string
	IsResolved *bool
	FromDate   *time.Time
	ToDate     *time.Time
	RequestID  *uuid.UUID
	UserID     *uuid.UUID
	Limit      int64
	Offset     int64
}

// SystemErrorRepository is the write-side persistence contract. Note the absence of Delete —
// system errors are never deleted, only resolved, to maintain a complete audit trail.
type SystemErrorRepository interface {
	Save(ctx context.Context, entity *SystemError) error
	FindByID(ctx context.Context, id uuid.UUID) (*SystemError, error)
	Update(ctx context.Context, entity *SystemError) error
	List(ctx context.Context, filter SystemErrorFilter) ([]*SystemError, int64, error)
}

// SystemErrorView is a flat read-model projection for API responses and admin dashboard rendering.
type SystemErrorView struct {
	ID          uuid.UUID      `json:"id"`
	Code        string         `json:"code"`
	Message     string         `json:"message"`
	StackTrace  *string        `json:"stack_trace,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Severity    string         `json:"severity"`
	ServiceName *string        `json:"service_name,omitempty"`
	RequestID   *uuid.UUID     `json:"request_id,omitempty"`
	UserID      *uuid.UUID     `json:"user_id,omitempty"`
	IPAddress   *string        `json:"ip_address,omitempty"`
	Path        *string        `json:"path,omitempty"`
	Method      *string        `json:"method,omitempty"`
	IsResolved  bool           `json:"is_resolved"`
	ResolvedAt  *time.Time     `json:"resolved_at,omitempty"`
	ResolvedBy  *uuid.UUID     `json:"resolved_by,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
}

// SystemErrorReadRepository provides read-only access optimized for listing and detail views.
type SystemErrorReadRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*SystemErrorView, error)
	List(ctx context.Context, filter SystemErrorFilter) ([]*SystemErrorView, int64, error)
}
