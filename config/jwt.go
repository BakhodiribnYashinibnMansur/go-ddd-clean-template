package config

import (
	"errors"
	"time"
)

// Standard error definitions for JWT configuration validation.
var (
	ErrMissingJWTAdminSecret  = errors.New("JWT admin secret key is required")
	ErrMissingJWTClientSecret = errors.New("JWT client secret key is required")
	ErrMissingJWTPrivateKey   = errors.New("JWT private key is required")
	ErrMissingJWTPublicKey    = errors.New("JWT public key is required")
	ErrInvalidAccessTTL       = errors.New("JWT access TTL must be greater than 0")
	ErrInvalidRefreshTTL      = errors.New("JWT refresh TTL must be greater than 0")
)

// JWT defines the parameters for asymmetric (RSA) token signing and lifecycle management.
type JWT struct {
	PrivateKey string        `yaml:"private_key" env:"PRIVATE_KEY"` // PEM-encoded RSA Private Key for signing tokens.
	PublicKey  string        `yaml:"public_key" env:"PUBLIC_KEY"`   // PEM-encoded RSA Public Key for verifying tokens.
	AccessTTL  time.Duration `yaml:"access_ttl" env:"ACCESS_TTL"`
	RefreshTTL time.Duration `yaml:"refresh_ttl" env:"REFRESH_TTL"`
	Issuer     string        `yaml:"issuer" env:"ISSUER"`
}

// Validate ensures that token durations are logical and properly configured.
func (j *JWT) Validate() error {
	if j.PrivateKey == "" {
		return ErrMissingJWTPrivateKey
	}
	if j.PublicKey == "" {
		return ErrMissingJWTPublicKey
	}
	if j.AccessTTL <= 0 {
		return ErrInvalidAccessTTL
	}
	if j.RefreshTTL <= 0 {
		return ErrInvalidRefreshTTL
	}
	return nil
}
