package config

import (
	"fmt"
	"strings"
)

// IsProd returns true if the environment is production.
func (c *Config) IsProd() bool {
	env := strings.ToLower(c.App.Environment)
	return env == "prod" || env == "production" || env == "PROD" || env == "Production"
}

// IsDev returns true if the environment is development.
func (c *Config) IsDev() bool {
	env := strings.ToLower(c.App.Environment)
	return env == "dev" || env == "development" || env == "" || env == "Dev" || env == "Development"
}

// IsTest returns true if the environment is test.
func (c *Config) IsTest() bool {
	env := strings.ToLower(c.App.Environment)
	return env == "test" || env == "testing" || env == "Test" || env == "Testing"
}

// Addr returns the HTTP server address.
func (h *HTTP) Addr() string {
	if h.Port == "" {
		return ":8080"
	}
	if strings.HasPrefix(h.Port, ":") {
		return h.Port
	}
	return ":" + h.Port
}

// IsDebug returns true if the log level is debug.
func (l *Log) IsDebug() bool {
	return strings.ToLower(l.Level) == "debug" || strings.ToLower(l.Level) == "DEBUG"
}

// URL returns connection string for Postgres.
func (p *Postgres) URL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		p.User, p.Password, p.Host, p.Port, p.Name, p.SSLMode)
}

// URL returns connection string for MySQL.
func (m *MySQL) URL() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		m.User, m.Password, m.Host, m.Port, m.Name)
}

// URL returns connection string for SqlLite.
func (s *SqlLite) DSN() string {
	return s.File
}
