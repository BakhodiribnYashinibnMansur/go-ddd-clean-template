package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testRSAKeyPEM generates a PEM-encoded RSA key pair for testing.
func testRSAKeyPEM(t *testing.T) (privatePEM, publicPEM []byte) {
	t.Helper()
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	privBytes := x509.MarshalPKCS1PrivateKey(privKey)
	privatePEM = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes})

	pubBytes, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	require.NoError(t, err)
	publicPEM = pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})
	return
}

func TestParseRSAPrivateKey(t *testing.T) {
	validPEM, _ := testRSAKeyPEM(t)

	t.Run("valid", func(t *testing.T) {
		k, err := ParseRSAPrivateKey(validPEM)
		require.NoError(t, err)
		assert.NotNil(t, k)
	})

	t.Run("nil_returns_ErrEmptyKey", func(t *testing.T) {
		_, err := ParseRSAPrivateKey(nil)
		assert.ErrorIs(t, err, ErrEmptyKey)
	})

	t.Run("empty_returns_ErrEmptyKey", func(t *testing.T) {
		_, err := ParseRSAPrivateKey([]byte(""))
		assert.ErrorIs(t, err, ErrEmptyKey)
	})

	t.Run("garbage_returns_error", func(t *testing.T) {
		_, err := ParseRSAPrivateKey([]byte("not-a-pem-key"))
		assert.Error(t, err)
		assert.NotErrorIs(t, err, ErrEmptyKey)
	})

	// Security: the error must not echo any bytes of the input.
	t.Run("error_does_not_leak_key_material", func(t *testing.T) {
		// Use a distinctive, easy-to-spot marker.
		marker := "SENSITIVE-KEY-MATERIAL-DO-NOT-LEAK-0123456789abcdef"
		_, err := ParseRSAPrivateKey([]byte(marker))
		require.Error(t, err)
		assert.NotContains(t, err.Error(), marker)
		assert.NotContains(t, err.Error(), "SENSITIVE")
	})
}

func TestParseRSAPublicKey(t *testing.T) {
	_, validPEM := testRSAKeyPEM(t)

	t.Run("valid", func(t *testing.T) {
		k, err := ParseRSAPublicKey(validPEM)
		require.NoError(t, err)
		assert.NotNil(t, k)
	})

	t.Run("nil_returns_ErrEmptyKey", func(t *testing.T) {
		_, err := ParseRSAPublicKey(nil)
		assert.ErrorIs(t, err, ErrEmptyKey)
	})

	t.Run("garbage_returns_error", func(t *testing.T) {
		_, err := ParseRSAPublicKey([]byte("not-a-pem-key"))
		assert.Error(t, err)
	})

	t.Run("error_does_not_leak_key_material", func(t *testing.T) {
		marker := "PUBLIC-KEY-MARKER-DO-NOT-LEAK-ff00ff00ff00"
		_, err := ParseRSAPublicKey([]byte(marker))
		require.Error(t, err)
		assert.NotContains(t, err.Error(), marker)
		// Length is allowed; no substring of key bytes is allowed.
		assert.NotContains(t, strings.ToLower(err.Error()), "marker")
	})
}
