package domain

import (
	"time"

	"github.com/google/uuid"
)

type SystemError struct {
	ID          uuid.UUID      `db:"id"           json:"id"`
	Code        string         `db:"code"         json:"code"`
	Message     string         `db:"message"      json:"message"`
	StackTrace  *string        `db:"stack_trace"  json:"stack_trace,omitempty"`
	Metadata    map[string]any `db:"metadata"     json:"metadata,omitempty"`
	Severity    string         `db:"severity"     json:"severity"`
	ServiceName *string        `db:"service_name" json:"service_name,omitempty"`

	RequestID *uuid.UUID `db:"request_id" json:"request_id,omitempty"`
	UserID    *uuid.UUID `db:"user_id"    json:"user_id,omitempty"`
	IPAddress *string    `db:"ip_address" json:"ip_address,omitempty"`
	Path      *string    `db:"path"       json:"path,omitempty"`
	Method    *string    `db:"method"     json:"method,omitempty"`

	IsResolved bool       `db:"is_resolved" json:"is_resolved"`
	ResolvedAt *time.Time `db:"resolved_at" json:"resolved_at,omitempty"`
	ResolvedBy *uuid.UUID `db:"resolved_by" json:"resolved_by,omitempty"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type SystemErrorFilter struct {
	Code       *string    `json:"code,omitempty"`
	Severity   *string    `json:"severity,omitempty"`
	IsResolved *bool      `json:"is_resolved,omitempty"`
	FromDate   *time.Time `json:"from_date,omitempty"`
	ToDate     *time.Time `json:"to_date,omitempty"`
	RequestID  *uuid.UUID `json:"request_id,omitempty"`
	UserID     *uuid.UUID `json:"user_id,omitempty"`
}

type SystemErrorsFilter struct {
	SystemErrorFilter
	Pagination *Pagination `json:"pagination"`
}
