package domain_test

import (
	"errors"
	"strings"
	"testing"

	"gct/internal/kernel/domain"

	"pgregory.net/rapid"
)

func genValidURL(t *rapid.T) string {
	scheme := rapid.SampledFrom([]string{"http", "https"}).Draw(t, "scheme")
	host := rapid.StringMatching(`[a-z][a-z0-9]{2,10}`).Draw(t, "host")
	tld := rapid.SampledFrom([]string{".com", ".org", ".net", ".io"}).Draw(t, "tld")
	path := rapid.SampledFrom([]string{"", "/", "/api", "/api/v1", "/path/to/resource"}).Draw(t, "path")
	return scheme + "://" + host + tld + path
}

func TestURL_Property_ValidURLAccepted(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		s := genValidURL(t)
		u, err := domain.NewURL(s)
		if err != nil {
			t.Fatalf("valid URL %q rejected: %v", s, err)
		}
		if u.IsZero() {
			t.Fatalf("valid URL %q returned zero value", s)
		}
	})
}

func TestURL_Property_Idempotency(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		s := genValidURL(t)
		u1, err := domain.NewURL(s)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		u2, err := domain.NewURL(u1.String())
		if err != nil {
			t.Fatalf("idempotency failed: %v", err)
		}
		if u1.String() != u2.String() {
			t.Fatalf("not idempotent: %q -> %q", u1.String(), u2.String())
		}
	})
}

func TestURL_Property_SchemeRejection(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		scheme := rapid.SampledFrom([]string{"ftp", "ssh", "ws", "wss", "file", "mailto"}).Draw(t, "scheme")
		host := rapid.StringMatching(`[a-z]{3,8}`).Draw(t, "host")
		s := scheme + "://" + host + ".com"

		_, err := domain.NewURL(s)
		if err == nil {
			t.Fatalf("non-http scheme %q accepted", s)
		}
		if !errors.Is(err, domain.ErrInvalidURL) {
			t.Fatalf("wrong error type: %v", err)
		}
	})
}

func TestURL_Property_LengthGuard(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		extra := rapid.IntRange(1, 500).Draw(t, "extra")
		padding := strings.Repeat("a", 2048+extra)
		s := "https://example.com/" + padding

		_, err := domain.NewURL(s)
		if err == nil {
			t.Fatalf("URL with len=%d accepted", len(s))
		}
		if !errors.Is(err, domain.ErrInvalidURL) {
			t.Fatalf("wrong error type: %v", err)
		}
	})
}

func TestURL_Property_NonZero(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		s := genValidURL(t)
		u, err := domain.NewURL(s)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if u.IsZero() {
			t.Fatalf("successfully constructed URL %q is zero", s)
		}
	})
}
