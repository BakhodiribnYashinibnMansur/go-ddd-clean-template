package domain

import shared "gct/internal/shared/domain"

var (
	ErrAuditLogNotFound = shared.NewDomainError("AUDIT_LOG_NOT_FOUND", "audit log not found")
)
