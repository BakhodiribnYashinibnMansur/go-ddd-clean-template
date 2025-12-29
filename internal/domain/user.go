package domain

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrPhoneRequired        = errors.New("phone is required")
	ErrPasswordRequired     = errors.New("password is required")
	ErrPasswordHashRequired = errors.New("password hash is required")
	ErrInvalidPassword      = errors.New("invalid password")
)

// User -.
type User struct {
	ID           int64      `db:"id"            json:"id"`
	Username     *string    `db:"username"      json:"username,omitempty"`
	Phone        string     `db:"phone"         json:"phone"`
	PasswordHash string     `db:"password_hash" json:"-"`
	Salt         *string    `db:"salt"          json:"-"`
	CreatedAt    time.Time  `db:"created_at"    json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"    json:"updated_at"`
	DeletedAt    int        `db:"deleted_at"    json:"deleted_at"`
	LastSeen     *time.Time `db:"last_seen"     json:"last_seen,omitempty"`
	Password     string     `db:"-"             json:"-"` // Transient field for input
}

// UserFilter represents a filter for user queries.
type UserFilter struct {
	ID    *int64  `json:"id"`
	Phone *string `json:"phone"`
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
func (u *User) GetID() int64            { return u.ID }
func (u *User) GetUsername() *string    { return u.Username }
func (u *User) GetPhone() string        { return u.Phone }
func (u *User) GetPasswordHash() string { return u.PasswordHash }
func (u *User) GetSalt() *string        { return u.Salt }
func (u *User) GetCreatedAt() time.Time { return u.CreatedAt }
func (u *User) GetUpdatedAt() time.Time { return u.UpdatedAt }
func (u *User) GetDeletedAt() int       { return u.DeletedAt }
func (u *User) GetLastSeen() *time.Time { return u.LastSeen }

// Setters for User
func (u *User) SetUsername(username *string) { u.Username = username; u.UpdatedAt = time.Now() }

func (u *User) SetPhone(phone string) error {
	if phone == "" {
		return ErrPhoneRequired
	}
	u.Phone = phone
	u.UpdatedAt = time.Now()
	return nil
}
func (u *User) SetLastSeen(lastSeen *time.Time) { u.LastSeen = lastSeen; u.UpdatedAt = time.Now() }
func (u *User) SetDeletedAt(deletedAt int)      { u.DeletedAt = deletedAt; u.UpdatedAt = time.Now() }

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

// Utility methods for User
func (u *User) ValidatePhone() error {
	if u.Phone == "" {
		return ErrPhoneRequired
	}
	return nil
}

func (u *User) Validate() error {
	if u.Phone == "" {
		return ErrPhoneRequired
	}
	if u.PasswordHash == "" {
		return ErrPasswordHashRequired
	}
	return nil
}

func (u *User) SoftDelete() {
	u.DeletedAt = int(time.Now().Unix())
	u.UpdatedAt = time.Now()
}

func (u *User) Restore() {
	u.DeletedAt = 0
	u.UpdatedAt = time.Now()
}

func (u *User) IsDeleted() bool {
	return u.DeletedAt > 0
}

func (u *User) UpdateTimestamp() {
	u.UpdatedAt = time.Now()
}

// Getters and Setters for UserFilter
func (f *UserFilter) GetID() *int64          { return f.ID }
func (f *UserFilter) SetID(id *int64)        { f.ID = id }
func (f *UserFilter) GetPhone() *string      { return f.Phone }
func (f *UserFilter) SetPhone(phone *string) { f.Phone = phone }
func (f *UserFilter) HasFilters() bool       { return f.ID != nil || f.Phone != nil }

// Getters and Setters for UsersFilter
func (f *UsersFilter) GetUserFilter() UserFilter            { return f.UserFilter }
func (f *UsersFilter) SetUserFilter(userFilter UserFilter)  { f.UserFilter = userFilter }
func (f *UsersFilter) GetPagination() *Pagination           { return f.Pagination }
func (f *UsersFilter) SetPagination(pagination *Pagination) { f.Pagination = pagination }

type SignInIn struct {
	Phone     string `json:"phone"      validate:"required"`
	Password  string `json:"password"   validate:"required"`
	DeviceID  string `json:"device_id"`
	UserAgent string `json:"user_agent"`
	IP        string `json:"ip"`
}

type SignInOut struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type SignUpIn struct {
	Username string `json:"username"`
	Phone    string `json:"phone"    validate:"required"`
	Password string `json:"password" validate:"required"`
}

type SignOutIn struct {
	SessionID string `json:"session_id" validate:"required"`
}
