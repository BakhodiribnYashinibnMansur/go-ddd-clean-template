package redact

import (
	"net/http"
	"strings"
	"testing"
)

func TestHeaders_MasksSensitive(t *testing.T) {
	h := http.Header{}
	h.Set("Authorization", "Bearer secret-token")
	h.Set("Cookie", "sid=abc")
	h.Set("X-Api-Key", "k-123")
	h.Set("User-Agent", "Mozilla/5.0")

	out := Headers(h)
	if strings.Contains(out, "secret-token") {
		t.Fatalf("authorization leaked: %s", out)
	}
	if strings.Contains(out, "sid=abc") {
		t.Fatalf("cookie leaked: %s", out)
	}
	if strings.Contains(out, "k-123") {
		t.Fatalf("api key leaked: %s", out)
	}
	if !strings.Contains(out, "Mozilla/5.0") {
		t.Fatalf("user-agent was redacted but shouldn't be: %s", out)
	}
	if !strings.Contains(out, RedactedValue) {
		t.Fatalf("expected %q placeholder in output: %s", RedactedValue, out)
	}
}

func TestHeaders_StableSortedOutput(t *testing.T) {
	h := http.Header{}
	h.Set("Z-Last", "z")
	h.Set("A-First", "a")
	h.Set("M-Middle", "m")
	out := Headers(h)
	idxA := strings.Index(out, "A-First")
	idxM := strings.Index(out, "M-Middle")
	idxZ := strings.Index(out, "Z-Last")
	if !(idxA < idxM && idxM < idxZ) {
		t.Fatalf("headers not alphabetically sorted: %s", out)
	}
}

func TestHeaders_Empty(t *testing.T) {
	if out := Headers(nil); out != "" {
		t.Fatalf("expected empty for nil, got %q", out)
	}
	if out := Headers(http.Header{}); out != "" {
		t.Fatalf("expected empty for empty, got %q", out)
	}
}

func TestJSONBody_MasksSensitiveFields(t *testing.T) {
	in := []byte(`{"email":"a@b.com","password":"hunter2","profile":{"refresh_token":"xyz","name":"Ali"}}`)
	out := JSONBody(in, "application/json")

	if strings.Contains(out, "hunter2") {
		t.Fatalf("password leaked: %s", out)
	}
	if strings.Contains(out, "xyz") {
		t.Fatalf("refresh_token leaked: %s", out)
	}
	if !strings.Contains(out, "a@b.com") {
		t.Fatalf("non-sensitive field stripped: %s", out)
	}
	if !strings.Contains(out, `"name":"Ali"`) {
		t.Fatalf("name lost during redaction: %s", out)
	}
}

func TestJSONBody_HandlesArrays(t *testing.T) {
	in := []byte(`{"users":[{"password":"a"},{"password":"b"}]}`)
	out := JSONBody(in, "application/json")
	if strings.Contains(out, `"a"`) || strings.Contains(out, `"b"`) {
		t.Fatalf("array item passwords leaked: %s", out)
	}
}

func TestJSONBody_PassthroughNonJSON(t *testing.T) {
	in := []byte("plain text password=hunter2")
	out := JSONBody(in, "text/plain")
	if out != string(in) {
		t.Fatalf("non-json body was modified: %s", out)
	}
}

func TestJSONBody_MalformedReturnsOriginal(t *testing.T) {
	in := []byte(`{broken`)
	out := JSONBody(in, "application/json")
	if out != string(in) {
		t.Fatalf("malformed body was modified: %s", out)
	}
}

func TestJSONBody_EmptyReturnsEmpty(t *testing.T) {
	if out := JSONBody(nil, "application/json"); out != "" {
		t.Fatalf("expected empty, got %q", out)
	}
}

func TestTruncate(t *testing.T) {
	cases := []struct {
		in, want string
		n        int
	}{
		{"hello", "hello", 10},
		{"hello", "hello", 5},
		{"hello", "hel…", 3},
		{"hello", "hello", 0}, // n<=0 means no truncation
		{"", "", 10},
	}
	for _, c := range cases {
		if got := Truncate(c.in, c.n); got != c.want {
			t.Errorf("Truncate(%q,%d)=%q want %q", c.in, c.n, got, c.want)
		}
	}
}
