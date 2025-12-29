package config

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrMissingDBHost     = errors.New("database host is required")
	ErrMissingDBPort     = errors.New("database port is required")
	ErrMissingDBName     = errors.New("database name is required")
	ErrMissingDBUser     = errors.New("database user is required")
	ErrMissingDBPassword = errors.New("database password is required")
)

// Database groups all supported databases -.
type Database struct {
	Postgres      Postgres      `envPrefix:"PG_"`
	MySQL         MySQL         `envPrefix:"MYSQL_"`
	MongoDB       MongoDB       `envPrefix:"MONGO_"`
	Redis         Redis         `envPrefix:"REDIS_"`
	Cassandra     Cassandra     `envPrefix:"CASSANDRA_"`
	Elasticsearch Elasticsearch `envPrefix:"ELASTIC_"`
	ClickHouse    ClickHouse    `envPrefix:"CH_"`
	SqlLite       SqlLite       `envPrefix:"SQLITE_"`
}

// BaseDB contains common database connection fields -.
type BaseDB struct {
	Host     string `env:"HOST,required"`
	Port     int    `env:"PORT,required"`
	Name     string `env:"NAME,required"`
	User     string `env:"USER,required"`
	Password string `env:"PASSWORD,required"`
	SSLMode  string `env:"SSL_MODE"          envDefault:"disable"`
	PoolMax  int    `env:"POOL_MAX"          envDefault:"10"`
}

// Database specific configs -.
type (
	Postgres      struct{ BaseDB }
	MySQL         struct{ BaseDB }
	MongoDB       struct{ BaseDB }
	Redis         struct{ BaseDB }
	Cassandra     struct{ BaseDB }
	Elasticsearch struct{ BaseDB }
	ClickHouse    struct{ BaseDB }

	SqlLite struct {
		File string `env:"FILE,required"`
	}
)

// URL returns connection string for Postgres.
func (p *Postgres) URL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		p.User, p.Password, p.Host, p.Port, p.Name, p.SSLMode)
}

// Validate validates database configuration.
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

// IsSecure returns true if SSL mode is enabled.
func (p *Postgres) IsSecure() bool {
	mode := strings.ToLower(p.SSLMode)
	return mode != "disable" && mode != ""
}

// URL returns connection string for MySQL.
func (m *MySQL) URL() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		m.User, m.Password, m.Host, m.Port, m.Name)
}

// DSN returns connection string for SqlLite.
func (s *SqlLite) DSN() string {
	return s.File
}
