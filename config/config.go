package config

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/caarlos0/env/v11"
	"github.com/subosito/gotenv"
)

var (
	instance *Config
	once     sync.Once
)

// Config - main configuration structure.
type Config struct {
	App          App
	HTTP         HTTP
	Log          Log
	Database     Database
	Connectivity Connectivity
	JWT          JWT
	Firebase     Firebase `envPrefix:"FIREBASE_"`
	APIKeys      APIKeys
	Metrics      Metrics
	Swagger      Swagger
	Proto        Proto
	Admin        Admin
	Cookie       Cookie
	Minio        MinioStore  `envPrefix:"MINIO_"`
	Redis        RedisStore  `envPrefix:"REDIS_"`
	Telegram     Telegram    `envPrefix:"TELEGRAM_"`
	Tracing      Tracing     `envPrefix:"TRACING_"`
	Limiter      Limiter     `envPrefix:"LIMITER_"`
	Security     Security    `envPrefix:"SECURITY_"`
	FeatureFlag  FeatureFlag `envPrefix:"FEATURE_FLAG_"`
	Asynq        AsynqConfig `envPrefix:"ASYNQ_"`
	Seeder       Seeder      `envPrefix:"SEEDER_"`
}

type Telegram struct {
	BotToken string `env:"BOT_TOKEN"`
	ChatID   string `env:"CHAT_ID"`
}

// Security -.
type Security struct {
	FetchMetadata bool `env:"FETCH_METADATA_ENABLED" envDefault:"true"`
}

// NewConfig returns app config (Singleton).
func NewConfig() (*Config, error) {
	var err error
	once.Do(func() {
		// Load .env file if it exists
		_ = gotenv.Load()

		cfg := &Config{}
		if e := env.Parse(cfg); e != nil {
			err = fmt.Errorf("config error: %w", e)
			return
		}

		// Clean up string fields from quotes
		cleanConfigStrings(reflect.ValueOf(cfg).Elem())

		instance = cfg
	})

	if err != nil {
		return nil, err
	}

	return instance, nil
}

// IsProd delegates to App.IsProd
func (c *Config) IsProd() bool {
	return c.App.IsProd()
}

// IsDev delegates to App.IsDev
func (c *Config) IsDev() bool {
	return c.App.IsDev()
}

// IsTest delegates to App.IsTest
func (c *Config) IsTest() bool {
	return c.App.IsTest()
}
