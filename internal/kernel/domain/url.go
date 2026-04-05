package domain

import (
	"errors"
	"fmt"
	neturl "net/url"
	"strings"
)

// ErrInvalidURL indicates that a string could not be parsed as an absolute http/https URL.
var ErrInvalidURL = errors.New("invalid url")

// maxURLLength caps the input size to prevent pathological parse input.
const maxURLLength = 2048

// URL is a validated absolute URL. Only http and https schemes are permitted and a host must be present.
type URL struct {
	raw string
}

// NewURL parses and validates s. It requires an absolute URL with scheme http or https and a non-empty host.
// Inputs longer than maxURLLength (2048) are rejected.
func NewURL(s string) (URL, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return URL{}, fmt.Errorf("%w: empty", ErrInvalidURL)
	}
	if len(s) > maxURLLength {
		return URL{}, fmt.Errorf("%w: too long", ErrInvalidURL)
	}
	u, err := neturl.Parse(s)
	if err != nil {
		return URL{}, fmt.Errorf("%w: %v", ErrInvalidURL, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return URL{}, fmt.Errorf("%w: scheme must be http or https", ErrInvalidURL)
	}
	if u.Host == "" {
		return URL{}, fmt.Errorf("%w: missing host", ErrInvalidURL)
	}
	return URL{raw: s}, nil
}

// String returns the stored URL string.
func (u URL) String() string { return u.raw }

// IsZero reports whether the URL is the zero value (uninitialised).
func (u URL) IsZero() bool { return u.raw == "" }
