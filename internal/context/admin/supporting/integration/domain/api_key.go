package domain

import "errors"

// ErrInvalidAPIKey is returned when an API key fails validation.
var ErrInvalidAPIKey = errors.New("invalid api key")

// minAPIKeyLength is the minimum acceptable length for an API key.
const minAPIKeyLength = 16

// APIKey is a secret API key. Its String() representation is REDACTED
// to avoid leaking secrets into logs.
type APIKey struct {
	raw string
}

// NewAPIKey validates and creates an APIKey. Requires a minimum length of 16
// characters. The raw value is stored as-is (not hashed) since API keys often
// need to be transmitted to upstream services.
func NewAPIKey(s string) (APIKey, error) {
	if len(s) < minAPIKeyLength {
		return APIKey{}, ErrInvalidAPIKey
	}
	return APIKey{raw: s}, nil
}

// Reveal returns the raw secret key. This is an explicit unsafe accessor and
// should only be called at boundaries that must transmit the real value
// (e.g. outbound HTTP requests to the upstream provider).
func (k APIKey) Reveal() string { return k.raw }

// String returns a redacted placeholder safe for logs and error messages.
func (k APIKey) String() string { return "[REDACTED]" }

// IsZero reports whether this APIKey is the zero value.
func (k APIKey) IsZero() bool { return k.raw == "" }

// Equal reports whether two APIKeys hold the same raw secret.
func (k APIKey) Equal(other APIKey) bool { return k.raw == other.raw }
