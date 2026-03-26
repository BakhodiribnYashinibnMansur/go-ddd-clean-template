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

// ---------------------------------------------------------------------------
// Phone: Edge cases
// ---------------------------------------------------------------------------

func TestNewPhone_WhitespaceTrimmed(t *testing.T) {
	p, err := domain.NewPhone("  +998901234567  ")
	if err != nil {
		t.Fatalf("expected whitespace to be trimmed: %v", err)
	}
	if p.Value() != "+998901234567" {
		t.Errorf("expected trimmed value, got %q", p.Value())
	}
}

func TestNewPhone_Unicode(t *testing.T) {
	_, err := domain.NewPhone("+٩٩٨٩٠١٢٣٤٥٦٧")
	// Should succeed or fail gracefully — never panic
	_ = err
}

func TestNewPhone_VeryLong(t *testing.T) {
	long := "+" + string(make([]byte, 1000))
	_, err := domain.NewPhone(long)
	// Should not panic
	_ = err
}

func TestNewPhone_TableDriven(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid UZ", "+998901234567", false},
		{"valid US", "+12025551234", false},
		{"valid short", "+1234567890", false},
		{"empty", "", true},
		{"no plus", "998901234567", true},
		{"too short", "+12345", true},
		{"only plus", "+", true},
		{"plus and spaces", "+       ", true},
		{"exactly 8 chars valid", "+1234567", false},    // 8 chars including + — passes min check
		{"7 chars with plus", "+123456", true},           // 7 chars — too short
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.NewPhone(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPhone(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Email: Edge cases
// ---------------------------------------------------------------------------

func TestNewEmail_WhitespaceTrimmed(t *testing.T) {
	e, err := domain.NewEmail("  user@example.com  ")
	if err != nil {
		t.Fatalf("expected whitespace to be trimmed: %v", err)
	}
	if e.Value() != "user@example.com" {
		t.Errorf("expected trimmed value, got %q", e.Value())
	}
}

func TestNewEmail_TableDriven(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid simple", "user@example.com", false},
		{"valid subdomain", "user@mail.example.co.uk", false},
		{"valid plus tag", "user+tag@example.com", false},
		{"empty", "", true},
		{"no at", "userexample.com", true},
		{"only at", "@", false},       // has @ so passes basic check
		{"just at", "a@b", false},     // has @ so passes basic check
		{"spaces only", "   ", true},  // trimmed to empty
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.NewEmail(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewEmail(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Password: Edge cases
// ---------------------------------------------------------------------------

func TestNewPasswordFromRaw_ExactlyMinLength(t *testing.T) {
	// 8 chars exactly — should succeed
	pw, err := domain.NewPasswordFromRaw("12345678")
	if err != nil {
		t.Fatalf("8 char password should be valid: %v", err)
	}
	if err := pw.Compare("12345678"); err != nil {
		t.Fatalf("compare should succeed: %v", err)
	}
}

func TestNewPasswordFromRaw_SevenChars(t *testing.T) {
	_, err := domain.NewPasswordFromRaw("1234567")
	if err == nil {
		t.Fatal("7 char password should fail")
	}
}

func TestNewPasswordFromRaw_UnicodePassword(t *testing.T) {
	pw, err := domain.NewPasswordFromRaw("パスワード12345678")
	if err != nil {
		t.Fatalf("unicode password should be valid: %v", err)
	}
	if err := pw.Compare("パスワード12345678"); err != nil {
		t.Fatalf("compare should succeed for unicode: %v", err)
	}
}

func TestPassword_DifferentHashesForSameInput(t *testing.T) {
	pw1, _ := domain.NewPasswordFromRaw("SamePassword1")
	pw2, _ := domain.NewPasswordFromRaw("SamePassword1")
	if pw1.Hash() == pw2.Hash() {
		t.Error("bcrypt should produce different hashes due to random salt")
	}
}
