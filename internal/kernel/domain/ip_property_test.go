package domain_test

import (
	"fmt"
	"strings"
	"testing"

	"gct/internal/kernel/domain"

	"pgregory.net/rapid"
)

func genValidIPv4(t *rapid.T) string {
	a := rapid.IntRange(0, 255).Draw(t, "octet1")
	b := rapid.IntRange(0, 255).Draw(t, "octet2")
	c := rapid.IntRange(0, 255).Draw(t, "octet3")
	d := rapid.IntRange(0, 255).Draw(t, "octet4")
	return fmt.Sprintf("%d.%d.%d.%d", a, b, c, d)
}

func TestIPAddress_Property_ValidIPv4Accepted(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		s := genValidIPv4(t)
		ip, err := domain.NewIPAddress(s)
		if err != nil {
			t.Fatalf("valid IPv4 %q rejected: %v", s, err)
		}
		if ip.IsZero() {
			t.Fatalf("valid IPv4 %q returned zero value", s)
		}
	})
}

func TestIPAddress_Property_RoundtripIdempotency(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		s := genValidIPv4(t)
		ip1, err := domain.NewIPAddress(s)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		ip2, err := domain.NewIPAddress(ip1.String())
		if err != nil {
			t.Fatalf("idempotency failed: %v", err)
		}
		if ip1.String() != ip2.String() {
			t.Fatalf("not idempotent: %q -> %q", ip1.String(), ip2.String())
		}
	})
}

func TestIPAddress_Property_NonZero(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		s := genValidIPv4(t)
		ip, err := domain.NewIPAddress(s)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ip.IsZero() {
			t.Fatalf("successfully constructed IP %q is zero", s)
		}
	})
}

func TestIPAddress_Property_WhitespaceInvariance(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		s := genValidIPv4(t)
		spaces := rapid.StringMatching(`^[ \t]{1,5}$`).Draw(t, "spaces")
		padded := spaces + s + spaces

		ip, err := domain.NewIPAddress(padded)
		if err != nil {
			t.Fatalf("padded IP %q rejected: %v", padded, err)
		}
		if ip.String() != strings.TrimSpace(padded) {
			t.Fatalf("String() = %q, want %q", ip.String(), strings.TrimSpace(padded))
		}
	})
}
