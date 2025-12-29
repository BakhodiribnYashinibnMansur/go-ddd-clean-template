package config

import (
	"errors"
	"time"
)

var (
	ErrMissingJWTAdminSecret  = errors.New("JWT admin secret key is required")
	ErrMissingJWTClientSecret = errors.New("JWT client secret key is required")
	ErrInvalidAccessTTL       = errors.New("JWT access TTL must be greater than 0")
	ErrInvalidRefreshTTL      = errors.New("JWT refresh TTL must be greater than 0")
)

// JWT configuration -.
type JWT struct {
	PrivateKey string        `env:"JWT_PRIVATE_KEY"` // RSA Private Key (PEM)
	PublicKey  string        `env:"JWT_PUBLIC_KEY"`  // RSA Public Key (PEM)
	AccessTTL  time.Duration `env:"JWT_ACCESS_TTL"  envDefault:"15m"`
	RefreshTTL time.Duration `env:"JWT_REFRESH_TTL" envDefault:"720h"` // 30 days
	Issuer     string        `env:"JWT_ISSUER"      envDefault:"auth-service"`
}

// Validate validates JWT configuration.
func (j *JWT) Validate() error {
	if j.AccessTTL <= 0 {
		return ErrInvalidAccessTTL
	}
	if j.RefreshTTL <= 0 {
		return ErrInvalidRefreshTTL
	}
	return nil
}
