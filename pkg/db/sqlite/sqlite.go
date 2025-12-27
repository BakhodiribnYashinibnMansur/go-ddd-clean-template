// Package sqlite implements SQLite connection.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/evrone/go-clean-template/config"
	"github.com/evrone/go-clean-template/pkg/logger"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

const (
	defaultMaxOpenConns = 1 // SQLite works best with single connection
	defaultMaxIdleConns = 1
	defaultPingTimeout  = 5 * time.Second
)

// SQLite struct wraps sql.DB for SQLite connections.
type SQLite struct {
	DB *sql.DB
}

// New creates a new SQLite connection.
func New(ctx context.Context, cfg config.SqlLite, l logger.Log, opts ...Option) (*SQLite, error) {
	dsn := cfg.DSN()

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		l.Errorw("failed to open SQLite connection", zap.Error(err))
		return nil, fmt.Errorf("open sqlite connection: %w", err)
	}

	// SQLite works best with single connection
	db.SetMaxOpenConns(defaultMaxOpenConns)
	db.SetMaxIdleConns(defaultMaxIdleConns)

	// Apply custom options
	for _, opt := range opts {
		opt(db)
	}

	// Verify connection with ping
	if err := verifyConnection(ctx, db, l); err != nil {
		db.Close()
		return nil, err
	}

	s := &SQLite{
		DB: db,
	}

	l.Infow("SQLite connected successfully", zap.String("file", dsn))

	return s, nil
}

// verifyConnection pings the database to ensure connectivity.
func verifyConnection(ctx context.Context, db *sql.DB, l logger.Log) error {
	pingCtx, cancel := context.WithTimeout(ctx, defaultPingTimeout)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		l.Errorw("failed to ping SQLite database", zap.Error(err))
		return fmt.Errorf("verify sqlite connection: %w", err)
	}

	return nil
}

// Close gracefully closes the SQLite connection.
func (s *SQLite) Close() error {
	if s != nil && s.DB != nil {
		return s.DB.Close()
	}
	return nil
}

// Stats returns current database statistics.
func (s *SQLite) Stats() sql.DBStats {
	if s == nil || s.DB == nil {
		return sql.DBStats{}
	}
	return s.DB.Stats()
}
