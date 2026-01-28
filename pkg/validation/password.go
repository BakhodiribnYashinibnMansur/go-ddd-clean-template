package validation

import "unicode"

type PasswordStrength string

const (
	PasswordStrengthSimple PasswordStrength = "simple"
	PasswordStrengthMedium PasswordStrength = "medium"
	PasswordStrengthStrong PasswordStrength = "strong"
	PasswordStrengthWeak   PasswordStrength = "weak"
)

func (s PasswordStrength) IsValid() bool {
	return IsEnumValid(s, []PasswordStrength{
		PasswordStrengthSimple,
		PasswordStrengthMedium,
		PasswordStrengthStrong,
		PasswordStrengthWeak,
	})
}

const (
	MinSimpleLen = 6
	MinMediumLen = 8
	MinStrongLen = 8
)

// GetPasswordStrength evaluates the strength of a password.
func GetPasswordStrength(password string) PasswordStrength {
	var (
		length     = len(password)
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	// Strong: At least 8 chars, Upper, Lower, Number, Special
	if length >= MinStrongLen && hasUpper && hasLower && hasNumber && hasSpecial {
		return PasswordStrengthStrong
	}

	// Medium: At least 8 chars, and (Upper+Lower+Number) OR (Upper+Lower+Special) etc.
	// Basically mixing 3 types of characters
	typesCount := 0
	if hasUpper {
		typesCount++
	}
	if hasLower {
		typesCount++
	}
	if hasNumber {
		typesCount++
	}
	if hasSpecial {
		typesCount++
	}

	if length >= MinMediumLen && typesCount >= 3 {
		return PasswordStrengthMedium
	}

	// Simple: At least 6 chars, any characters
	if length >= MinSimpleLen {
		return PasswordStrengthSimple
	}

	return PasswordStrengthWeak
}

// IsValidPassword checks if the password meets the minimum required strength (Strong by default for backward compatibility or explicit check)
func IsValidPassword(password string) bool {
	return GetPasswordStrength(password) == PasswordStrengthStrong
}
