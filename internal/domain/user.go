package domain

import "time"

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
}
