package domain

import (
	"errors"
	"net"
	"strings"
)

// ErrInvalidIPAddress indicates that a string could not be parsed as an IPv4 or IPv6 address.
var ErrInvalidIPAddress = errors.New("invalid ip address")

// IPAddress is a validated IP address (IPv4 or IPv6).
// The raw form is whatever net.ParseIP accepted (trimmed of surrounding whitespace).
type IPAddress struct {
	raw string
}

// NewIPAddress validates the input via net.ParseIP and returns an IPAddress.
// Empty input is rejected with ErrInvalidIPAddress; use IsZero on a zero value instead.
func NewIPAddress(s string) (IPAddress, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return IPAddress{}, ErrInvalidIPAddress
	}
	if net.ParseIP(s) == nil {
		return IPAddress{}, ErrInvalidIPAddress
	}
	return IPAddress{raw: s}, nil
}

// String returns the stored textual representation.
func (i IPAddress) String() string { return i.raw }

// IsZero reports whether the IPAddress is the zero value (uninitialised).
func (i IPAddress) IsZero() bool { return i.raw == "" }
