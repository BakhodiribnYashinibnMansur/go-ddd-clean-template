package jwt

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testPepper is deterministic, 32 bytes, only used for hash correctness tests.
var testPepper = []byte("0123456789abcdef0123456789abcdef")

func newTestHasher(t *testing.T) *RefreshHasher {
	t.Helper()
	h, err := NewRefreshHasher(testPepper)
	require.NoError(t, err)
	return h
}

func TestNewRefreshHasher_PepperLength(t *testing.T) {
	_, err := NewRefreshHasher(make([]byte, 31))
	assert.ErrorIs(t, err, ErrPepperTooShort)

	_, err = NewRefreshHasher(nil)
	assert.ErrorIs(t, err, ErrPepperTooShort)

	_, err = NewRefreshHasher(make([]byte, 32))
	assert.NoError(t, err)
}

func TestNewRefreshHasher_CopiesPepper(t *testing.T) {
	pepper := make([]byte, 32)
	for i := range pepper {
		pepper[i] = byte(i)
	}
	h, err := NewRefreshHasher(pepper)
	require.NoError(t, err)

	// Baseline hash.
	hash1 := h.Hash("secret", "id")

	// Mutate caller's slice; hasher must be unaffected.
	for i := range pepper {
		pepper[i] = 0xff
	}
	hash2 := h.Hash("secret", "id")
	assert.Equal(t, hash1, hash2)
}

func TestRefreshHasher_HashVerify_RoundTrip(t *testing.T) {
	h := newTestHasher(t)
	stored := h.Hash("my-secret", "my-id")
	assert.True(t, h.Verify("my-secret", "my-id", stored))
}

func TestRefreshHasher_HashDeterministic(t *testing.T) {
	h := newTestHasher(t)
	assert.Equal(t, h.Hash("s", "id"), h.Hash("s", "id"))
	assert.NotEqual(t, h.Hash("s", "id"), h.Hash("s2", "id"))
	assert.NotEqual(t, h.Hash("s", "id"), h.Hash("s", "id2"))
}

func TestRefreshHasher_Verify_WrongSecret(t *testing.T) {
	h := newTestHasher(t)
	stored := h.Hash("real-secret", "id")
	assert.False(t, h.Verify("tampered-secret", "id", stored))
}

func TestRefreshHasher_Verify_WrongTokenID(t *testing.T) {
	h := newTestHasher(t)
	stored := h.Hash("secret", "real-id")
	assert.False(t, h.Verify("secret", "wrong-id", stored))
}

func TestRefreshHasher_Verify_DifferentPepper(t *testing.T) {
	h1 := newTestHasher(t)
	otherPepper := make([]byte, 32)
	for i := range otherPepper {
		otherPepper[i] = 0xaa
	}
	h2, err := NewRefreshHasher(otherPepper)
	require.NoError(t, err)

	stored := h1.Hash("secret", "id")
	assert.False(t, h2.Verify("secret", "id", stored))
}

func TestRefreshHasher_Verify_TamperedHashBits(t *testing.T) {
	h := newTestHasher(t)
	stored := []byte(h.Hash("secret", "id"))

	// Flip one bit in the stored hash — verify must fail.
	tampered := make([]byte, len(stored))
	copy(tampered, stored)
	tampered[0] ^= 0x01
	assert.False(t, h.Verify("secret", "id", string(tampered)))
}

func TestRefreshHasher_Verify_EmptyStoredHash(t *testing.T) {
	h := newTestHasher(t)
	assert.False(t, h.Verify("secret", "id", ""))
}

func TestGenerateRefreshToken_HappyPath(t *testing.T) {
	h := newTestHasher(t)
	tok, err := GenerateRefreshToken(h, "user-1", "session-1", "client-1", 7*24*time.Hour)
	require.NoError(t, err)

	assert.NotEmpty(t, tok.ID)
	assert.NotEmpty(t, tok.Secret)
	assert.NotEmpty(t, tok.Hashed)
	assert.Equal(t, "user-1", tok.UserID)
	assert.Equal(t, "session-1", tok.SessionID)
	assert.Equal(t, "client-1", tok.ClientID)
	assert.False(t, tok.IsExpired())
	assert.True(t, tok.ExpiresAt.After(time.Now()))

	// Hashed equals what the hasher would compute for the same inputs.
	assert.Equal(t, h.Hash(tok.Secret, tok.ID), tok.Hashed)
	assert.True(t, h.Verify(tok.Secret, tok.ID, tok.Hashed))
}

func TestGenerateRefreshToken_RejectsNilHasher(t *testing.T) {
	_, err := GenerateRefreshToken(nil, "u", "s", "c", time.Hour)
	assert.Error(t, err)
}

func TestGenerateRefreshToken_UniqueIDsAndSecrets(t *testing.T) {
	h := newTestHasher(t)
	tok1, err := GenerateRefreshToken(h, "u", "s", "c", time.Hour)
	require.NoError(t, err)
	tok2, err := GenerateRefreshToken(h, "u", "s", "c", time.Hour)
	require.NoError(t, err)

	assert.NotEqual(t, tok1.ID, tok2.ID)
	assert.NotEqual(t, tok1.Secret, tok2.Secret)
	assert.NotEqual(t, tok1.Hashed, tok2.Hashed)
}

func TestGenerateRefreshToken_NoBase64Padding(t *testing.T) {
	h := newTestHasher(t)
	tok, err := GenerateRefreshToken(h, "u", "s", "c", time.Hour)
	require.NoError(t, err)

	// RawURLEncoding output must never contain padding.
	assert.NotContains(t, tok.ID, "=")
	assert.NotContains(t, tok.Secret, "=")
	assert.NotContains(t, tok.Hashed, "=")
	// Or other URL-unsafe characters.
	assert.NotContains(t, tok.ID, "+")
	assert.NotContains(t, tok.ID, "/")
}

func TestRefreshToken_String(t *testing.T) {
	h := newTestHasher(t)
	tok, err := GenerateRefreshToken(h, "u", "sess-1", "c", time.Hour)
	require.NoError(t, err)

	s := tok.String()
	assert.True(t, strings.HasPrefix(s, "rft_v1."))
	parts := strings.SplitN(s[len("rft_"):], ".", 4)
	require.Len(t, parts, 4)
	assert.Equal(t, "v1", parts[0])
	assert.Equal(t, "sess-1", parts[1])
	assert.Equal(t, tok.ID, parts[2])
	assert.Equal(t, tok.Secret, parts[3])
}

func TestParseRefreshToken_RoundTrip(t *testing.T) {
	h := newTestHasher(t)
	orig, err := GenerateRefreshToken(h, "u", "sess-1", "c", time.Hour)
	require.NoError(t, err)

	parsed, err := ParseRefreshToken(orig.String())
	require.NoError(t, err)
	assert.Equal(t, orig.ID, parsed.ID)
	assert.Equal(t, orig.SessionID, parsed.SessionID)
	assert.Equal(t, orig.Secret, parsed.Secret)

	// Parsed token authenticates against original hash.
	assert.True(t, h.Verify(parsed.Secret, parsed.ID, orig.Hashed))
}

func TestParseRefreshToken_Errors(t *testing.T) {
	tests := []struct {
		name  string
		token string
	}{
		{"empty_string", ""},
		{"no_prefix", "v1.sess.id.secret"},
		{"wrong_prefix", "atk_v1.sess.id.secret"},
		{"too_few_parts", "rft_v1.only"},
		{"wrong_version", "rft_v999.sess.id.secret"},
		{"empty_session_id", "rft_v1..id.secret"},
		{"empty_id", "rft_v1.sess..secret"},
		{"empty_secret", "rft_v1.sess.id."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseRefreshToken(tt.token)
			assert.ErrorIs(t, err, ErrRefreshTokenInvalid)
		})
	}
}

func TestRefreshToken_IsExpired(t *testing.T) {
	h := newTestHasher(t)
	tok, err := GenerateRefreshToken(h, "u", "s", "c", -time.Hour)
	require.NoError(t, err)
	assert.True(t, tok.IsExpired())

	tok2, err := GenerateRefreshToken(h, "u", "s", "c", time.Hour)
	require.NoError(t, err)
	assert.False(t, tok2.IsExpired())
}
