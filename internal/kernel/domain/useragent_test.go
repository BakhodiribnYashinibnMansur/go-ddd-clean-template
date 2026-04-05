package domain_test

import (
	"strings"
	"testing"

	"gct/internal/kernel/domain"
)

func TestNewUserAgent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		input     string
		wantLen   int
		wantEmpty bool
		wantExact string
	}{
		{name: "normal", input: "Mozilla/5.0 (X11; Linux x86_64)", wantExact: "Mozilla/5.0 (X11; Linux x86_64)"},
		{name: "trimmed", input: "   curl/8.0  ", wantExact: "curl/8.0"},
		{name: "empty", input: "", wantEmpty: true},
		{name: "only whitespace", input: "    ", wantEmpty: true},
		{name: "truncated at 512", input: strings.Repeat("a", 1000), wantLen: 512},
		{name: "exactly 512 preserved", input: strings.Repeat("b", 512), wantLen: 512},
		{name: "511 preserved", input: strings.Repeat("c", 511), wantLen: 511},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := domain.NewUserAgent(tc.input)

			if tc.wantEmpty {
				if !got.IsZero() {
					t.Fatalf("expected IsZero, got %q", got.String())
				}
				return
			}
			if got.IsZero() {
				t.Fatalf("unexpected zero value for input %q", tc.input)
			}
			if tc.wantExact != "" && got.String() != tc.wantExact {
				t.Fatalf("String()=%q want %q", got.String(), tc.wantExact)
			}
			if tc.wantLen != 0 && len([]rune(got.String())) != tc.wantLen {
				t.Fatalf("rune len=%d want %d", len([]rune(got.String())), tc.wantLen)
			}
		})
	}
}

func TestUserAgent_IsZero(t *testing.T) {
	t.Parallel()
	var zero domain.UserAgent
	if !zero.IsZero() {
		t.Fatalf("zero value should report IsZero")
	}
	if zero.String() != "" {
		t.Fatalf("zero String() = %q, want empty", zero.String())
	}
}
