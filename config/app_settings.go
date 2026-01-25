package config

import "strings"

type (
	// App -.
	App struct {
		Name        string `yaml:"name"`
		Version     string `yaml:"version"`
		Environment string `env:"APP_ENV"              envDefault:"development"`
		CSRFSecret  string `env:"CSRF_SECRET,required"` // Dedicated secret for CSRF token generation
	}

	// HTTP -.
	HTTP struct {
		Port            string `env:"HTTP_PORT,required"`
		UsePreforkMode  bool   `env:"HTTP_USE_PREFORK_MODE" envDefault:"false"`
		GinMode         string `env:"GIN_MODE" envDefault:"debug"`
		ShutdownTimeout int64  `env:"HTTP_SHUTDOWN_TIMEOUT" envDefault:"5"` // Seconds
	}

	// Log -.
	Log struct {
		Level   string `yaml:"level"`
		ShowGin bool   `yaml:"show_gin" env:"LOG_SHOW_GIN" envDefault:"true"`
	}

	// APIKeys configuration -.
	APIKeys struct {
		XApiKey string `env:"X_API_KEY,required"`
	}

	// Metrics -.
	Metrics struct {
		Enabled bool `yaml:"enabled"`
	}

	// Swagger -.
	Swagger struct {
		Enabled bool `yaml:"enabled"`
	}

	// Proto -.
	Proto struct {
		Enabled bool `yaml:"enabled"`
	}

	// Admin -.
	Admin struct {
		Enabled bool `yaml:"enabled"`
	}

	// Cookie -.
	Cookie struct {
		Domain   string `yaml:"domain"`
		Path     string `yaml:"path"`
		HttpOnly bool   `yaml:"http_only"`
		MaxAge   int    `yaml:"max_age"`
		Secure   bool   `yaml:"secure"`
	}

	// CORS -.
	CORS struct {
		AllowedOrigins   []string `yaml:"allowed_origins"`
		AllowedMethods   []string `yaml:"allowed_methods"`
		AllowedHeaders   []string `yaml:"allowed_headers"`
		ExposedHeaders   []string `yaml:"exposed_headers"`
		AllowCredentials bool     `yaml:"allow_credentials"`
		MaxAge           int      `yaml:"max_age"`
	}

	// Middleware -.
	Middleware struct {
		RequestID    bool `yaml:"request_id"`
		Security     bool `yaml:"security"`
		MetaData     bool `yaml:"metadata"`
		Logger       bool `yaml:"logger"`
		Recovery     bool `yaml:"recovery"`
		Persist5xx   bool `yaml:"persist_5xx"`
		Mock         bool `yaml:"mock"`
		RateLimiter  bool `yaml:"rate_limiter"`
		AuditHistory bool `yaml:"audit_history"`
		AuditChange  bool `yaml:"audit_change"`
		Metrics      bool `yaml:"metrics"`
		HealthCheck  bool `yaml:"health_check"`
	}
)

// IsProd returns true if the environment is production.
func (a *App) IsProd() bool {
	env := strings.ToLower(a.Environment)
	return env == "prod" || env == "production"
}

// IsDev returns true if the environment is development.
func (a *App) IsDev() bool {
	env := strings.ToLower(a.Environment)
	return env == "dev" || env == "development" || env == ""
}

// IsTest returns true if the environment is test.
func (a *App) IsTest() bool {
	env := strings.ToLower(a.Environment)
	return env == "test" || env == "testing"
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
	return strings.ToLower(l.Level) == "debug"
}

// IsEnabled returns true if metrics are enabled.
func (m *Metrics) IsEnabled() bool {
	return m.Enabled
}

// IsEnabled returns true if swagger is enabled.
func (s *Swagger) IsEnabled() bool {
	return s.Enabled
}

// IsHttpOnly returns true if the cookie should be HttpOnly.
func (c *Cookie) IsHttpOnly() bool {
	return c.HttpOnly
}

// IsSecure returns true if the cookie should be Secure.
func (c *Cookie) IsSecure() bool {
	return c.Secure
}
