package domain

import shared "gct/internal/shared/domain"

var (
	ErrWebhookNotFound = shared.NewDomainError("WEBHOOK_NOT_FOUND", "webhook not found")
)
