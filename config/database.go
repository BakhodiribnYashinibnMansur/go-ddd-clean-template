package config

import (
	"errors"
	"fmt"
	"strings"
)

// Standard error definitions for database configuration validation.
var (
	ErrMissingDBHost     = errors.New("database host is required")
	ErrMissingDBPort     = errors.New("database port is required")
	ErrMissingDBName     = errors.New("database name is required")
	ErrMissingDBUser     = errors.New("database user is required")
	ErrMissingDBPassword = errors.New("database password is required")
)

// Database aggregates configurations for all supported storage engines.
// Each field uses an environment prefix to avoid naming collisions when loading from OS variables.
type Database struct {
	Postgres      Postgres      `envPrefix:"PG_"`
	MySQL         MySQL         `envPrefix:"MYSQL_"`
	MongoDB       MongoDB       `envPrefix:"MONGO_"`
	Redis         Redis         `envPrefix:"REDIS_"`
	Cassandra     Cassandra     `envPrefix:"CASSANDRA_"`
	Elasticsearch Elasticsearch `envPrefix:"ELASTIC_"`
	ClickHouse    ClickHouse    `envPrefix:"CH_"`
}

// BaseDB identifies common connectivity fields used across most relational and NoSQL databases.
type BaseDB struct {
	Enabled  bool   `yaml:"enabled" env:"ENABLED" envDefault:"false"` // Toggle to enable/disable this database connection.
	Host     string `env:"HOST,required"`                             // IP address or hostname of the server.
	Port     int    `env:"PORT,required"`                             // Communication port for the protocol.
	Name     string `env:"NAME,required"`                             // Target database/schema name.
	User     string `env:"USER,required"`                             // Authentication username.
	Password string `env:"PASSWORD,required"`                         // Authentication password.
	SSLMode  string `env:"SSL_MODE" envDefault:"disable"`             // encryption settings.
	PoolMax  int    `env:"POOL_MAX" envDefault:"10"`                  // Maximum concurrent connections in the pool.
}

// Specialized driver configurations inheriting common connection logic.
type (
	Postgres      struct{ BaseDB }
	MySQL         struct{ BaseDB }
	MongoDB       struct{ BaseDB }
	Redis         struct{ BaseDB }
	Cassandra     struct{ BaseDB }
	Elasticsearch struct{ BaseDB }
	ClickHouse    struct{ BaseDB }
)

// URL formats the connection string for the Postgres driver.
func (p *Postgres) URL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		p.User, p.Password, p.Host, p.Port, p.Name, p.SSLMode)
}

// Validate checks for the presence of mandatory Postgres connectivity fields.
func (p *Postgres) Validate() error {
	if p.Host == "" {
		return ErrMissingDBHost
	}
	if p.Port == 0 {
		return ErrMissingDBPort
	}
	if p.Name == "" {
		return ErrMissingDBName
	}
	if p.User == "" {
		return ErrMissingDBUser
	}
	if p.Password == "" {
		return ErrMissingDBPassword
	}
	return nil
}

// IsSecure returns true if the connection requires an encrypted SSL handshake.
func (p *Postgres) IsSecure() bool {
	mode := strings.ToLower(p.SSLMode)
	return mode != "disable" && mode != ""
}

// URL generates the DSN required for MySQL protocol interaction.
func (m *MySQL) URL() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		m.User, m.Password, m.Host, m.Port, m.Name)
}
