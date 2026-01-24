package domain

import "github.com/google/uuid"

// SignInIn represents the input for sign in operation
type SignInIn struct {
	Login    string    `json:"login"`
	Password string    `json:"password"   validate:"required"`
	Session  SessionIn `json:"session" validate:"required"`
}

// SignInOut represents the output for sign in operation
type SignInOut struct {
	UserID       uuid.UUID `json:"user_id"`
	SessionID    uuid.UUID `json:"session_id"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
}

// SignUpIn represents the input for sign up operation
type SignUpIn struct {
	Phone    string    `json:"phone"      validate:"required,phone"`
	Password string    `json:"password"   validate:"required,strong_password"`
	Username string    `json:"username"`
	Email    string    `json:"email"      validate:"omitempty,email"`
	Session  SessionIn `json:"session"    validate:"required"`
}

// SignOutIn represents the input for sign out operation
type SignOutIn struct {
	SessionID uuid.UUID `json:"session_id" validate:"required"`
	UserID    uuid.UUID `json:"user_id"`
}

// RefreshIn represents the input for refresh token operation
type RefreshIn struct {
	SessionID uuid.UUID `json:"session_id" validate:"required"`
}

// RevokeSessionsIn represents the input for revoking sessions
type RevokeSessionsIn struct {
	UserID uuid.UUID `json:"user_id"`
}
