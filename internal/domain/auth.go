package domain

import "github.com/google/uuid"

// SignInIn represents the input for sign in operation
type SignInIn struct {
	Login    *string    `json:"login,omitempty"    validate:"required,min=1"    example:"user1"`
	Password *string    `json:"password,omitempty" validate:"required,min=1"    example:"pass123"`
	Session  *SessionIn `json:"session"           validate:"required" binding:"required"`
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
	Phone    *string    `json:"phone,omitempty"    validate:"required,phone"        example:"+998901234567"`
	Password *string    `json:"password,omitempty" validate:"required,strong_password" example:"SecureP@ss123"`
	Username *string    `json:"username,omitempty" validate:"omitempty,min=3"     example:"john_doe"`
	Email    *string    `json:"email,omitempty"    validate:"omitempty,email"     example:"john@example.com"`
	Session  *SessionIn `json:"session"           validate:"required" binding:"required"`
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
