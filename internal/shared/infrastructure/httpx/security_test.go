package httpx

import (
	"testing"
)

func TestGenerateToken_NotEmpty(t *testing.T) {
	token := GenerateToken()
	if token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestGenerateToken_Unique(t *testing.T) {
	tokens := make(map[string]bool)
	for i := 0; i < 100; i++ {
		token := GenerateToken()
		if tokens[token] {
			t.Fatalf("duplicate token generated on iteration %d: %s", i, token)
		}
		tokens[token] = true
	}
}

func TestGenerateToken_UUIDFormat(t *testing.T) {
	token := GenerateToken()
	// UUID format: 8-4-4-4-12 = 36 chars
	if len(token) != 36 {
		t.Errorf("expected token length 36 (UUID format), got %d", len(token))
	}
	// Check dashes at correct positions
	if token[8] != '-' || token[13] != '-' || token[18] != '-' || token[23] != '-' {
		t.Errorf("token does not match UUID format: %s", token)
	}
}
