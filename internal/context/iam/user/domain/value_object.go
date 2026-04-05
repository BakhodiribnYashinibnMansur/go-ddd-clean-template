package domain

import (
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// ---------------------------------------------------------------------------
// Phone
// ---------------------------------------------------------------------------

// Phone is an immutable value object representing an E.164-style phone number.
// Validation requires a "+" prefix and minimum 8 characters. Once created, the value cannot be changed.
type Phone struct{ value string }

// NewPhone validates and creates a Phone. Returns ErrInvalidPhone with a descriptive suffix on failure.
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

// Email is an immutable value object with minimal validation (non-empty, contains "@").
// Full RFC 5322 validation is intentionally omitted — the real check is the verification email itself.
type Email struct{ value string }

// NewEmail validates and creates an Email. Returns ErrInvalidEmail on failure.
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

// Password is a value object wrapping a bcrypt hash. The raw password is never stored.
// bcrypt.DefaultCost is used for hashing; adjust via bcrypt package if needed.
type Password struct{ hash string }

// NewPasswordFromRaw validates minimum length (8 chars) and hashes the password with bcrypt.
// Returns ErrWeakPassword if too short. The raw input is discarded after hashing.
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

// NewPasswordFromHash reconstructs a Password from an existing bcrypt hash (e.g., from DB).
// No validation is performed — the caller must ensure the hash is a valid bcrypt output.
func NewPasswordFromHash(hash string) Password {
	return Password{hash: hash}
}

// Hash returns the bcrypt hash string.
func (p Password) Hash() string { return p.hash }

// Compare checks whether the raw password matches the stored hash using bcrypt's constant-time comparison.
// Returns ErrInvalidPassword on mismatch (not the underlying bcrypt error, to avoid leaking internals).
func (p Password) Compare(raw string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(p.hash), []byte(raw)); err != nil {
		return ErrInvalidPassword
	}
	return nil
}
