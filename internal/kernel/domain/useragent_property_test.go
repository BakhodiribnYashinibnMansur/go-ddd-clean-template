package domain_test

import (
	"strings"
	"testing"
	"unicode/utf8"

	"gct/internal/kernel/domain"

	"pgregory.net/rapid"
)

func TestUserAgent_Property_Infallible(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		s := rapid.String().Draw(t, "input")
		// must not panic
		_ = domain.NewUserAgent(s)
	})
}

func TestUserAgent_Property_LengthBound(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		s := rapid.String().Draw(t, "input")
		ua := domain.NewUserAgent(s)
		if n := utf8.RuneCountInString(ua.String()); n > 512 {
			t.Fatalf("String() has %d runes, want <= 512", n)
		}
	})
}

func TestUserAgent_Property_WhitespaceTrim(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		s := rapid.String().Draw(t, "input")
		ua := domain.NewUserAgent(s)
		result := ua.String()
		if result != strings.TrimSpace(result) {
			t.Fatalf("String() = %q has leading/trailing whitespace", result)
		}
	})
}

func TestUserAgent_Property_Idempotency(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		s := rapid.String().Draw(t, "input")
		ua1 := domain.NewUserAgent(s)
		ua2 := domain.NewUserAgent(ua1.String())
		if ua1.String() != ua2.String() {
			t.Fatalf("not idempotent: %q -> %q", ua1.String(), ua2.String())
		}
	})
}
