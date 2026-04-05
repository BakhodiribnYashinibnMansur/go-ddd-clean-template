package domain

import shared "gct/internal/kernel/domain"

// Domain errors for the notification bounded context.
// Returned by repositories when the requested notification does not exist in the data store.
var (
	ErrNotificationNotFound = shared.NewDomainError("NOTIFICATION_NOT_FOUND", "notification not found")
)
