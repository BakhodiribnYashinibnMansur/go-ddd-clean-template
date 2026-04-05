package domain_test

import (
	"errors"
	"testing"

	"gct/internal/kernel/domain"
)

func TestNewIPAddress(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "valid ipv4", input: "192.168.1.1", want: "192.168.1.1"},
		{name: "valid ipv4 loopback", input: "127.0.0.1", want: "127.0.0.1"},
		{name: "valid ipv6", input: "2001:db8::1", want: "2001:db8::1"},
		{name: "valid ipv6 loopback", input: "::1", want: "::1"},
		{name: "trimmed whitespace", input: "  10.0.0.1  ", want: "10.0.0.1"},
		{name: "invalid string", input: "not-an-ip", wantErr: true},
		{name: "invalid octet", input: "999.999.999.999", wantErr: true},
		{name: "empty", input: "", wantErr: true},
		{name: "only whitespace", input: "   ", wantErr: true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := domain.NewIPAddress(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if !errors.Is(err, domain.ErrInvalidIPAddress) {
					t.Fatalf("expected ErrInvalidIPAddress, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.String() != tc.want {
				t.Fatalf("String()=%q want %q", got.String(), tc.want)
			}
			if got.IsZero() {
				t.Fatalf("IsZero() = true for valid value")
			}
		})
	}
}

func TestIPAddress_IsZero(t *testing.T) {
	t.Parallel()
	var zero domain.IPAddress
	if !zero.IsZero() {
		t.Fatalf("zero value should report IsZero")
	}
	if zero.String() != "" {
		t.Fatalf("zero String() = %q, want empty", zero.String())
	}
}
