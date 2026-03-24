package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testRSAKeyPEM generates a PEM-encoded RSA private key for testing.
func testRSAKeyPEM(t *testing.T) (privatePEM, publicPEM string) {
	t.Helper()
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	privBytes := x509.MarshalPKCS1PrivateKey(privKey)
	privBlock := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes}
	privatePEM = string(pem.EncodeToMemory(privBlock))

	pubBytes, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	require.NoError(t, err)
	pubBlock := &pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes}
	publicPEM = string(pem.EncodeToMemory(pubBlock))

	return privatePEM, publicPEM
}

func TestParseRSAPrivateKey(t *testing.T) {
	validPEM, _ := testRSAKeyPEM(t)

	tests := []struct {
		name    string
		keyStr  string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid_pem_key",
			keyStr:  validPEM,
			wantErr: false,
		},
		{
			name:    "empty_string",
			keyStr:  "",
			wantErr: true,
			errMsg:  "key string is empty",
		},
		{
			name:    "invalid_pem_data",
			keyStr:  "not-a-valid-pem-key",
			wantErr: true,
		},
		{
			name:    "key_with_literal_newlines",
			keyStr:  replaceNewlines(validPEM),
			wantErr: false,
		},
		{
			name:    "key_with_surrounding_quotes",
			keyStr:  `"` + validPEM + `"`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := ParseRSAPrivateKey(tt.keyStr)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, key)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, key)
			}
		})
	}
}

func TestParseRSAPublicKey(t *testing.T) {
	_, validPEM := testRSAKeyPEM(t)

	tests := []struct {
		name    string
		keyStr  string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid_pem_key",
			keyStr:  validPEM,
			wantErr: false,
		},
		{
			name:    "empty_string",
			keyStr:  "",
			wantErr: true,
			errMsg:  "key string is empty",
		},
		{
			name:    "invalid_pem_data",
			keyStr:  "not-a-valid-pem-key",
			wantErr: true,
		},
		{
			name:    "key_with_literal_newlines",
			keyStr:  replaceNewlines(validPEM),
			wantErr: false,
		},
		{
			name:    "key_with_surrounding_quotes",
			keyStr:  `"` + validPEM + `"`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := ParseRSAPublicKey(tt.keyStr)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, key)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, key)
			}
		})
	}
}

// replaceNewlines replaces real newlines with literal \n to simulate env-file format.
func replaceNewlines(s string) string {
	result := ""
	for _, c := range s {
		if c == '\n' {
			result += `\n`
		} else {
			result += string(c)
		}
	}
	return result
}
