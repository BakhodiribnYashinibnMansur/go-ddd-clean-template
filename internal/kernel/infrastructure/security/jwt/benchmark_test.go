package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"
)

func benchKeys(b *testing.B) *rsa.PrivateKey {
	b.Helper()
	pk, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		b.Fatal(err)
	}
	return pk
}

func BenchmarkGenerateAccessToken(b *testing.B) {
	pk := benchKeys(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := GenerateAccessToken("user-123", "sess-456", "iss", "aud", "kid-1", pk, 15*time.Minute); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseAccessToken(b *testing.B) {
	pk := benchKeys(b)
	tok, err := GenerateAccessToken("user-123", "sess-456", "iss", "aud", "kid-1", pk, 15*time.Minute)
	if err != nil {
		b.Fatal(err)
	}
	pub := &pk.PublicKey
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := ParseAccessToken(tok, pub, "iss", "aud", 0); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGenerateRefreshToken(b *testing.B) {
	h, err := NewRefreshHasher(testPepper)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := GenerateRefreshToken(h, "user-123", "sess-456", "client-789", 7*24*time.Hour); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRefreshHasher_Verify(b *testing.B) {
	h, err := NewRefreshHasher(testPepper)
	if err != nil {
		b.Fatal(err)
	}
	tok, err := GenerateRefreshToken(h, "u", "s", "c", 7*24*time.Hour)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if !h.Verify(tok.Secret, tok.ID, tok.Hashed) {
			b.Fatal("verify failed")
		}
	}
}
