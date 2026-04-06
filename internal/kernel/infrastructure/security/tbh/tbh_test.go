package tbh_test

import (
	"testing"

	"gct/internal/kernel/infrastructure/security/tbh"
)

var testPepper = []byte("test-pepper-secret-32-bytes!!")

func TestCompute_Deterministic(t *testing.T) {
	a := tbh.Compute(testPepper, "192.168.1.1", "Mozilla/5.0")
	b := tbh.Compute(testPepper, "192.168.1.1", "Mozilla/5.0")
	if a != b {
		t.Fatalf("expected deterministic output, got %q and %q", a, b)
	}
}

func TestCompute_DifferentIP(t *testing.T) {
	a := tbh.Compute(testPepper, "192.168.1.1", "Mozilla/5.0")
	b := tbh.Compute(testPepper, "10.0.0.1", "Mozilla/5.0")
	if a == b {
		t.Fatal("different IPs must produce different hashes")
	}
}

func TestCompute_DifferentUA(t *testing.T) {
	a := tbh.Compute(testPepper, "192.168.1.1", "Mozilla/5.0")
	b := tbh.Compute(testPepper, "192.168.1.1", "curl/7.88")
	if a == b {
		t.Fatal("different user-agents must produce different hashes")
	}
}

func TestVerify_Match(t *testing.T) {
	hash := tbh.Compute(testPepper, "192.168.1.1", "Mozilla/5.0")
	if !tbh.Verify(testPepper, "192.168.1.1", "Mozilla/5.0", hash) {
		t.Fatal("Verify should return true for matching inputs")
	}
}

func TestVerify_Mismatch(t *testing.T) {
	hash := tbh.Compute(testPepper, "192.168.1.1", "Mozilla/5.0")
	if tbh.Verify(testPepper, "10.0.0.1", "Mozilla/5.0", hash) {
		t.Fatal("Verify should return false for mismatched inputs")
	}
}

func TestCompute_EmptyIPAndUA(t *testing.T) {
	hash := tbh.Compute(testPepper, "", "")
	if hash == "" {
		t.Fatal("empty IP/UA should still produce a valid hash")
	}
	// Should also be deterministic.
	if hash2 := tbh.Compute(testPepper, "", ""); hash != hash2 {
		t.Fatal("empty IP/UA should be deterministic")
	}
}

func TestCompute_OutputLength(t *testing.T) {
	hash := tbh.Compute(testPepper, "192.168.1.1", "Mozilla/5.0")
	// 16 bytes base64url-encoded without padding: ceil(16*4/3) = 22 chars.
	if len(hash) != 22 {
		t.Fatalf("expected 22-char base64url output, got %d chars: %q", len(hash), hash)
	}
}
