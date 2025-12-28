package domain

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User -.
type User struct {
	ID           int64      `json:"id" db:"id"`
	Username     *string    `json:"username,omitempty" db:"username"`
	Phone        string     `json:"phone" db:"phone"`
	PasswordHash string     `json:"-" db:"password_hash"`
	Salt         *string    `json:"-" db:"salt"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt    int        `json:"deleted_at" db:"deleted_at"`
	LastSeen     *time.Time `json:"last_seen,omitempty" db:"last_seen"`
	Password     string     `json:"-" db:"-"` // Transient field for input
}

// UserFilter represents a filter for user queries.
type UserFilter struct {
	ID    *int64  `json:"id"`
	Phone *string `json:"phone"`
}

type UsersFilter struct {
	UserFilter
	Pagination *Pagination `json:"pagination"`
}

// SetPassword hashes the password and sets PasswordHash.
func (u *User) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	return nil
}

// ComparePassword compares the provided password with the stored hash.
func (u *User) ComparePassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

func (u *User) ValidatePhone() error {
	if u.Phone == "" {
		return errors.New("phone is required")
	}
	return nil
}
