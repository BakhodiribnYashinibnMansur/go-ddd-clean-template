package domain

import (
	"time"

	shared "gct/internal/kernel/domain"

	"github.com/google/uuid"
)

// SystemError is the aggregate root for recorded system errors, acting as the structured error log.
// Each instance captures a single error occurrence with its full request context (IP, path, user).
// Resolution is a one-way transition: once resolved, the error cannot be re-opened.
type SystemError struct {
	shared.AggregateRoot
	code        string
	message     string
	stackTrace  *string
	metadata    map[string]string
	severity    string
	serviceName *string
	requestID   *uuid.UUID
	userID      *uuid.UUID
	ipAddress   *string
	path        *string
	method      *string
	isResolved  bool
	resolvedAt  *time.Time
	resolvedBy  *uuid.UUID
}

// NewSystemError creates a new SystemError aggregate with the minimum required fields.
// Optional context (stack trace, request info) should be attached via the Set* methods before persistence.
// Raises a SystemErrorRecorded event that can trigger alerting for high-severity errors.
func NewSystemError(code, message, severity string) *SystemError {
	se := &SystemError{
		AggregateRoot: shared.NewAggregateRoot(),
		code:          code,
		message:       message,
		severity:      severity,
		metadata:      make(map[string]string),
		isResolved:    false,
	}
	se.AddEvent(NewSystemErrorRecorded(se.ID(), code, severity))
	return se
}

// ReconstructSystemError rebuilds a SystemError aggregate from persisted data. No events are raised.
func ReconstructSystemError(
	id uuid.UUID,
	createdAt time.Time,
	code, message string,
	stackTrace *string,
	metadata map[string]string,
	severity string,
	serviceName *string,
	requestID *uuid.UUID,
	userID *uuid.UUID,
	ipAddress *string,
	path *string,
	method *string,
	isResolved bool,
	resolvedAt *time.Time,
	resolvedBy *uuid.UUID,
) *SystemError {
	if metadata == nil {
		metadata = make(map[string]string)
	}
	return &SystemError{
		AggregateRoot: shared.NewAggregateRootWithID(id, createdAt, createdAt, nil),
		code:          code,
		message:       message,
		stackTrace:    stackTrace,
		metadata:      metadata,
		severity:      severity,
		serviceName:   serviceName,
		requestID:     requestID,
		userID:        userID,
		ipAddress:     ipAddress,
		path:          path,
		method:        method,
		isResolved:    isResolved,
		resolvedAt:    resolvedAt,
		resolvedBy:    resolvedBy,
	}
}

// Resolve marks the system error as resolved by a specific admin user.
// This is an irreversible state transition — calling Resolve on an already-resolved error is idempotent.
func (se *SystemError) Resolve(resolvedBy uuid.UUID) {
	now := time.Now()
	se.isResolved = true
	se.resolvedAt = &now
	se.resolvedBy = &resolvedBy
	se.Touch()
	se.AddEvent(NewSystemErrorResolved(se.ID(), resolvedBy))
}

// Enrichment setters — these attach optional request context to the error before it is persisted.
// They are designed to be called in a builder pattern immediately after NewSystemError.

func (se *SystemError) SetStackTrace(st *string) { se.stackTrace = st }
func (se *SystemError) SetMetadata(m map[string]string) { se.metadata = m }
func (se *SystemError) SetServiceName(s *string) { se.serviceName = s }
func (se *SystemError) SetRequestID(id *uuid.UUID) { se.requestID = id }
func (se *SystemError) SetUserID(id *uuid.UUID) { se.userID = id }
func (se *SystemError) SetIPAddress(ip *string) { se.ipAddress = ip }
func (se *SystemError) SetPath(p *string) { se.path = p }
func (se *SystemError) SetMethod(m *string) { se.method = m }

// ---------------------------------------------------------------------------
// Getters
// ---------------------------------------------------------------------------

func (se *SystemError) Code() string          { return se.code }
func (se *SystemError) Message() string       { return se.message }
func (se *SystemError) StackTrace() *string   { return se.stackTrace }
func (se *SystemError) Metadata() map[string]string { return se.metadata }
func (se *SystemError) Severity() string      { return se.severity }
func (se *SystemError) ServiceName() *string  { return se.serviceName }
func (se *SystemError) RequestID() *uuid.UUID { return se.requestID }
func (se *SystemError) UserID() *uuid.UUID    { return se.userID }
func (se *SystemError) IPAddress() *string    { return se.ipAddress }
func (se *SystemError) Path() *string         { return se.path }
func (se *SystemError) Method() *string       { return se.method }
func (se *SystemError) IsResolved() bool      { return se.isResolved }
func (se *SystemError) ResolvedAt() *time.Time { return se.resolvedAt }
func (se *SystemError) ResolvedBy() *uuid.UUID { return se.resolvedBy }
