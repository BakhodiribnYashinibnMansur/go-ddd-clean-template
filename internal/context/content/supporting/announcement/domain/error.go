package domain

import shared "gct/internal/kernel/domain"

// Domain errors for the announcement bounded context.
// These are returned by aggregate methods and repository lookups — callers should
// match on these sentinels to distinguish recoverable domain violations from infrastructure failures.
var (
	// ErrAnnouncementNotFound signals that no announcement exists for the requested identifier.
	// Repository implementations must return this (not a generic SQL "no rows") so the application
	// layer can map it to an appropriate HTTP 404 response.
	ErrAnnouncementNotFound = shared.NewDomainError("ANNOUNCEMENT_NOT_FOUND", "announcement not found")

	// ErrAlreadyPublished prevents duplicate publish transitions.
	// Publishing is a one-way state change — once published, an announcement cannot revert to draft.
	ErrAlreadyPublished = shared.NewDomainError("ANNOUNCEMENT_ALREADY_PUBLISHED", "announcement is already published")
)
