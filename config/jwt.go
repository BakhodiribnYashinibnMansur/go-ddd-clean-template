package config

import "time"

// JWT configuration -.
type JWT struct {
	AdminSecretKey  string        `env:"JWT_ADMIN_SECRET_KEY,required"`
	ClientSecretKey string        `env:"JWT_CLIENT_SECRET_KEY,required"`
	AccessTTL       time.Duration `env:"JWT_ACCESS_TTL" envDefault:"1h"`
	RefreshTTL      time.Duration `env:"JWT_REFRESH_TTL" envDefault:"24h"`
	Issuer          string        `env:"JWT_ISSUER" envDefault:"auth-service"`
}
