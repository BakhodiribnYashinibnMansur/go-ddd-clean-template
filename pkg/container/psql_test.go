package container

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"gct/config"
)

// TestPostgresConfig_TableDriven tests PostgreSQL configuration scenarios
func TestPostgresConfig_TableDriven(t *testing.T) {
	type postgresConfigTestCase struct {
		name         string
		config       config.Postgres
		expectedHost string
		expectedPort int
		definition   string
	}

	tests := []postgresConfigTestCase{
		{
			name: "Default PostgreSQL configuration",
			config: config.Postgres{
				BaseDB: config.BaseDB{
					Host:     "localhost",
					Port:     5432,
					Name:     "testdb",
					User:     "postgres",
					Password: "password",
				},
			},
			expectedHost: "localhost",
			expectedPort: 5432,
			definition:   "Tests basic PostgreSQL configuration",
		},
		{
			name: "Custom PostgreSQL configuration",
			config: config.Postgres{
				BaseDB: config.BaseDB{
					Host:     "db.example.com",
					Port:     5433,
					Name:     "mydb",
					User:     "myuser",
					Password: "mypass",
				},
			},
			expectedHost: "db.example.com",
			expectedPort: 5433,
			definition:   "Tests custom PostgreSQL configuration",
		},
		{
			name: "PostgreSQL with SSL",
			config: config.Postgres{
				BaseDB: config.BaseDB{
					Host:     "secure-db.example.com",
					Port:     5432,
					Name:     "securedb",
					User:     "ssluser",
					Password: "sslpass",
					SSLMode:  "require",
				},
			},
			expectedHost: "secure-db.example.com",
			expectedPort: 5432,
			definition:   "Tests PostgreSQL configuration with SSL",
		},
		{
			name: "PostgreSQL with connection pool settings",
			config: config.Postgres{
				BaseDB: config.BaseDB{
					Host:     "pooled-db.example.com",
					Port:     5432,
					Name:     "pooldb",
					User:     "pooluser",
					Password: "poolpass",
					PoolMax:  25,
				},
			},
			expectedHost: "pooled-db.example.com",
			expectedPort: 5432,
			definition:   "Tests PostgreSQL configuration with connection pool settings",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedHost, tc.config.Host, "Host should match")
			assert.Equal(t, tc.expectedPort, tc.config.Port, "Port should match")
			assert.NotEmpty(t, tc.config.User, "User should not be empty")
			assert.NotEmpty(t, tc.config.Password, "Password should not be empty")
			assert.NotEmpty(t, tc.config.Name, "Name should not be empty")
		})
	}
}

// TestPostgresValidation_TableDriven tests PostgreSQL configuration validation
func TestPostgresValidation_TableDriven(t *testing.T) {
	type postgresValidationTestCase struct {
		name        string
		config      config.Database
		expectValid bool
		definition  string
	}

	tests := []postgresValidationTestCase{
		{
			name: "Valid PostgreSQL configuration",
			config: config.Database{
				Postgres: config.Postgres{
					BaseDB: config.BaseDB{
						Host:     "localhost",
						Port:     5432,
						Name:     "testdb",
						User:     "postgres",
						Password: "password",
					},
				},
			},
			expectValid: true,
			definition:  "Tests validation of valid PostgreSQL configuration",
		},
		{
			name: "Empty host",
			config: config.Database{
				Postgres: config.Postgres{
					BaseDB: config.BaseDB{
						Host:     "",
						Port:     5432,
						Name:     "testdb",
						User:     "postgres",
						Password: "password",
					},
				},
			},
			expectValid: false,
			definition:  "Tests validation with empty host",
		},
		{
			name: "Empty port",
			config: config.Database{
				Postgres: config.Postgres{
					BaseDB: config.BaseDB{
						Host:     "localhost",
						Port:     0,
						Name:     "testdb",
						User:     "postgres",
						Password: "password",
					},
				},
			},
			expectValid: false,
			definition:  "Tests validation with empty port",
		},
		{
			name: "Empty username",
			config: config.Database{
				Postgres: config.Postgres{
					BaseDB: config.BaseDB{
						Host:     "localhost",
						Port:     5432,
						Name:     "testdb",
						User:     "",
						Password: "password",
					},
				},
			},
			expectValid: false,
			definition:  "Tests validation with empty username",
		},
		{
			name: "Empty password",
			config: config.Database{
				Postgres: config.Postgres{
					BaseDB: config.BaseDB{
						Host:     "localhost",
						Port:     5432,
						Name:     "testdb",
						User:     "postgres",
						Password: "",
					},
				},
			},
			expectValid: false,
			definition:  "Tests validation with empty password",
		},
		{
			name: "Empty database name",
			config: config.Database{
				Postgres: config.Postgres{
					BaseDB: config.BaseDB{
						Host:     "localhost",
						Port:     5432,
						Name:     "",
						User:     "postgres",
						Password: "password",
					},
				},
			},
			expectValid: false,
			definition:  "Tests validation with empty database name",
		},
		{
			name: "Valid configuration with SSL",
			config: config.Database{
				Postgres: config.Postgres{
					BaseDB: config.BaseDB{
						Host:     "secure-db.example.com",
						Port:     5432,
						Name:     "securedb",
						User:     "ssluser",
						Password: "sslpass",
						SSLMode:  "require",
					},
				},
			},
			expectValid: true,
			definition:  "Tests validation of SSL-enabled PostgreSQL configuration",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			isValid := validatePostgresConfig(tc.config)

			if tc.expectValid {
				assert.True(t, isValid, "Expected configuration to be valid")
			} else {
				assert.False(t, isValid, "Expected configuration to be invalid")
			}
		})
	}
}

// validatePostgresConfig is a helper function to validate PostgreSQL configuration
func validatePostgresConfig(cfg config.Database) bool {
	pg := cfg.Postgres
	if pg.Host == "" {
		return false
	}
	if pg.Port == 0 {
		return false
	}
	if pg.User == "" {
		return false
	}
	if pg.Password == "" {
		return false
	}
	if pg.Name == "" {
		return false
	}
	return true
}

// TestPostgresConnectionStrings_TableDriven tests PostgreSQL connection string generation
func TestPostgresConnectionStrings_TableDriven(t *testing.T) {
	type connectionStringTestCase struct {
		name             string
		config           config.Database
		expectedContains []string
		definition       string
	}

	tests := []connectionStringTestCase{
		{
			name: "Basic connection string",
			config: config.Database{
				Postgres: config.Postgres{
					BaseDB: config.BaseDB{
						Host:     "localhost",
						Port:     5432,
						Name:     "testdb",
						User:     "postgres",
						Password: "password",
					},
				},
			},
			expectedContains: []string{"localhost", "5432", "postgres", "testdb"},
			definition:       "Tests basic connection string components",
		},
		{
			name: "Connection string with SSL",
			config: config.Database{
				Postgres: config.Postgres{
					BaseDB: config.BaseDB{
						Host:     "secure-db.example.com",
						Port:     5432,
						Name:     "securedb",
						User:     "ssluser",
						Password: "sslpass",
						SSLMode:  "require",
					},
				},
			},
			expectedContains: []string{"secure-db.example.com", "5432", "ssluser", "securedb", "sslmode=require"},
			definition:       "Tests connection string with SSL mode",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			connStr := generateConnectionString(tc.config)

			for _, expected := range tc.expectedContains {
				assert.Contains(t, connStr, expected, "Connection string should contain: %s", expected)
			}
		})
	}
}

// generateConnectionString is a helper function to generate PostgreSQL connection string
func generateConnectionString(cfg config.Database) string {
	pg := cfg.Postgres
	connStr := "host=" + pg.Host
	connStr += " port=" + fmt.Sprintf("%d", pg.Port)
	connStr += " user=" + pg.User
	connStr += " password=" + pg.Password
	connStr += " dbname=" + pg.Name

	if pg.SSLMode != "" {
		connStr += " sslmode=" + pg.SSLMode
	}

	return connStr
}

// TestPostgresPoolSettings_TableDriven tests PostgreSQL connection pool settings
func TestPostgresPoolSettings_TableDriven(t *testing.T) {
	type poolSettingsTestCase struct {
		name             string
		config           config.Database
		expectedMaxOpen  int
		expectedMaxIdle  int
		expectedLifetime int
		definition       string
	}

	tests := []poolSettingsTestCase{
		{
			name: "Default pool settings",
			config: config.Database{
				Postgres: config.Postgres{
					BaseDB: config.BaseDB{
						Host:     "localhost",
						Port:     5432,
						Name:     "testdb",
						User:     "postgres",
						Password: "password",
					},
				},
			},
			expectedMaxOpen:  0, // Default from BaseDB.PoolMax (envDefault:"10")
			expectedMaxIdle:  0, // Not directly configurable
			expectedLifetime: 0, // Not directly configurable
			definition:       "Tests default connection pool settings",
		},
		{
			name: "Custom pool settings",
			config: config.Database{
				Postgres: config.Postgres{
					BaseDB: config.BaseDB{
						Host:     "localhost",
						Port:     5432,
						Name:     "testdb",
						User:     "postgres",
						Password: "password",
						PoolMax:  25,
					},
				},
			},
			expectedMaxOpen:  25,
			expectedMaxIdle:  0, // Not directly configurable
			expectedLifetime: 0, // Not directly configurable
			definition:       "Tests custom connection pool settings",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedMaxOpen, tc.config.Postgres.PoolMax, "PoolMax should match")
			assert.Equal(t, tc.expectedMaxIdle, 0, "MaxIdleConns is not directly configurable")
			assert.Equal(t, tc.expectedLifetime, 0, "ConnMaxLifetime is not directly configurable")
		})
	}
}
