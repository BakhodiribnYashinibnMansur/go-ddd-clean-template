package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"
)

func BenchmarkGenerateAccessToken(b *testing.B) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GenerateAccessToken("user-123", "session-456", "test-issuer", "test-audience", privateKey, 15*time.Minute)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseAccessToken(b *testing.B) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		b.Fatal(err)
	}

	tokenString, err := GenerateAccessToken("user-123", "session-456", "test-issuer", "test-audience", privateKey, 15*time.Minute)
	if err != nil {
		b.Fatal(err)
	}

	publicKey := &privateKey.PublicKey

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ParseAccessToken(tokenString, publicKey, "test-issuer", "test-audience")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGenerateRefreshToken(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GenerateRefreshToken("user-123", "session-456", "client-789", 7*24*time.Hour)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkVerifyRefreshToken(b *testing.B) {
	token, err := GenerateRefreshToken("user-123", "session-456", "client-789", 7*24*time.Hour)
	if err != nil {
		b.Fatal(err)
	}

	tokenString := token.String()
	hashedSecret := token.Hashed

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := VerifyRefreshToken(tokenString, hashedSecret)
		if err != nil {
			b.Fatal(err)
		}
	}
}
