package config

// Sqlc configuration for SQL code generation
type Sqlc struct {
	Enabled bool `yaml:"enabled" env:"ENABLED" envDefault:"false"`
}
