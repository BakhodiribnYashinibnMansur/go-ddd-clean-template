package domain

import "github.com/google/uuid"

// SignInIn represents the input for sign in operation
type SignInIn struct {
	Phone     string `json:"phone" validate:"required"`
	Password  string `json:"password" validate:"required"`
	DeviceID  string `json:"device_id"`
	IP        string `json:"ip"`
	UserAgent string `json:"user_agent"`
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
	Phone    string `json:"phone" validate:"required"`
	Password string `json:"password" validate:"required"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// SignOutIn represents the input for sign out operation
type SignOutIn struct {
	SessionID string `json:"session_id" validate:"required"`
	UserID    string `json:"user_id" validate:"required"`
}
