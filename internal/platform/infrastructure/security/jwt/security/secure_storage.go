package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

var (
	ErrInvalidKey  = errors.New("invalid encryption key")
	ErrInvalidData = errors.New("invalid data")
)

// SecureStorage provides encryption for sensitive data
type SecureStorage struct {
	gcm cipher.AEAD
}

// NewSecureStorage creates a new secure storage instance
func NewSecureStorage(key string) (*SecureStorage, error) {
	// Convert key to 32 bytes for AES-256
	keyHash := sha256.Sum256([]byte(key))

	block, err := aes.NewCipher(keyHash[:])
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	return &SecureStorage{gcm: gcm}, nil
}

// Encrypt encrypts data using AES-GCM
func (s *SecureStorage) Encrypt(data string) (string, error) {
	nonce := make([]byte, s.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := s.gcm.Seal(nonce, nonce, []byte(data), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts data using AES-GCM
func (s *SecureStorage) Decrypt(encryptedData string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	nonceSize := s.gcm.NonceSize()
	if len(data) < nonceSize {
		return "", ErrInvalidData
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := s.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// DeviceID generates a unique device ID
func DeviceID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
