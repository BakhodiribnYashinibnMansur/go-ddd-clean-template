package config

import "strings"

type (
	// App -.
	App struct {
		Name        string `yaml:"name"               env:"APP_NAME" validate:"required"`
		Version     string `yaml:"version"            env:"APP_VERSION" validate:"required"`
		Environment string `env:"APP_ENV"              envDefault:"development" validate:"oneof=development production test dev prod testing"`
		CSRFSecret  string `env:"CSRF_SECRET,required" validate:"required,min=32"` // Dedicated secret for CSRF token generation
	}

	// HTTP -.
	HTTP struct {
		Port            string `yaml:"port" env:"HTTP_PORT" validate:"required,numeric,min=1,max=65535"`
		UsePreforkMode  bool   `yaml:"use_prefork_mode" env:"HTTP_USE_PREFORK_MODE" envDefault:"false"`
		GinMode         string `env:"GIN_MODE" envDefault:"debug" validate:"oneof=debug release test"`
		ShutdownTimeout int64  `yaml:"shutdown_timeout" env:"HTTP_SHUTDOWN_TIMEOUT" envDefault:"5" validate:"min=1,max=60"` // Seconds
	}

	// Log -.
	Log struct {
		Level             string `yaml:"level"`
		Format            string `yaml:"format" env:"LOG_FORMAT" envDefault:"console" validate:"oneof=console json"`
		ShowGin           bool   `yaml:"show_gin" env:"LOG_SHOW_GIN" envDefault:"true"`
		SlowOpThresholdMs int64  `yaml:"slow_op_threshold_ms" env:"LOG_SLOW_OP_THRESHOLD_MS" envDefault:"500"`

		// Persistence — buffer logs in Redis, flush to PostgreSQL periodically
		PersistEnabled bool   `yaml:"persist_enabled" env:"LOG_PERSIST_ENABLED" envDefault:"false"`
		PersistLevel   string `yaml:"persist_level" env:"LOG_PERSIST_LEVEL" envDefault:"warn" validate:"oneof=debug info warn error"`
		RedisKey       string `yaml:"redis_key" env:"LOG_REDIS_KEY" envDefault:"app:logs"`
		FlushInterval  int64  `yaml:"flush_interval_sec" env:"LOG_FLUSH_INTERVAL_SEC" envDefault:"60"`
		FlushBatchSize int    `yaml:"flush_batch_size" env:"LOG_FLUSH_BATCH_SIZE" envDefault:"1000"`
		RetentionDays  int    `yaml:"retention_days" env:"LOG_RETENTION_DAYS" envDefault:"30"`

		// ReqLog — incoming HTTP request/response logging (http_request_logs table)
		ReqLogEnabled           bool     `yaml:"req_log_enabled" env:"REQLOG_ENABLED" envDefault:"true"`
		ReqLogMaxBodyBytes      int      `yaml:"req_log_max_body_bytes" env:"REQLOG_MAX_BODY_BYTES" envDefault:"8192"`
		ReqLogSuccessSampleRate float64  `yaml:"req_log_success_sample_rate" env:"REQLOG_SUCCESS_SAMPLE_RATE" envDefault:"0.05" validate:"min=0,max=1"`
		ReqLogSlowThresholdMs   int      `yaml:"req_log_slow_threshold_ms" env:"REQLOG_SLOW_THRESHOLD_MS" envDefault:"500"`
		ReqLogSkipPaths         []string `yaml:"req_log_skip_paths" env:"REQLOG_SKIP_PATHS" envSeparator:"," envDefault:"/health,/health/ready,/health/live,/metrics,/favicon.ico"`
		ReqLogSkipPrefixes      []string `yaml:"req_log_skip_prefixes" env:"REQLOG_SKIP_PREFIXES" envSeparator:"," envDefault:"/swagger/,/static/,/assets/"`
		ReqLogBodySuppressPaths []string `yaml:"req_log_body_suppress_paths" env:"REQLOG_BODY_SUPPRESS_PATHS" envSeparator:"," envDefault:"/api/v1/auth/login,/api/v1/auth/register,/api/v1/auth/reset-password,/api/v1/otp/verify"`

		// APILog — outgoing 3rd-party HTTP call logging (external_api_logs table)
		APILogSlowThresholdMs   int     `yaml:"api_log_slow_threshold_ms" env:"APILOG_SLOW_THRESHOLD_MS" envDefault:"2000"`
		APILogSuccessSampleRate float64 `yaml:"api_log_success_sample_rate" env:"APILOG_SUCCESS_SAMPLE_RATE" envDefault:"0" validate:"min=0,max=1"`
	}

	// APIKeys configuration -.
	APIKeys struct {
		SignExpireTime int64 `yaml:"sign_expire_time" env:"SIGN_EXPIRE_TIME" envDefault:"10"`
	}

	// Metrics -.
	Metrics struct {
		Enabled               bool   `yaml:"enabled"`
		SlowQueryThreshold    string `yaml:"slow_query_threshold" env:"METRICS_SLOW_QUERY_THRESHOLD" envDefault:"100ms"`
		LatencyEnabled        bool   `yaml:"latency_enabled" env:"METRICS_LATENCY_ENABLED" envDefault:"true"`
		LatencyP95Threshold   string `yaml:"latency_p95_threshold" env:"METRICS_LATENCY_P95_THRESHOLD" envDefault:"200ms"`
		LatencyP99Threshold   string `yaml:"latency_p99_threshold" env:"METRICS_LATENCY_P99_THRESHOLD" envDefault:"500ms"`
		LatencyWindowSec      int    `yaml:"latency_window_sec" env:"METRICS_LATENCY_WINDOW_SEC" envDefault:"60"`
		LatencyLogIntervalSec int    `yaml:"latency_log_interval_sec" env:"METRICS_LATENCY_LOG_INTERVAL_SEC" envDefault:"10"`
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
		Enabled bool   `yaml:"enabled"`
		URL     string `yaml:"url" env:"ADMIN_URL" envDefault:"http://localhost:3000"`
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
		Signature    bool `yaml:"signature"`
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
