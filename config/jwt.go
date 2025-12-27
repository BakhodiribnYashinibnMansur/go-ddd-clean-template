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
	AdminSecretKey  string        `env:"JWT_ADMIN_SECRET_KEY,required"`
	ClientSecretKey string        `env:"JWT_CLIENT_SECRET_KEY,required"`
	AccessTTL       time.Duration `env:"JWT_ACCESS_TTL" envDefault:"1h"`
	RefreshTTL      time.Duration `env:"JWT_REFRESH_TTL" envDefault:"24h"`
	Issuer          string        `env:"JWT_ISSUER" envDefault:"auth-service"`
}

// Validate validates JWT configuration.
func (j *JWT) Validate() error {
	if j.AdminSecretKey == "" {
		return ErrMissingJWTAdminSecret
	}
	if j.ClientSecretKey == "" {
		return ErrMissingJWTClientSecret
	}
	if j.AccessTTL <= 0 {
		return ErrInvalidAccessTTL
	}
	if j.RefreshTTL <= 0 {
		return ErrInvalidRefreshTTL
	}
	return nil
}
