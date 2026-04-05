package domain_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"testing"

	"gct/internal/context/admin/supporting/integration/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewJWTAPIKey_Valid(t *testing.T) {
	t.Parallel()
	k, err := domain.NewJWTAPIKey("0123456789abcdef0123456789abcdef") // 32 chars
	require.NoError(t, err)
	require.False(t, k.IsZero())
	assert.Equal(t, "[REDACTED]", k.String())
	assert.Equal(t, "0123456789abcdef0123456789abcdef", k.Reveal())
}

func TestNewJWTAPIKey_RejectsShortInput(t *testing.T) {
	t.Parallel()
	k, err := domain.NewJWTAPIKey("too-short")
	require.Error(t, err)
	require.True(t, errors.Is(err, domain.ErrInvalidJWTAPIKey))
	assert.True(t, k.IsZero())
}

func TestJWTAPIKey_HashIsDeterministic(t *testing.T) {
	t.Parallel()
	pepper := []byte("the-quick-brown-fox-pepper-value")
	plain := "0123456789abcdef0123456789abcdef"
	k, err := domain.NewJWTAPIKey(plain)
	require.NoError(t, err)

	h1 := k.Hash(pepper)
	h2 := k.Hash(pepper)
	assert.Equal(t, h1, h2)

	// Matches direct HMAC-SHA256 computation.
	mac := hmac.New(sha256.New, pepper)
	mac.Write([]byte(plain))
	assert.Equal(t, mac.Sum(nil), h1)
}

func TestJWTAPIKey_HashDiffersWithDifferentPepper(t *testing.T) {
	t.Parallel()
	plain := "0123456789abcdef0123456789abcdef"
	k, _ := domain.NewJWTAPIKey(plain)
	h1 := k.Hash([]byte("pepperpepperpepperpepperpepper-A"))
	h2 := k.Hash([]byte("pepperpepperpepperpepperpepper-B"))
	assert.NotEqual(t, h1, h2)
}

func TestVerifyJWTAPIKeyHash_Match(t *testing.T) {
	t.Parallel()
	pepper := []byte("verification-pepper-0000000000000")
	plain := "0123456789abcdef0123456789abcdef"
	k, _ := domain.NewJWTAPIKey(plain)
	stored := k.Hash(pepper)

	assert.True(t, domain.VerifyJWTAPIKeyHash(plain, pepper, stored))
}

func TestVerifyJWTAPIKeyHash_Mismatch(t *testing.T) {
	t.Parallel()
	pepper := []byte("verification-pepper-0000000000000")
	k, _ := domain.NewJWTAPIKey("0123456789abcdef0123456789abcdef")
	stored := k.Hash(pepper)

	assert.False(t, domain.VerifyJWTAPIKeyHash("fedcba9876543210fedcba9876543210", pepper, stored))
}

func TestGenerateJWTAPIKey(t *testing.T) {
	t.Parallel()
	k, err := domain.GenerateJWTAPIKey()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(k.Reveal()), 32)

	// Two generations should differ.
	k2, err := domain.GenerateJWTAPIKey()
	require.NoError(t, err)
	assert.NotEqual(t, k.Reveal(), k2.Reveal())
}
