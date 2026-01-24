package config

// Limiter - rate limiting configuration.
type Limiter struct {
	Enabled bool   `yaml:"enabled"`
	Limit   int64  `yaml:"limit"`  // Number of requests
	Period  string `yaml:"period"` // Period (S: second, M: minute, H: hour)
}
