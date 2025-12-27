package config

import "strings"

type (
	// App -.
	App struct {
		Name        string `env:"APP_NAME,required"`
		Version     string `env:"APP_VERSION,required"`
		Environment string `env:"APP_ENV" envDefault:"development"`
	}

	// HTTP -.
	HTTP struct {
		Port           string `env:"HTTP_PORT,required"`
		UsePreforkMode bool   `env:"HTTP_USE_PREFORK_MODE" envDefault:"false"`
	}

	// Log -.
	Log struct {
		Level string `env:"LOG_LEVEL,required"`
	}

	// APIKeys configuration -.
	APIKeys struct {
		XApiKey string `env:"X_API_KEY,required"`
	}

	// Metrics -.
	Metrics struct {
		Enabled bool `env:"METRICS_ENABLED" envDefault:"true"`
	}

	// Swagger -.
	Swagger struct {
		Enabled bool `env:"SWAGGER_ENABLED" envDefault:"false"`
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
