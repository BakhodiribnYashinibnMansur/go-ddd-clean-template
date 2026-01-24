package validation

import "regexp"

var (
	// Basic email regex (can be replaced with more complex one or net/mail)
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
)

const (
	MinEmailLen = 3
	MaxEmailLen = 254
)

// IsValidEmail checks if the email string is valid.
func IsValidEmail(email string) bool {
	if len(email) < MinEmailLen || len(email) > MaxEmailLen {
		return false
	}
	return emailRegex.MatchString(email)
}
