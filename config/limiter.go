package config

// Limiter - rate limiting configuration.
type Limiter struct {
	Enabled bool   `env:"LIMITER_ENABLED" envDefault:"true"`
	Limit   int64  `env:"LIMITER_LIMIT" envDefault:"100"` // Number of requests
	Period  string `env:"LIMITER_PERIOD" envDefault:"M"`  // Period (S: second, M: minute, H: hour)
}
