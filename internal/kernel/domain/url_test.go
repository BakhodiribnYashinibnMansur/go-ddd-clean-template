package domain_test

import (
	"errors"
	"strings"
	"testing"

	"gct/internal/kernel/domain"
)

func TestNewURL(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "valid http", input: "http://example.com", want: "http://example.com"},
		{name: "valid https", input: "https://example.com/path?x=1", want: "https://example.com/path?x=1"},
		{name: "valid https with port", input: "https://example.com:8443/a", want: "https://example.com:8443/a"},
		{name: "trimmed whitespace", input: "  https://example.com  ", want: "https://example.com"},
		{name: "missing scheme", input: "example.com/path", wantErr: true},
		{name: "ftp scheme rejected", input: "ftp://example.com", wantErr: true},
		{name: "missing host", input: "https://", wantErr: true},
		{name: "empty", input: "", wantErr: true},
		{name: "only whitespace", input: "   ", wantErr: true},
		{name: "too long", input: "https://example.com/" + strings.Repeat("a", 2100), wantErr: true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := domain.NewURL(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if !errors.Is(err, domain.ErrInvalidURL) {
					t.Fatalf("expected ErrInvalidURL, got %v", err)
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

func TestURL_IsZero(t *testing.T) {
	t.Parallel()
	var zero domain.URL
	if !zero.IsZero() {
		t.Fatalf("zero value should report IsZero")
	}
	if zero.String() != "" {
		t.Fatalf("zero String() = %q, want empty", zero.String())
	}
}
