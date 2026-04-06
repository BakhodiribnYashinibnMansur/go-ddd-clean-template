package fingerprint

import (
	"testing"
)

func TestCompute_Deterministic(t *testing.T) {
	t.Parallel()
	fp1 := Compute("Mozilla/5.0", "en-US", "Chromium")
	fp2 := Compute("Mozilla/5.0", "en-US", "Chromium")
	if fp1 != fp2 {
		t.Fatalf("expected deterministic output, got %q and %q", fp1, fp2)
	}
}

func TestCompute_DifferentUA(t *testing.T) {
	t.Parallel()
	fp1 := Compute("Mozilla/5.0", "en-US", "Chromium")
	fp2 := Compute("Safari/17.0", "en-US", "Chromium")
	if fp1 == fp2 {
		t.Fatal("different user agents should produce different fingerprints")
	}
}

func TestCompute_DifferentLanguage(t *testing.T) {
	t.Parallel()
	fp1 := Compute("Mozilla/5.0", "en-US", "Chromium")
	fp2 := Compute("Mozilla/5.0", "de-DE", "Chromium")
	if fp1 == fp2 {
		t.Fatal("different accept-language should produce different fingerprints")
	}
}

func TestCompute_EmptyInputs(t *testing.T) {
	t.Parallel()
	fp := Compute("", "", "")
	if fp == "" {
		t.Fatal("fingerprint should not be empty even with empty inputs")
	}
	if len(fp) != 64 {
		t.Fatalf("expected 64-char hex string, got %d chars: %q", len(fp), fp)
	}
}

func TestCompute_OutputIs64CharHex(t *testing.T) {
	t.Parallel()
	fp := Compute("Mozilla/5.0", "en-US", "Chromium")
	if len(fp) != 64 {
		t.Fatalf("expected 64-char hex string, got %d chars: %q", len(fp), fp)
	}
	for _, c := range fp {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Fatalf("unexpected character %q in fingerprint", c)
		}
	}
}

func TestCompute_TrimsWhitespace(t *testing.T) {
	t.Parallel()
	fp1 := Compute("Mozilla/5.0", "en-US", "Chromium")
	fp2 := Compute("  Mozilla/5.0  ", "  en-US  ", "  Chromium  ")
	if fp1 != fp2 {
		t.Fatal("whitespace-padded inputs should produce same fingerprint as trimmed inputs")
	}
}
