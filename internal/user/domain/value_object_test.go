package domain_test

import (
	"errors"
	"testing"

	domain "gct/internal/user/domain"
)

// ---------------------------------------------------------------------------
// Phone
// ---------------------------------------------------------------------------

func TestNewPhone_Valid(t *testing.T) {
	p, err := domain.NewPhone("+998901234567")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Value() != "+998901234567" {
		t.Fatalf("expected +998901234567, got %s", p.Value())
	}
	if p.String() != p.Value() {
		t.Fatal("String() should equal Value()")
	}
}

func TestNewPhone_Empty(t *testing.T) {
	_, err := domain.NewPhone("")
	if err == nil {
		t.Fatal("expected error for empty phone")
	}
	if !errors.Is(err, domain.ErrInvalidPhone) {
		t.Fatalf("expected ErrInvalidPhone, got %v", err)
	}
}

func TestNewPhone_MissingPlus(t *testing.T) {
	_, err := domain.NewPhone("998901234567")
	if err == nil {
		t.Fatal("expected error for missing +")
	}
}

func TestNewPhone_TooShort(t *testing.T) {
	_, err := domain.NewPhone("+12345")
	if err == nil {
		t.Fatal("expected error for short phone")
	}
}

// ---------------------------------------------------------------------------
// Email
// ---------------------------------------------------------------------------

func TestNewEmail_Valid(t *testing.T) {
	e, err := domain.NewEmail("user@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.Value() != "user@example.com" {
		t.Fatalf("expected user@example.com, got %s", e.Value())
	}
}

func TestNewEmail_Empty(t *testing.T) {
	_, err := domain.NewEmail("")
	if err == nil {
		t.Fatal("expected error for empty email")
	}
	if !errors.Is(err, domain.ErrInvalidEmail) {
		t.Fatalf("expected ErrInvalidEmail, got %v", err)
	}
}

func TestNewEmail_MissingAt(t *testing.T) {
	_, err := domain.NewEmail("userexample.com")
	if err == nil {
		t.Fatal("expected error for missing @")
	}
}

// ---------------------------------------------------------------------------
// Password
// ---------------------------------------------------------------------------

func TestNewPasswordFromRaw_Valid(t *testing.T) {
	pw, err := domain.NewPasswordFromRaw("SecureP@ss1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pw.Hash() == "" {
		t.Fatal("hash should not be empty")
	}
	if err := pw.Compare("SecureP@ss1"); err != nil {
		t.Fatalf("Compare should succeed: %v", err)
	}
}

func TestNewPasswordFromRaw_TooShort(t *testing.T) {
	_, err := domain.NewPasswordFromRaw("short")
	if err == nil {
		t.Fatal("expected error for short password")
	}
	if !errors.Is(err, domain.ErrWeakPassword) {
		t.Fatalf("expected ErrWeakPassword, got %v", err)
	}
}

func TestPassword_CompareWrong(t *testing.T) {
	pw, _ := domain.NewPasswordFromRaw("CorrectPassword1")
	err := pw.Compare("WrongPassword1")
	if err == nil {
		t.Fatal("expected error for wrong password")
	}
	if !errors.Is(err, domain.ErrInvalidPassword) {
		t.Fatalf("expected ErrInvalidPassword, got %v", err)
	}
}

func TestNewPasswordFromHash(t *testing.T) {
	pw, _ := domain.NewPasswordFromRaw("TestPassword1")
	reconstructed := domain.NewPasswordFromHash(pw.Hash())
	if reconstructed.Hash() != pw.Hash() {
		t.Fatal("reconstructed hash should match original")
	}
	if err := reconstructed.Compare("TestPassword1"); err != nil {
		t.Fatalf("Compare should succeed on reconstructed: %v", err)
	}
}
