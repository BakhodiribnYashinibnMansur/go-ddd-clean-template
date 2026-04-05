package domain

import shared "gct/internal/platform/domain"

// Domain errors for the audit bounded context.
// Audit logs are append-only, so the only expected lookup failure is "not found."
var (
	// ErrAuditLogNotFound signals that no audit log entry exists for the requested identifier.
	// Repository implementations must return this sentinel so the application layer can
	// distinguish missing records from infrastructure errors.
	ErrAuditLogNotFound = shared.NewDomainError("AUDIT_LOG_NOT_FOUND", "audit log not found")
)
