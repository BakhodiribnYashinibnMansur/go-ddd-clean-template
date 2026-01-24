package config

import (
	"errors"
	"time"
)

// Standard error definitions for JWT configuration validation.
var (
	ErrMissingJWTAdminSecret  = errors.New("JWT admin secret key is required")
	ErrMissingJWTClientSecret = errors.New("JWT client secret key is required")
	ErrInvalidAccessTTL       = errors.New("JWT access TTL must be greater than 0")
	ErrInvalidRefreshTTL      = errors.New("JWT refresh TTL must be greater than 0")
)

// JWT defines the parameters for asymmetric (RSA) token signing and lifecycle management.
type JWT struct {
	PrivateKey string        `env:"JWT_PRIVATE_KEY"`                       // PEM-encoded RSA Private Key for signing tokens.
	PublicKey  string        `env:"JWT_PUBLIC_KEY"`                        // PEM-encoded RSA Public Key for verifying tokens.
	AccessTTL  time.Duration `env:"JWT_ACCESS_TTL"  envDefault:"15m"`      // Expiration time for short-lived access tokens.
	RefreshTTL time.Duration `env:"JWT_REFRESH_TTL" envDefault:"720h"`     // Expiration time for long-lived refresh tokens (default 30 days).
	Issuer     string        `env:"JWT_ISSUER"      envDefault:"gct-auth"` // Domain or service name claiming issuance.
}

// Validate ensures that token durations are logical and properly configured.
func (j *JWT) Validate() error {
	if j.AccessTTL <= 0 {
		return ErrInvalidAccessTTL
	}
	if j.RefreshTTL <= 0 {
		return ErrInvalidRefreshTTL
	}
	return nil
}
