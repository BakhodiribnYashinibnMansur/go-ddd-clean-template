package validation

import "testing"

func TestPasswordStrength_IsValid(t *testing.T) {
	tests := []struct {
		name string
		s    PasswordStrength
		want bool
	}{
		{"simple", PasswordStrengthSimple, true},
		{"medium", PasswordStrengthMedium, true},
		{"strong", PasswordStrengthStrong, true},
		{"weak", PasswordStrengthWeak, true},
		{"invalid", PasswordStrength("unknown"), false},
		{"empty", PasswordStrength(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.IsValid(); got != tt.want {
				t.Errorf("PasswordStrength(%q).IsValid() = %v, want %v", tt.s, got, tt.want)
			}
		})
	}
}

func TestGetPasswordStrength(t *testing.T) {
	tests := []struct {
		name     string
		password string
		want     PasswordStrength
	}{
		// Weak: less than 6 characters
		{"empty", "", PasswordStrengthWeak},
		{"1 char", "a", PasswordStrengthWeak},
		{"5 chars", "abcde", PasswordStrengthWeak},

		// Simple: 6+ chars but fewer than 3 char types or < 8 chars
		{"6 lowercase", "abcdef", PasswordStrengthSimple},
		{"7 lowercase", "abcdefg", PasswordStrengthSimple},
		{"6 mixed no special", "Abcde1", PasswordStrengthSimple},
		{"all digits 8", "12345678", PasswordStrengthSimple},

		// Medium: 8+ chars with 3 character types
		{"upper lower digit 8", "Abcdefg1", PasswordStrengthMedium},
		{"upper lower special 8", "Abcdefg!", PasswordStrengthMedium},
		{"lower digit special 8", "abcdefg1!", PasswordStrengthMedium},

		// Strong: 8+ chars with all 4 character types
		{"all types 8", "Abcdef1!", PasswordStrengthStrong},
		{"all types long", "MyP@ssw0rd!!", PasswordStrengthStrong},
		{"all types minimal", "Ab1!abcd", PasswordStrengthStrong},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetPasswordStrength(tt.password)
			if got != tt.want {
				t.Errorf("GetPasswordStrength(%q) = %q, want %q", tt.password, got, tt.want)
			}
		})
	}
}

func TestIsValidPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{"strong password", "Abcdef1!", true},
		{"medium password", "Abcdefg1", false},
		{"simple password", "abcdef", false},
		{"weak password", "abc", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidPassword(tt.password)
			if got != tt.want {
				t.Errorf("IsValidPassword(%q) = %v, want %v", tt.password, got, tt.want)
			}
		})
	}
}
