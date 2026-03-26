package domain

import shared "gct/internal/shared/domain"

// Sentinel domain errors for the Webhook bounded context.
var (
	ErrWebhookNotFound = shared.NewDomainError("WEBHOOK_NOT_FOUND", "webhook not found")
)
