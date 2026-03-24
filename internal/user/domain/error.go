package domain

import shared "gct/internal/shared/domain"

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
