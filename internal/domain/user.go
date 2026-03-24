package domain

import (
	"errors"
	"fmt"
	"gct/internal/shared/infrastructure/validation"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrPhoneRequired        = errors.New("phone is required")
	ErrPasswordRequired     = errors.New("password is required")
	ErrPasswordHashRequired = errors.New("password hash is required")
	ErrInvalidPassword      = errors.New("invalid password")
	ErrUserNotApproved      = errors.New("user is not approved by admin")
)

// NewUser creates a new User with initialized fields
func NewUser() *User {
	return &User{
		Attributes: make(map[string]any),
		Active:     true,
		DeletedAt:  0,
	}
}

// User represents a system user with RBAC and attributes for ABAC.
type User struct {
	ID           uuid.UUID      `db:"id"            json:"id"                                 example:"550e8400-e29b-41d4-a716-446655440000"`
	RoleID       *uuid.UUID     `db:"role_id"       json:"role_id,omitempty"                  example:"660e8400-e29b-41d4-a716-446655440001"`
	Username     *string        `db:"username"      json:"username,omitempty"                 validate:"omitempty,min=3"    example:"john_doe"     minLength:"3"  maxLength:"50" extensions:"x-nullable"`
	Email        *string        `db:"email"         json:"email,omitempty"                    validate:"omitempty,email"    example:"john@example.com"   format:"email" extensions:"x-nullable"`
	Phone        *string        `db:"phone"         json:"phone,omitempty"                    validate:"required,phone"     example:"+998901234567" minLength:"1"`
	PasswordHash string         `db:"password_hash" json:"-"`
	Salt         *string        `db:"salt"          json:"-"`
	Attributes   map[string]any `json:"attributes"` // JSONB used to be here, now EAV table
	Active       bool           `db:"active"        json:"active"                             example:"true"`
	IsApproved   bool           `db:"is_approved"   json:"is_approved"                        example:"true"`
	LastSeen     *time.Time     `db:"last_seen"     json:"last_seen,omitempty"                example:"2024-01-25T10:30:00Z"  format:"date-time"`
	DeletedAt    int64          `db:"deleted_at"    json:"deleted_at"                         example:"0"`
	CreatedAt    time.Time      `db:"created_at"    json:"created_at"                         example:"2024-01-01T00:00:00Z"  format:"date-time"`
	UpdatedAt    time.Time      `db:"updated_at"    json:"updated_at"                         example:"2024-01-25T10:30:00Z"  format:"date-time"`

	Password  string     `db:"-" json:"password,omitempty" validate:"omitempty,min=8" example:"SecureP@ss123" minLength:"8"` // Transient field for input
	Relations []Relation `json:"relations,omitempty"`
}

// UserFilter represents a filter for user queries.
type UserFilter struct {
	ID         *uuid.UUID `json:"id"`
	RoleID     *uuid.UUID `json:"role_id"`
	Username   *string    `json:"username"`
	Phone      *string    `json:"phone"`
	Email      *string    `json:"email"`
	Active     *bool      `json:"active"`
	IsApproved *bool      `json:"is_approved"`
}

func (f UserFilter) IsIDNull() bool {
	return f.ID == nil
}

func (f UserFilter) IsPhoneNull() bool {
	return f.Phone == nil
}

type UsersFilter struct {
	UserFilter
	Pagination *Pagination `json:"pagination"`
}

func (f UsersFilter) IsPaginationNull() bool {
	return f.Pagination == nil
}

func (f UsersFilter) IsValidLimit() bool {
	return !f.IsPaginationNull() && f.Pagination.Limit > 0 && f.Pagination.Limit <= 1000
}

func (f UsersFilter) IsValidOffset() bool {
	return !f.IsPaginationNull() && f.Pagination.Offset > 0
}

// BulkActionRequest represents a request to perform a bulk action on multiple users.
type BulkActionRequest struct {
	IDs    []string `json:"ids" binding:"required"`
	Action string   `json:"action" binding:"required"` // "deactivate" or "delete"
}

// ChangeRoleRequest represents a request to change a user's role.
type ChangeRoleRequest struct {
	Role string `json:"role" binding:"required"`
}

// Getters for User
func (u *User) GetID() uuid.UUID              { return u.ID }
func (u *User) GetRoleID() *uuid.UUID         { return u.RoleID }
func (u *User) GetUsername() *string          { return u.Username }
func (u *User) GetEmail() *string             { return u.Email }
func (u *User) GetPhone() *string             { return u.Phone }
func (u *User) GetPasswordHash() string       { return u.PasswordHash }
func (u *User) GetAttributes() map[string]any { return u.Attributes }
func (u *User) IsActive() bool                { return u.Active }
func (u *User) GetCreatedAt() time.Time       { return u.CreatedAt }
func (u *User) GetUpdatedAt() time.Time       { return u.UpdatedAt }

// Setters for User
func (u *User) SetUsername(username *string) { u.Username = username; u.UpdatedAt = time.Now() }
func (u *User) SetEmail(email *string)       { u.Email = email; u.UpdatedAt = time.Now() }
func (u *User) SetPhone(phone *string) {
	if phone != nil {
		u.Phone = phone
		u.UpdatedAt = time.Now()
	}
}

// SetPassword hashes the password and sets PasswordHash.
func (u *User) SetPassword(password string) error {
	if password == "" {
		return ErrPasswordRequired
	}
	if !validation.IsValidPassword(password) {
		return ErrInvalidPassword
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to generate password hash: %w", err)
	}
	u.PasswordHash = string(hash)
	u.Password = password
	u.UpdatedAt = time.Now()
	return nil
}

// ComparePassword compares the provided password with the stored hash.
func (u *User) ComparePassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

func (u *User) SoftDelete() {
	u.DeletedAt = time.Now().Unix()
	u.Active = false
	u.UpdatedAt = time.Now()
}

func (u *User) Restore() {
	u.DeletedAt = 0
	u.Active = true
	u.UpdatedAt = time.Now()
}

func (u *User) IsDeleted() bool {
	return u.DeletedAt > 0
}

// Getter/Setters for filters
func (f *UserFilter) GetID() *uuid.UUID { return f.ID }
