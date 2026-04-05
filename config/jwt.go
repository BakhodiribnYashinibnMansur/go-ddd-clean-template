package config

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"time"
)

// Shared JWT settings. All per-integration data (audience, TTLs, API keys,
// RSA key pair) lives in the DB row for that integration — see the
// Integration BC. This struct holds only what is truly global to every
// JWT this server issues or accepts.
var (
	ErrMissingJWTIssuer     = errors.New("JWT issuer is required")
	ErrMissingRefreshPepper = errors.New("JWT refresh pepper is required")
	ErrMissingAPIKeyPepper  = errors.New("JWT API-key pepper is required")
	ErrWeakPepper           = errors.New("JWT pepper must decode to at least 32 bytes")
	ErrInvalidPepper        = errors.New("JWT pepper must be base64-encoded")
	ErrMissingKeysDir       = errors.New("JWT keys directory is required")
	ErrKeysDirNotFound      = errors.New("JWT keys directory does not exist")
)

const minPepperBytes = 32
const defaultLeeway = 30 * time.Second
const defaultCacheTTL = 30 * time.Second

// JWT holds infra-level JWT settings that are identical for every
// integration the server knows about.
type JWT struct {
	Issuer        string        `yaml:"issuer" env:"ISSUER"`                 // "iss" claim for every JWT.
	Leeway        time.Duration `yaml:"leeway" env:"LEEWAY"`                 // Clock-skew tolerance on exp/iat/nbf. Defaults to 30s.
	RefreshPepper string        `yaml:"refresh_pepper" env:"REFRESH_PEPPER"` // base64 HMAC key (>=32B) for refresh-token hashing.
	APIKeyPepper  string        `yaml:"api_key_pepper" env:"API_KEY_PEPPER"` // base64 HMAC key (>=32B) for hashing JWT API keys.
	KeysDir       string        `yaml:"keys_dir" env:"KEYS_DIR"`             // directory holding per-integration RSA PEM pairs.
	CacheTTL      time.Duration `yaml:"cache_ttl" env:"CACHE_TTL"`           // TTL for in-process integration cache. Defaults to 30s.

	// AutoGenerateKeys: on boot, generate a new RSA-4096 pair for any active
	// integration whose private-key file is missing under KeysDir.
	AutoGenerateKeys bool `yaml:"auto_generate_keys" env:"AUTO_GENERATE_KEYS" envDefault:"true"`
	// AutoRotate: schedule background rotation (see keyring package).
	AutoRotate bool `yaml:"auto_rotate" env:"AUTO_ROTATE" envDefault:"true"`
	// KeyBits: RSA key size when auto-generating. Allowed: 2048, 3072, 4096.
	KeyBits int `yaml:"key_bits" env:"KEY_BITS" envDefault:"4096"`
}

// Validate enforces non-empty issuer, peppers ≥32 bytes, and an accessible
// KeysDir. Leeway + CacheTTL get sensible defaults if unset.
func (j *JWT) Validate() error {
	if j.Issuer == "" {
		return ErrMissingJWTIssuer
	}
	if j.RefreshPepper == "" {
		return ErrMissingRefreshPepper
	}
	if _, err := j.DecodeRefreshPepper(); err != nil {
		return fmt.Errorf("refresh_pepper: %w", err)
	}
	if j.APIKeyPepper == "" {
		return ErrMissingAPIKeyPepper
	}
	if _, err := j.DecodeAPIKeyPepper(); err != nil {
		return fmt.Errorf("api_key_pepper: %w", err)
	}
	if j.KeysDir == "" {
		return ErrMissingKeysDir
	}
	if info, err := os.Stat(j.KeysDir); err != nil || !info.IsDir() {
		return fmt.Errorf("%w: %s", ErrKeysDirNotFound, j.KeysDir)
	}
	if j.Leeway == 0 {
		j.Leeway = defaultLeeway
	}
	if j.CacheTTL == 0 {
		j.CacheTTL = defaultCacheTTL
	}
	switch j.KeyBits {
	case 0:
		j.KeyBits = 4096
	case 2048, 3072, 4096:
		// ok
	default:
		return fmt.Errorf("invalid key_bits %d: allowed 2048, 3072, 4096", j.KeyBits)
	}
	return nil
}

// DecodeRefreshPepper decodes the base64 pepper into raw bytes and enforces
// the minimum length. Accepts std and URL base64, padded or unpadded.
// Generate one with: openssl rand -base64 48
func (j *JWT) DecodeRefreshPepper() ([]byte, error) {
	return decodePepper(j.RefreshPepper)
}

// DecodeAPIKeyPepper mirrors DecodeRefreshPepper for the API-key pepper.
func (j *JWT) DecodeAPIKeyPepper() ([]byte, error) {
	return decodePepper(j.APIKeyPepper)
}

func decodePepper(s string) ([]byte, error) {
	if s == "" {
		return nil, ErrMissingRefreshPepper
	}
	for _, enc := range []*base64.Encoding{
		base64.StdEncoding,
		base64.RawStdEncoding,
		base64.URLEncoding,
		base64.RawURLEncoding,
	} {
		if b, err := enc.DecodeString(s); err == nil {
			if len(b) < minPepperBytes {
				return nil, ErrWeakPepper
			}
			return b, nil
		}
	}
	return nil, ErrInvalidPepper
}
