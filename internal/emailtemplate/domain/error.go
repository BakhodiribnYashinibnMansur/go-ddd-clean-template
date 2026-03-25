package domain

import shared "gct/internal/shared/domain"

var (
	ErrEmailTemplateNotFound = shared.NewDomainError("EMAIL_TEMPLATE_NOT_FOUND", "email template not found")
)
