package security

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSecureStorage(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{
			name:    "valid_key",
			key:     "my-secret-encryption-key",
			wantErr: false,
		},
		{
			name:    "short_key",
			key:     "k",
			wantErr: false, // sha256 normalizes any length to 32 bytes
		},
		{
			name:    "empty_key",
			key:     "",
			wantErr: false, // sha256 handles empty string
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage, err := NewSecureStorage(tt.key)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, storage)
			}
		})
	}
}

func TestSecureStorage_EncryptDecrypt(t *testing.T) {
	storage, err := NewSecureStorage("test-encryption-key-12345")
	require.NoError(t, err)

	tests := []struct {
		name      string
		plaintext string
	}{
		{
			name:      "simple_text",
			plaintext: "hello world",
		},
		{
			name:      "empty_string",
			plaintext: "",
		},
		{
			name:      "json_data",
			plaintext: `{"user_id":"abc-123","role":"admin"}`,
		},
		{
			name:      "unicode_text",
			plaintext: "hello world, tes vett testovych textov",
		},
		{
			name:      "long_text",
			plaintext: "a]b[c{d}e(f)g!h@i#j$k%l^m&n*o-p=q+r/s\\t|u~v`w'x\"y;z:1,2.3<4>5?6 7\t8\n9",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := storage.Encrypt(tt.plaintext)
			require.NoError(t, err)
			assert.NotEmpty(t, encrypted)
			assert.NotEqual(t, tt.plaintext, encrypted)

			decrypted, err := storage.Decrypt(encrypted)
			require.NoError(t, err)
			assert.Equal(t, tt.plaintext, decrypted)
		})
	}
}

func TestSecureStorage_EncryptProducesDifferentCiphertext(t *testing.T) {
	storage, err := NewSecureStorage("test-key")
	require.NoError(t, err)

	plaintext := "same-input"
	c1, err := storage.Encrypt(plaintext)
	require.NoError(t, err)
	c2, err := storage.Encrypt(plaintext)
	require.NoError(t, err)

	// Different nonces should produce different ciphertexts
	assert.NotEqual(t, c1, c2)

	// But both should decrypt to the same value
	d1, err := storage.Decrypt(c1)
	require.NoError(t, err)
	d2, err := storage.Decrypt(c2)
	require.NoError(t, err)
	assert.Equal(t, d1, d2)
}

func TestSecureStorage_DecryptInvalidData(t *testing.T) {
	storage, err := NewSecureStorage("test-key")
	require.NoError(t, err)

	tests := []struct {
		name    string
		data    string
		wantErr bool
	}{
		{
			name:    "invalid_base64",
			data:    "not-valid-base64!!!",
			wantErr: true,
		},
		{
			name:    "too_short_data",
			data:    "YQ==", // base64 of "a", shorter than nonce
			wantErr: true,
		},
		{
			name:    "empty_string",
			data:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := storage.Decrypt(tt.data)
			assert.Error(t, err)
		})
	}
}

func TestSecureStorage_DifferentKeys(t *testing.T) {
	storage1, err := NewSecureStorage("key-one")
	require.NoError(t, err)
	storage2, err := NewSecureStorage("key-two")
	require.NoError(t, err)

	encrypted, err := storage1.Encrypt("secret data")
	require.NoError(t, err)

	// Decrypting with a different key should fail
	_, err = storage2.Decrypt(encrypted)
	assert.Error(t, err)
}

func TestDeviceID(t *testing.T) {
	id1, err := DeviceID()
	require.NoError(t, err)
	assert.NotEmpty(t, id1)

	id2, err := DeviceID()
	require.NoError(t, err)
	assert.NotEqual(t, id1, id2, "two device IDs should not be equal")
}
