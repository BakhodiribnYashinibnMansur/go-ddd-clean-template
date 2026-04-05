package domain

import (
	"fmt"
	"strings"
)

// HTTPMethod is a validated HTTP method enum.
type HTTPMethod string

const (
	HTTPMethodGet     HTTPMethod = "GET"
	HTTPMethodPost    HTTPMethod = "POST"
	HTTPMethodPut     HTTPMethod = "PUT"
	HTTPMethodPatch   HTTPMethod = "PATCH"
	HTTPMethodDelete  HTTPMethod = "DELETE"
	HTTPMethodHead    HTTPMethod = "HEAD"
	HTTPMethodOptions HTTPMethod = "OPTIONS"
)

// NewHTTPMethod normalizes the input to upper-case and validates it against the known set.
func NewHTTPMethod(s string) (HTTPMethod, error) {
	m := HTTPMethod(strings.ToUpper(strings.TrimSpace(s)))
	if !m.IsValid() {
		return "", fmt.Errorf("invalid HTTP method: %q", s)
	}
	return m, nil
}

// String returns the canonical upper-case method name.
func (m HTTPMethod) String() string { return string(m) }

// IsValid reports whether m is one of the recognized HTTP methods.
func (m HTTPMethod) IsValid() bool {
	switch m {
	case HTTPMethodGet, HTTPMethodPost, HTTPMethodPut, HTTPMethodPatch,
		HTTPMethodDelete, HTTPMethodHead, HTTPMethodOptions:
		return true
	}
	return false
}
