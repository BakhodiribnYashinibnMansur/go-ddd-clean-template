package domain

import (
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// ---------------------------------------------------------------------------
// Phone
// ---------------------------------------------------------------------------

// Phone is a validated phone number value object.
type Phone struct{ value string }

// NewPhone creates a Phone after validation: not empty, starts with +, min 8 chars.
func NewPhone(raw string) (Phone, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return Phone{}, fmt.Errorf("%w: empty", ErrInvalidPhone)
	}
	if !strings.HasPrefix(raw, "+") {
		return Phone{}, fmt.Errorf("%w: must start with +", ErrInvalidPhone)
	}
	if len(raw) < 8 {
		return Phone{}, fmt.Errorf("%w: too short", ErrInvalidPhone)
	}
	return Phone{value: raw}, nil
}

func (p Phone) String() string { return p.value }
func (p Phone) Value() string  { return p.value }

// ---------------------------------------------------------------------------
// Email
// ---------------------------------------------------------------------------

// Email is a validated email address value object.
type Email struct{ value string }

// NewEmail creates an Email after validation: not empty, contains @.
func NewEmail(raw string) (Email, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return Email{}, fmt.Errorf("%w: empty", ErrInvalidEmail)
	}
	if !strings.Contains(raw, "@") {
		return Email{}, fmt.Errorf("%w: missing @", ErrInvalidEmail)
	}
	return Email{value: raw}, nil
}

func (e Email) String() string { return e.value }
func (e Email) Value() string  { return e.value }

// ---------------------------------------------------------------------------
// Password
// ---------------------------------------------------------------------------

// Password stores a bcrypt hash.
type Password struct{ hash string }

// NewPasswordFromRaw validates the raw password (min 8 chars) and hashes it with bcrypt.
func NewPasswordFromRaw(raw string) (Password, error) {
	if len(raw) < 8 {
		return Password{}, ErrWeakPassword
	}
	h, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)
	if err != nil {
		return Password{}, fmt.Errorf("password hash: %w", err)
	}
	return Password{hash: string(h)}, nil
}

// NewPasswordFromHash reconstructs a Password from an existing bcrypt hash (e.g. from DB).
func NewPasswordFromHash(hash string) Password {
	return Password{hash: hash}
}

// Hash returns the bcrypt hash string.
func (p Password) Hash() string { return p.hash }

// Compare checks whether the raw password matches the stored hash.
func (p Password) Compare(raw string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(p.hash), []byte(raw)); err != nil {
		return ErrInvalidPassword
	}
	return nil
}
