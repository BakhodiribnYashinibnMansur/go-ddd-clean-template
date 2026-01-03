package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrPhoneRequired        = errors.New("phone is required")
	ErrPasswordRequired     = errors.New("password is required")
	ErrPasswordHashRequired = errors.New("password hash is required")
	ErrInvalidPassword      = errors.New("invalid password")
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
	ID           uuid.UUID      `db:"id"            json:"id"`
	RoleID       *uuid.UUID     `db:"role_id"       json:"role_id,omitempty"`
	Username     *string        `db:"username"      json:"username,omitempty"`
	Email        *string        `db:"email"         json:"email,omitempty"`
	Phone        *string        `db:"phone"         json:"phone,omitempty"`
	PasswordHash string         `db:"password_hash" json:"-"`
	Salt         *string        `db:"salt"          json:"-"`
	Attributes   map[string]any `db:"attributes"    json:"attributes"` // JSONB for ABAC (region, branch, dept)
	Active       bool           `db:"active"        json:"active"`
	LastSeen     *time.Time     `db:"last_seen"     json:"last_seen,omitempty"`
	DeletedAt    int64          `db:"deleted_at"    json:"deleted_at"`
	CreatedAt    time.Time      `db:"created_at"    json:"created_at"`
	UpdatedAt    time.Time      `db:"updated_at"    json:"updated_at"`

	Password string `db:"-"             json:"password,omitempty"` // Transient field for input
}

// UserFilter represents a filter for user queries.
type UserFilter struct {
	ID       *uuid.UUID `json:"id"`
	RoleID   *uuid.UUID `json:"role_id"`
	Username *string    `json:"username"`
	Phone    *string    `json:"phone"`
	Email    *string    `json:"email"`
	Active   *bool      `json:"active"`
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
	return !f.IsPaginationNull() && f.Pagination.Limit > 0
}

func (f UsersFilter) IsValidOffset() bool {
	return !f.IsPaginationNull() && f.Pagination.Offset > 0
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
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
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
