package entity

import (
	"errors"
	"strings"
)

// ErrInvalidMimeType is returned when a MIME type string fails validation.
var ErrInvalidMimeType = errors.New("invalid mime type")

// MimeType represents a validated MIME type in "type/subtype" format.
// It is an immutable value object; once created, the value cannot be changed.
type MimeType struct {
	raw string
}

// NewMimeType validates and creates a MimeType. The input must contain exactly
// one "/" separator with non-empty type and subtype portions. Parameters after
// ";" (such as charset) are preserved as part of the subtype portion.
func NewMimeType(s string) (MimeType, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return MimeType{}, ErrInvalidMimeType
	}
	parts := strings.Split(s, "/")
	if len(parts) != 2 {
		return MimeType{}, ErrInvalidMimeType
	}
	if strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return MimeType{}, ErrInvalidMimeType
	}
	return MimeType{raw: s}, nil
}

// String returns the raw MIME type representation.
func (m MimeType) String() string { return m.raw }

// Type returns the "type" portion (before the "/").
func (m MimeType) Type() string {
	idx := strings.Index(m.raw, "/")
	if idx < 0 {
		return ""
	}
	return m.raw[:idx]
}

// Subtype returns the "subtype" portion (after the "/"), including any
// parameters (e.g. "html; charset=utf-8").
func (m MimeType) Subtype() string {
	idx := strings.Index(m.raw, "/")
	if idx < 0 {
		return ""
	}
	return m.raw[idx+1:]
}

// IsZero reports whether this MimeType is the zero value.
func (m MimeType) IsZero() bool { return m.raw == "" }
