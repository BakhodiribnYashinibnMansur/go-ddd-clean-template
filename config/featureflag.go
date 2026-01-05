package config

// FeatureFlag configuration for feature flag management.
type FeatureFlag struct {
	Enabled          bool   `env:"FEATURE_FLAG_ENABLED" envDefault:"true"`
	ConfigPath       string `env:"FEATURE_FLAG_CONFIG_PATH" envDefault:"./config/flags.yaml"`
	PollingInterval  int    `env:"FEATURE_FLAG_POLLING_INTERVAL" envDefault:"60"` // seconds
	UseFileRetriever bool   `env:"FEATURE_FLAG_USE_FILE" envDefault:"true"`
	UseRedis         bool   `env:"FEATURE_FLAG_USE_REDIS" envDefault:"false"`
	RedisKey         string `env:"FEATURE_FLAG_REDIS_KEY" envDefault:"feature_flags"`
}
