package config

// AsynqConfig - Asynq task queue configuration.
type AsynqConfig struct {
	// Redis configuration for Asynq (uses same Redis instance)
	RedisAddr     string `env:"ADDR"`
	RedisPassword string `env:"PASSWORD"`
	RedisDB       int    `env:"DB" envDefault:"0"`

	// Worker configuration
	Concurrency int `env:"CONCURRENCY" envDefault:"10"` // Number of concurrent workers

	// Queue priorities (higher number = higher priority)
	Queues map[string]int `env:"QUEUES"`

	// Retry configuration
	MaxRetry int `env:"MAX_RETRY" envDefault:"3"`

	// Enable/Disable worker
	WorkerEnabled bool `env:"WORKER_ENABLED" envDefault:"true"`
}

// GetDefaultQueues returns default queue priorities.
func (a *AsynqConfig) GetDefaultQueues() map[string]int {
	if len(a.Queues) > 0 {
		return a.Queues
	}

	// Default queue priorities
	return map[string]int{
		"critical": 6, // Highest priority
		"default":  3, // Medium priority
		"low":      1, // Lowest priority
	}
}
