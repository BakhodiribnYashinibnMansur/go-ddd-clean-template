package domain

import "github.com/google/uuid"

// SignInIn represents the input for sign in operation
type SignInIn struct {
	Phone     string    `json:"phone"      validate:"required"`
	Password  string    `json:"password"   validate:"required"`
	DeviceID  uuid.UUID `json:"device_id"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
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
	Phone     string    `json:"phone"      validate:"required"`
	Password  string    `json:"password"   validate:"required"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	DeviceID  uuid.UUID `json:"device_id"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
}

// SignOutIn represents the input for sign out operation
type SignOutIn struct {
	SessionID uuid.UUID `binding:"required" json:"session_id"`
	UserID    uuid.UUID `json:"user_id"`
}

// RefreshIn represents the input for refresh token operation
type RefreshIn struct {
	RefreshToken string    `binding:"required" json:"refresh_token"`
	SessionID    uuid.UUID `binding:"required" json:"session_id"`
}

// RevokeSessionsIn represents the input for revoking sessions
type RevokeSessionsIn struct {
	UserID uuid.UUID `json:"user_id"`
}
