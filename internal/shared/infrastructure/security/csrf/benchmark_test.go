package csrf

import (
	"testing"
	"time"
)

func BenchmarkGenerateToken(b *testing.B) {
	gen := NewGenerator(Config{
		Secret:     []byte("bench-secret-key-32-bytes-long!!"),
		Expiration: DefaultExpiration,
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gen.GenerateToken("session-123")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkValidateToken(b *testing.B) {
	gen := NewGenerator(Config{
		Secret:     []byte("bench-secret-key-32-bytes-long!!"),
		Expiration: DefaultExpiration,
	})

	token, err := gen.GenerateToken("session-123")
	if err != nil {
		b.Fatal(err)
	}

	expiresAt := time.Now().Add(DefaultExpiration)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := gen.ValidateToken(token.Value, token.Hash, "session-123", expiresAt)
		if err != nil {
			b.Fatal(err)
		}
	}
}
