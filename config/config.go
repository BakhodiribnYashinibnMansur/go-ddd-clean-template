package config

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/caarlos0/env/v11"
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
	Cookie       Cookie
	Minio        MinioStore `envPrefix:"MINIO_"`
}

// NewConfig returns app config (Singleton).
func NewConfig() (*Config, error) {
	var err error
	once.Do(func() {
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
