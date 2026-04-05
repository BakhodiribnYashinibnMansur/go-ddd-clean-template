package postgres

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func mustParseConfig(t *testing.T) *pgxpool.Config {
	t.Helper()
	cfg, err := pgxpool.ParseConfig("postgres://user:pass@localhost:5432/db")
	if err != nil {
		t.Fatalf("pgxpool.ParseConfig: %v", err)
	}
	return cfg
}

func TestWithMaxConns(t *testing.T) {
	tests := []struct {
		name     string
		maxConns int32
	}{
		{"set to 10", 10},
		{"set to 1", 1},
		{"set to 100", 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := mustParseConfig(t)
			WithMaxConns(tt.maxConns)(cfg)
			if cfg.MaxConns != tt.maxConns {
				t.Errorf("MaxConns = %d, want %d", cfg.MaxConns, tt.maxConns)
			}
		})
	}
}

func TestWithMinConns(t *testing.T) {
	tests := []struct {
		name     string
		minConns int32
	}{
		{"set to 0", 0},
		{"set to 5", 5},
		{"set to 20", 20},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := mustParseConfig(t)
			WithMinConns(tt.minConns)(cfg)
			if cfg.MinConns != tt.minConns {
				t.Errorf("MinConns = %d, want %d", cfg.MinConns, tt.minConns)
			}
		})
	}
}

func TestWithMaxConnLifetime(t *testing.T) {
	tests := []struct {
		name string
		d    time.Duration
	}{
		{"1 hour", time.Hour},
		{"30 minutes", 30 * time.Minute},
		{"0", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := mustParseConfig(t)
			WithMaxConnLifetime(tt.d)(cfg)
			if cfg.MaxConnLifetime != tt.d {
				t.Errorf("MaxConnLifetime = %v, want %v", cfg.MaxConnLifetime, tt.d)
			}
		})
	}
}

func TestWithMaxConnIdleTime(t *testing.T) {
	tests := []struct {
		name string
		d    time.Duration
	}{
		{"5 minutes", 5 * time.Minute},
		{"10 seconds", 10 * time.Second},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := mustParseConfig(t)
			WithMaxConnIdleTime(tt.d)(cfg)
			if cfg.MaxConnIdleTime != tt.d {
				t.Errorf("MaxConnIdleTime = %v, want %v", cfg.MaxConnIdleTime, tt.d)
			}
		})
	}
}

func TestWithHealthCheckPeriod(t *testing.T) {
	tests := []struct {
		name string
		d    time.Duration
	}{
		{"1 minute", time.Minute},
		{"30 seconds", 30 * time.Second},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := mustParseConfig(t)
			WithHealthCheckPeriod(tt.d)(cfg)
			if cfg.HealthCheckPeriod != tt.d {
				t.Errorf("HealthCheckPeriod = %v, want %v", cfg.HealthCheckPeriod, tt.d)
			}
		})
	}
}

func TestWithConnectTimeout(t *testing.T) {
	tests := []struct {
		name string
		d    time.Duration
	}{
		{"5 seconds", 5 * time.Second},
		{"1 minute", time.Minute},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := mustParseConfig(t)
			WithConnectTimeout(tt.d)(cfg)
			if cfg.ConnConfig.ConnectTimeout != tt.d {
				t.Errorf("ConnConfig.ConnectTimeout = %v, want %v", cfg.ConnConfig.ConnectTimeout, tt.d)
			}
		})
	}
}

func TestWithStatementTimeout(t *testing.T) {
	tests := []struct {
		name string
		d    time.Duration
		want string
	}{
		{"500ms", 500 * time.Millisecond, "500ms"},
		{"1 second", time.Second, "1000ms"},
		{"2 seconds", 2 * time.Second, "2000ms"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := mustParseConfig(t)
			WithStatementTimeout(tt.d)(cfg)
			got := cfg.ConnConfig.RuntimeParams["statement_timeout"]
			if got != tt.want {
				t.Errorf("RuntimeParams[\"statement_timeout\"] = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestWithApplicationName(t *testing.T) {
	tests := []struct {
		name    string
		appName string
	}{
		{"my-service", "my-service"},
		{"gct-app", "gct-app"},
		{"empty string", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := mustParseConfig(t)
			WithApplicationName(tt.appName)(cfg)
			got := cfg.ConnConfig.RuntimeParams["application_name"]
			if got != tt.appName {
				t.Errorf("RuntimeParams[\"application_name\"] = %q, want %q", got, tt.appName)
			}
		})
	}
}

func TestMultipleOptions(t *testing.T) {
	cfg := mustParseConfig(t)

	opts := []Option{
		WithMaxConns(20),
		WithMinConns(5),
		WithMaxConnLifetime(time.Hour),
		WithMaxConnIdleTime(10 * time.Minute),
		WithHealthCheckPeriod(30 * time.Second),
		WithConnectTimeout(5 * time.Second),
		WithStatementTimeout(500 * time.Millisecond),
		WithApplicationName("test-app"),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.MaxConns != 20 {
		t.Errorf("MaxConns = %d, want 20", cfg.MaxConns)
	}
	if cfg.MinConns != 5 {
		t.Errorf("MinConns = %d, want 5", cfg.MinConns)
	}
	if cfg.MaxConnLifetime != time.Hour {
		t.Errorf("MaxConnLifetime = %v, want %v", cfg.MaxConnLifetime, time.Hour)
	}
	if cfg.MaxConnIdleTime != 10*time.Minute {
		t.Errorf("MaxConnIdleTime = %v, want %v", cfg.MaxConnIdleTime, 10*time.Minute)
	}
	if cfg.HealthCheckPeriod != 30*time.Second {
		t.Errorf("HealthCheckPeriod = %v, want %v", cfg.HealthCheckPeriod, 30*time.Second)
	}
	if cfg.ConnConfig.ConnectTimeout != 5*time.Second {
		t.Errorf("ConnConfig.ConnectTimeout = %v, want %v", cfg.ConnConfig.ConnectTimeout, 5*time.Second)
	}
	if got := cfg.ConnConfig.RuntimeParams["statement_timeout"]; got != "500ms" {
		t.Errorf("RuntimeParams[\"statement_timeout\"] = %q, want \"500ms\"", got)
	}
	if got := cfg.ConnConfig.RuntimeParams["application_name"]; got != "test-app" {
		t.Errorf("RuntimeParams[\"application_name\"] = %q, want \"test-app\"", got)
	}
}

func TestWithStatementTimeout_InitializesRuntimeParams(t *testing.T) {
	cfg := mustParseConfig(t)
	// Ensure RuntimeParams is nil before applying the option.
	cfg.ConnConfig.RuntimeParams = nil

	WithStatementTimeout(500 * time.Millisecond)(cfg)

	if cfg.ConnConfig.RuntimeParams == nil {
		t.Fatal("RuntimeParams is nil after applying WithStatementTimeout; expected initialized map")
	}
	got := cfg.ConnConfig.RuntimeParams["statement_timeout"]
	if got != "500ms" {
		t.Errorf("RuntimeParams[\"statement_timeout\"] = %q, want \"500ms\"", got)
	}
}
