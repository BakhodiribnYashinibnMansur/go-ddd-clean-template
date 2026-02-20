package config

// AsynqConfig - Asynq task queue configuration.
type AsynqConfig struct {
	Enabled bool `yaml:"enabled"` // Enable/disable Asynq.

	// Redis configuration for Asynq (uses same Redis instance)
	RedisAddr     string `yaml:"redis_addr"`
	RedisPassword string `yaml:"redis_password"`
	RedisDB       int    `yaml:"redis_db"`

	// Worker configuration
	Concurrency int `yaml:"concurrency"` // Number of concurrent workers

	// Queue priorities (higher number = higher priority)
	Queues map[string]int `yaml:"queues"`

	// Retry configuration
	MaxRetry int `yaml:"max_retry"`

	// Enable/Disable worker
	WorkerEnabled bool `yaml:"worker_enabled"`
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
