package domain

import shared "gct/internal/shared/domain"

// Domain errors for the emailtemplate bounded context.
var (
	// ErrEmailTemplateNotFound signals that no email template exists for the requested identifier.
	// Repository implementations must return this sentinel so the application layer can map it to HTTP 404.
	ErrEmailTemplateNotFound = shared.NewDomainError("EMAIL_TEMPLATE_NOT_FOUND", "email template not found")
)
