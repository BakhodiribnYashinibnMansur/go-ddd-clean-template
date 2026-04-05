package domain

import shared "gct/internal/platform/domain"

// Sentinel domain errors for the User bounded context.
// The presentation layer maps these codes to HTTP status codes:
//   - NOT_FOUND errors -> 404
//   - INACTIVE / NOT_APPROVED -> 403
//   - INVALID_PASSWORD / WEAK_PASSWORD -> 400
//   - MAX_SESSIONS -> 409
var (
	ErrUserNotFound       = shared.NewDomainError("USER_NOT_FOUND", "user not found")
	ErrPhoneExists        = shared.NewDomainError("USER_PHONE_EXISTS", "phone already registered")
	ErrInvalidPassword    = shared.NewDomainError("USER_INVALID_PASSWORD", "invalid password")
	ErrUserInactive       = shared.NewDomainError("USER_INACTIVE", "user is inactive")
	ErrUserNotApproved    = shared.NewDomainError("USER_NOT_APPROVED", "user not approved")
	ErrMaxSessionsReached = shared.NewDomainError("USER_MAX_SESSIONS", "maximum sessions reached")
	ErrSessionNotFound    = shared.NewDomainError("USER_SESSION_NOT_FOUND", "session not found")
	ErrWeakPassword       = shared.NewDomainError("USER_WEAK_PASSWORD", "password must be at least 8 characters")
	ErrInvalidPhone       = shared.NewDomainError("USER_INVALID_PHONE", "invalid phone number")
	ErrInvalidEmail       = shared.NewDomainError("USER_INVALID_EMAIL", "invalid email address")
)
