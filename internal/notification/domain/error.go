package domain

import shared "gct/internal/shared/domain"

var (
	ErrNotificationNotFound = shared.NewDomainError("NOTIFICATION_NOT_FOUND", "notification not found")
)
