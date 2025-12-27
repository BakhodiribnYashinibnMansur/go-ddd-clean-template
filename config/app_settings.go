package config

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
