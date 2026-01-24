// Package config manages application-wide settings by parsing environment variables and .env files.
// It leverages a singleton pattern to ensure consistent configuration across all packages.
package config

import (
	"fmt"
	"os"
	"reflect"
	"sync"

	"github.com/caarlos0/env/v11"
	"github.com/subosito/gotenv"
	"gopkg.in/yaml.v3"
)

var (
	instance *Config   // Cached single instance of the configuration.
	once     sync.Once // Ensures thread-safe initialization of the singleton.
)

// Config represents the root configuration tree for the entire application.
// It is composed of specialized sub-structures, each mapped to specific logical components.
type Config struct {
	App          App          // Global application metadata (env, name, version).
	HTTP         HTTP         // Web server settings (port, timeouts).
	Log          Log          // Logging preferences (level, format).
	Database     Database     // Persistent storage connection details (Postgres).
	Connectivity Connectivity // Remote service health check parameters.
	JWT          JWT          // Authentication token parameters (secrets, TTL).
	Firebase     Firebase     `envPrefix:"FIREBASE_"` // Firebase Admin SDK integration.
	APIKeys      APIKeys      // Registered keys for service-to-service auth.
	Metrics      Metrics      // Observability and monitoring exports.
	Swagger      Swagger      // API documentation visibility and metadata.
	Proto        Proto        // Protocol Buffer and gRPC generated settings.
	Admin        Admin        // Reserved administrative account credentials.
	Cookie       Cookie       // HTTP cookie attributes (SameSite, Secure).
	CORS         CORS         `yaml:"cors"`                       // Cross-Origin Resource Sharing settings.
	Minio        MinioStore   `envPrefix:"MINIO_"`                // S3-compatible storage configuration.
	Redis        RedisStore   `envPrefix:"REDIS_"`                // Distributed caching and locking.
	Telegram     Telegram     `envPrefix:"TELEGRAM_"`             // Bot integration for notifications.
	Tracing      Tracing      `envPrefix:"TRACING_"`              // Distributed tracing export settings.
	Limiter      Limiter      `envPrefix:"LIMITER_"`              // Global and per-IP rate limit rules.
	Security     Security     `envPrefix:"SECURITY_"`             // Cross-cutting safety flags.
	FeatureFlag  FeatureFlag  `envPrefix:"FEATURE_FLAG_"`         // dynamic toggle controls.
	Asynq        AsynqConfig  `yaml:"asynq" envPrefix:"ASYNQ_"`   // background task queue settings.
	Seeder       Seeder       `yaml:"seeder" envPrefix:"SEEDER_"` // Mock data generation parameters.
	Middleware   Middleware   `yaml:"middleware"`                 // Middleware toggle flags.
}

// Telegram holds credentials for interacting with the Telegram Bot API.
type Telegram struct {
	BotToken string `env:"BOT_TOKEN"`
	ChatID   string `env:"CHAT_ID"`
}

// Security contains flags to toggle specialized safety measures.
type Security struct {
	FetchMetadata bool `env:"FETCH_METADATA_ENABLED" envDefault:"true"`
}

// NewConfig initializes or returns the existing application configuration.
// It attempts to load from a .env file before parsing system environment variables.
func NewConfig() (*Config, error) {
	var err error
	once.Do(func() {
		// Attempt to load .env into process environment. Failures are ignored if the file is missing.
		_ = gotenv.Load()

		cfg := &Config{}
		// Map environment variables to struct fields using "env" tags.
		if e := env.Parse(cfg); e != nil {
			err = fmt.Errorf("config parse error: %w", e)
			return
		}

		// Perform sanitation (e.g. removing stray quotes from env strings).
		cleanConfigStrings(reflect.ValueOf(cfg).Elem())

		// Load YAML configuration
		yamlFile, errYaml := os.ReadFile("config.yaml")
		if errYaml == nil {
			if errYaml := yaml.Unmarshal(yamlFile, cfg); errYaml != nil {
				err = fmt.Errorf("yaml parse error: %w", errYaml)
				return
			}
		}

		instance = cfg
	})

	if err != nil {
		return nil, err
	}

	return instance, nil
}

// IsProd returns true if the current environment is set to production.
func (c *Config) IsProd() bool {
	return c.App.IsProd()
}

// IsDev returns true if the current environment is set to development.
func (c *Config) IsDev() bool {
	return c.App.IsDev()
}

// IsTest returns true if the current environment is set to testing.
func (c *Config) IsTest() bool {
	return c.App.IsTest()
}
