package domain

import "strings"

// maxUserAgentLength bounds stored User-Agent strings to keep logs and DB columns predictable.
const maxUserAgentLength = 512

// UserAgent is a bounded-length user-agent string. It is trimmed of whitespace and truncated to 512 runes.
// Unlike IPAddress or URL, construction never fails — callers may pass arbitrary client input.
type UserAgent struct {
	raw string
}

// NewUserAgent trims whitespace and truncates the input to maxUserAgentLength runes.
// Empty input yields the zero value (IsZero == true).
func NewUserAgent(s string) UserAgent {
	s = strings.TrimSpace(s)
	if s == "" {
		return UserAgent{}
	}
	runes := []rune(s)
	if len(runes) > maxUserAgentLength {
		s = string(runes[:maxUserAgentLength])
	}
	return UserAgent{raw: s}
}

// String returns the stored user-agent string.
func (ua UserAgent) String() string { return ua.raw }

// IsZero reports whether the UserAgent is the zero value (empty after trimming).
func (ua UserAgent) IsZero() bool { return ua.raw == "" }
