package http

import "github.com/google/uuid"

// CreateUserRequest is the request DTO for creating a user.
type CreateUserRequest struct {
	Phone      string         `json:"phone" binding:"required"`
	Password   string         `json:"password" binding:"required"`
	Email      *string        `json:"email,omitempty"`
	Username   *string        `json:"username,omitempty"`
	RoleID     *uuid.UUID     `json:"role_id,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

// UpdateUserRequest is the request DTO for updating a user.
type UpdateUserRequest struct {
	Email      *string        `json:"email,omitempty"`
	Username   *string        `json:"username,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

// ChangeRoleRequest is the request DTO for changing a user's role.
type ChangeRoleRequest struct {
	RoleID uuid.UUID `json:"role_id" binding:"required"`
}

// BulkActionRequest is the request DTO for bulk user actions.
type BulkActionRequest struct {
	IDs    []uuid.UUID `json:"ids" binding:"required"`
	Action string      `json:"action" binding:"required"`
}

// SignInRequest is the request DTO for user sign-in.
type SignInRequest struct {
	Login      string `json:"login" binding:"required"`
	Password   string `json:"password" binding:"required"`
	DeviceType string `json:"device_type" binding:"required"`
}

// SignUpRequest is the request DTO for user sign-up.
type SignUpRequest struct {
	Phone    string  `json:"phone" binding:"required"`
	Password string  `json:"password" binding:"required"`
	Username *string `json:"username,omitempty"`
	Email    *string `json:"email,omitempty"`
}

// SignOutRequest is the request DTO for user sign-out.
type SignOutRequest struct {
	UserID    uuid.UUID `json:"user_id" binding:"required"`
	SessionID uuid.UUID `json:"session_id" binding:"required"`
}
