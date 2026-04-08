package http

import "github.com/google/uuid"

// CreateUserRequest is the request DTO for creating a user.
type CreateUserRequest struct {
	Phone      string            `json:"phone" binding:"required,max=20"`
	Password   string            `json:"password" binding:"required,min=8,max=128"`
	Email      *string           `json:"email,omitempty" binding:"omitempty,max=255"`
	Username   *string           `json:"username,omitempty" binding:"omitempty,max=100"`
	RoleID     *uuid.UUID        `json:"role_id,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

// UpdateUserRequest is the request DTO for updating a user.
type UpdateUserRequest struct {
	Email      *string           `json:"email,omitempty" binding:"omitempty,max=255"`
	Username   *string           `json:"username,omitempty" binding:"omitempty,max=100"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

// ChangeRoleRequest is the request DTO for changing a user's role.
type ChangeRoleRequest struct {
	RoleID uuid.UUID `json:"role_id" binding:"required"`
}

// BulkActionRequest is the request DTO for bulk user actions.
type BulkActionRequest struct {
	IDs    []uuid.UUID `json:"ids" binding:"required,max=100"`
	Action string      `json:"action" binding:"required,max=50"`
}

// SignInRequest is the request DTO for user sign-in.
type SignInRequest struct {
	Login      string `json:"login" binding:"required,max=255"`
	Password   string `json:"password" binding:"required,max=128"`
	DeviceType string `json:"device_type" binding:"required,max=20"`
}

// SignUpRequest is the request DTO for user sign-up.
type SignUpRequest struct {
	Phone    string  `json:"phone" binding:"required,max=20"`
	Password string  `json:"password" binding:"required,min=8,max=128"`
	Username *string `json:"username,omitempty" binding:"omitempty,max=100"`
	Email    *string `json:"email,omitempty" binding:"omitempty,max=255"`
}

// SignOutRequest is the request DTO for user sign-out.
type SignOutRequest struct {
	UserID    uuid.UUID `json:"user_id" binding:"required"`
	SessionID uuid.UUID `json:"session_id" binding:"required"`
}
