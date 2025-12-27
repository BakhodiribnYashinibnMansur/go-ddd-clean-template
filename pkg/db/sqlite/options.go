package sqlite

import (
	"database/sql"
	"time"
)

// Option defines a function type for configuring SQLite connection.
type Option func(*sql.DB)

// WithMaxOpenConns sets the maximum number of open connections.
// Note: SQLite works best with 1 connection for write operations.
func WithMaxOpenConns(n int) Option {
	return func(db *sql.DB) {
		db.SetMaxOpenConns(n)
	}
}

// WithMaxIdleConns sets the maximum number of idle connections.
func WithMaxIdleConns(n int) Option {
	return func(db *sql.DB) {
		db.SetMaxIdleConns(n)
	}
}

// WithConnMaxLifetime sets the maximum lifetime of a connection.
func WithConnMaxLifetime(d time.Duration) Option {
	return func(db *sql.DB) {
		db.SetConnMaxLifetime(d)
	}
}

// WithConnMaxIdleTime sets the maximum idle time of a connection.
func WithConnMaxIdleTime(d time.Duration) Option {
	return func(db *sql.DB) {
		db.SetConnMaxIdleTime(d)
	}
}
