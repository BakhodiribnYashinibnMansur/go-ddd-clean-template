package domain_test

import (
	"strings"
	"testing"

	"gct/internal/kernel/domain"

	"pgregory.net/rapid"
)

var validMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}

func TestHTTPMethod_Property_CaseNormalization(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		canonical := rapid.SampledFrom(validMethods).Draw(t, "method")
		// randomize case of each character
		runes := []rune(canonical)
		for i, r := range runes {
			if rapid.Bool().Draw(t, "lower") {
				runes[i] = []rune(strings.ToLower(string(r)))[0]
			}
		}
		input := string(runes)

		m, err := domain.NewHTTPMethod(input)
		if err != nil {
			t.Fatalf("valid method %q rejected: %v", input, err)
		}
		if m.String() != canonical {
			t.Fatalf("String() = %q, want %q", m.String(), canonical)
		}
	})
}

func TestHTTPMethod_Property_Idempotency(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		canonical := rapid.SampledFrom(validMethods).Draw(t, "method")
		m1, err := domain.NewHTTPMethod(canonical)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		m2, err := domain.NewHTTPMethod(m1.String())
		if err != nil {
			t.Fatalf("idempotency failed: %v", err)
		}
		if m1 != m2 {
			t.Fatalf("not idempotent: %v != %v", m1, m2)
		}
	})
}

func TestHTTPMethod_Property_ClosedSet(t *testing.T) {
	t.Parallel()
	validSet := map[string]bool{
		"GET": true, "POST": true, "PUT": true, "PATCH": true,
		"DELETE": true, "HEAD": true, "OPTIONS": true,
	}
	rapid.Check(t, func(t *rapid.T) {
		s := rapid.String().Draw(t, "input")
		upper := strings.ToUpper(strings.TrimSpace(s))
		_, err := domain.NewHTTPMethod(s)
		if validSet[upper] {
			if err != nil {
				t.Fatalf("valid method %q rejected: %v", s, err)
			}
		} else {
			if err == nil {
				t.Fatalf("invalid method %q accepted", s)
			}
		}
	})
}

func TestHTTPMethod_Property_WhitespaceTrim(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		canonical := rapid.SampledFrom(validMethods).Draw(t, "method")
		spaces := rapid.StringMatching(`^[ \t]{0,5}$`).Draw(t, "spaces")
		input := spaces + canonical + spaces

		m, err := domain.NewHTTPMethod(input)
		if err != nil {
			t.Fatalf("method with whitespace %q rejected: %v", input, err)
		}
		if m.String() != canonical {
			t.Fatalf("String() = %q, want %q", m.String(), canonical)
		}
	})
}
