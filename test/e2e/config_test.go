package e2e

import (
	"testing"
	"time"

	"gct/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewTestConfig tests the test configuration loading
func TestNewTestConfig(t *testing.T) {
	// Reset any existing test config
	config.ResetTestConfig()

	// Load test configuration
	cfg, err := config.NewTestConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify test environment settings
	assert.Equal(t, "go-clean-template-test", cfg.App.Name)
	assert.Equal(t, "1.0.0", cfg.App.Version)
	assert.Equal(t, "test", cfg.App.Environment)
	assert.True(t, cfg.IsTest())
	assert.False(t, cfg.IsProd())
	assert.False(t, cfg.IsDev())

	// Verify HTTP settings
	assert.Equal(t, "8080", cfg.HTTP.Port)
	assert.Equal(t, ":8080", cfg.HTTP.Addr())
	assert.False(t, cfg.HTTP.UsePreforkMode)

	// Verify database settings
	assert.Equal(t, "localhost", cfg.Database.Postgres.Host)
	assert.Equal(t, 5432, cfg.Database.Postgres.Port)
	assert.Equal(t, "test_db", cfg.Database.Postgres.Name)
	assert.Equal(t, "test_user", cfg.Database.Postgres.User)
	assert.Equal(t, "test_password", cfg.Database.Postgres.Password)
	assert.Equal(t, "disable", cfg.Database.Postgres.SSLMode)

	// Verify Redis settings
	assert.Equal(t, "localhost", cfg.Database.Redis.Host)
	assert.Equal(t, 6379, cfg.Database.Redis.Port)
	assert.Equal(t, "1", cfg.Database.Redis.Name)

	// Verify SQLite settings
	assert.Equal(t, "./test_data.db", cfg.Database.SqlLite.File)

	// Verify JWT settings
	assert.Equal(t, "auth-service-test", cfg.JWT.Issuer)
	assert.Equal(t, 15*time.Minute, cfg.JWT.AccessTTL)
	assert.Equal(t, 720*time.Hour, cfg.JWT.RefreshTTL)

	// Verify API keys
	assert.Equal(t, "test_api_key_12345", cfg.APIKeys.XApiKey)

	// Verify metrics and docs
	assert.True(t, cfg.Metrics.Enabled)
	assert.True(t, cfg.Swagger.Enabled)
	assert.True(t, cfg.Proto.Enabled)

	// Verify log level
	assert.Equal(t, "debug", cfg.Log.Level)
	assert.True(t, cfg.Log.IsDebug())
}

// TestConfigSingleton tests that config behaves as expected
func TestConfigSingleton(t *testing.T) {
	// Reset any existing test config
	config.ResetTestConfig()

	// Load configuration twice
	cfg1, err := config.NewTestConfig()
	require.NoError(t, err)

	cfg2, err := config.NewTestConfig()
	require.NoError(t, err)

	// Both should be the same instance (singleton)
	assert.Same(t, cfg1, cfg2)
}

// TestDatabaseConnectionStrings tests database URL generation
func TestDatabaseConnectionStrings(t *testing.T) {
	// Reset any existing test config
	config.ResetTestConfig()

	// Load test config
	cfg, err := config.NewTestConfig()
	require.NoError(t, err)

	// Test PostgreSQL URL
	pgURL := cfg.Database.Postgres.URL()
	expectedPGURL := "postgres://test_user:test_password@localhost:5432/test_db?sslmode=disable"
	assert.Equal(t, expectedPGURL, pgURL)

	// Test MySQL URL
	mysqlURL := cfg.Database.MySQL.URL()
	expectedMySQLURL := "test_user:test_password@tcp(localhost:3306)/test_db?parseTime=true"
	assert.Equal(t, expectedMySQLURL, mysqlURL)

	// Test SQLite DSN
	sqliteDSN := cfg.Database.SqlLite.DSN()
	expectedSQLiteDSN := "./test_data.db"
	assert.Equal(t, expectedSQLiteDSN, sqliteDSN)

	// Test PostgreSQL validation
	err = cfg.Database.Postgres.Validate()
	assert.NoError(t, err)
	assert.False(t, cfg.Database.Postgres.IsSecure())
}
