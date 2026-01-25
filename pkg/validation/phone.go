package validation

import (
	"regexp"
	"strings"
	"unicode"
)

var (
	// Phone regex: adapts to common formats, primarily expecting digits
	// Adjust this regex based on your specific requirements (e.g., specific country codes)
	// This one allows optional +, spaces, dashes, and parentheses, but requires at least 7 digits.
	phoneRegex = regexp.MustCompile(`^(\+)?([\s\(\)\-]*[0-9]){7,}$`)
)

const (
	MinPhoneLen = 7
	MaxPhoneLen = 15
)

// IsValidPhone checks if the phone string is valid.
func IsValidPhone(phone string) bool {
	// custom logic: e.g. must be Uzbekistan number +998...
	// for now, generic check
	if len(phone) < MinPhoneLen || len(phone) > MaxPhoneLen {
		// Just a length check on sanitized digits could optionally be added
	}
	return phoneRegex.MatchString(phone)
}

// SanitizePhone removes all non-numeric characters from the phone string, preserving leading + if present
func SanitizePhone(phone string) string {
	if phone == EmptyString {
		return EmptyString
	}

	sb := strings.Builder{}
	if strings.HasPrefix(phone, PhonePrefix) {
		sb.WriteRune('+')
	}

	for _, ch := range phone {
		if unicode.IsDigit(ch) {
			sb.WriteRune(ch)
		}
	}
	return sb.String()
}
