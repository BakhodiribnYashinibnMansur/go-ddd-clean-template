package domain_test

import (
	"testing"

	domain "gct/internal/user/domain"
)

// ---------------------------------------------------------------------------
// Fuzz: Phone validation — should never panic
// ---------------------------------------------------------------------------

func FuzzNewPhone(f *testing.F) {
	// Seed corpus
	f.Add("+998901234567")
	f.Add("")
	f.Add("+1")
	f.Add("998901234567")
	f.Add("+00000000")
	f.Add("+++++")
	f.Add("+998 90 123 45 67")
	f.Add("\x00\x01\x02")
	f.Add("+998901234567890123456789012345678901234567890")

	f.Fuzz(func(t *testing.T, raw string) {
		// Should never panic regardless of input
		phone, err := domain.NewPhone(raw)
		if err == nil {
			// Valid phone must have a value
			if phone.Value() == "" {
				t.Error("valid phone should have non-empty value")
			}
			// Must start with +
			if phone.Value()[0] != '+' {
				t.Error("valid phone must start with +")
			}
			// Min 8 chars
			if len(phone.Value()) < 8 {
				t.Error("valid phone must be at least 8 chars")
			}
		}
	})
}

// ---------------------------------------------------------------------------
// Fuzz: Email validation — should never panic
// ---------------------------------------------------------------------------

func FuzzNewEmail(f *testing.F) {
	f.Add("user@example.com")
	f.Add("")
	f.Add("noatsign")
	f.Add("@")
	f.Add("a@b")
	f.Add("user@example.com.co.uk")
	f.Add("  spaces@example.com  ")
	f.Add("\x00@\x00")
	f.Add(string(make([]byte, 10000))) // very long input

	f.Fuzz(func(t *testing.T, raw string) {
		email, err := domain.NewEmail(raw)
		if err == nil {
			if email.Value() == "" {
				t.Error("valid email should have non-empty value")
			}
		}
	})
}

// ---------------------------------------------------------------------------
// Fuzz: Password validation — should never panic
// ---------------------------------------------------------------------------

func FuzzNewPasswordFromRaw(f *testing.F) {
	f.Add("SecureP@ss1")
	f.Add("")
	f.Add("short")
	f.Add("12345678")
	f.Add(string(make([]byte, 100)))
	f.Add("パスワード12345678") // unicode

	f.Fuzz(func(t *testing.T, raw string) {
		pw, err := domain.NewPasswordFromRaw(raw)
		if err == nil {
			// Valid password must have a hash
			if pw.Hash() == "" {
				t.Error("valid password should have non-empty hash")
			}
			// Must be able to compare successfully
			if compareErr := pw.Compare(raw); compareErr != nil {
				t.Errorf("valid password compare should succeed: %v", compareErr)
			}
		}
	})
}
