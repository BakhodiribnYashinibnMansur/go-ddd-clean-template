package validation

import (
	"testing"
	"unicode"
	"unicode/utf8"
)

func FuzzGetPasswordStrength(f *testing.F) {
	// 1. Add seed corpus (interesting inputs)
	f.Add("123456")
	f.Add("password")
	f.Add("StrongP@ssw0rd")
	f.Add("")
	f.Add("Hello World")
	f.Add("      ")
	f.Add("!@#$%^&*()")

	// 2. Fuzzing implementation
	f.Fuzz(func(t *testing.T, password string) {
		// Call the target function
		strength := GetPasswordStrength(password)

		// Property 1: The returned strength must always be a valid enum value
		if !strength.IsValid() {
			t.Errorf("GetPasswordStrength(%q) returned invalid strength: %v", password, strength)
		}

		// Property 2: Basic length invariant (Byte length as per implementation)
		// If length < MinSimpleLen (6), it MUST be Weak.
		if len(password) < MinSimpleLen {
			if strength != PasswordStrengthWeak {
				t.Errorf("Password %q (len %d) should be Weak (len < %d) but got %s", password, len(password), MinSimpleLen, strength)
			}
		}

		// Property 3: Consistency with IsValidPassword
		// IsValidPassword returns true ONLY if strength is Strong
		isValid := IsValidPassword(password)
		if isValid && strength != PasswordStrengthStrong {
			t.Errorf("IsValidPassword is true but strength is %s for password %q", strength, password)
		}
		if !isValid && strength == PasswordStrengthStrong {
			t.Errorf("IsValidPassword is false but strength is Strong for password %q", password)
		}

		// Property 4: Monotonicity of sorts (Validation logic check)
		// If it's pure ASCII digits, it can never be Strong (needs Lower, Upper, Special)
		isAllDigit := true
		for _, c := range password {
			if !unicode.IsDigit(c) {
				isAllDigit = false
				break
			}
		}
		if len(password) > 0 && isAllDigit {
			if strength == PasswordStrengthStrong {
				t.Errorf("All digit password %q cannot be Strong", password)
			}
		}

		// Property 5: Panic safety (implicitly checked by Fuzzing, but good to note)
		// Go strings are valid UTF-8 sequences usually in fuzzing, but range loop handles invalid UTF-8 gracefully (replacement char).

		// Additional specific logic check derived from implementation:
		// If len >= MinStrongLen (8) and we constructed it to contain everything, it should be Strong.
		// (Hard to verify in reverse without re-implementing logic).

		// Let's verify character counting logic doesn't crash on high unicode
		_ = utf8.RuneCountInString(password)
	})
}
