// Package mysql implements MySQL connection.
package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"

	"gct/config"
	"gct/pkg/logger"
)

const (
	// ENV PROD
	maxOpenConnsProd = 50
	maxIdleConnsProd = 10

	// ENV DEV
	maxOpenConnsDev = 8
	maxIdleConnsDev = 3

	defaultConnMaxLifetime = 10 * time.Hour
	defaultConnMaxIdleTime = 30 * time.Minute
	defaultConnectTimeout  = 10 * time.Second
	defaultPingTimeout     = 5 * time.Second
)

// MySQL struct wraps sql.DB for MySQL connections.
type MySQL struct {
	DB *sql.DB
}

// New creates a new MySQL connection pool with optimized settings.
func New(ctx context.Context, env string, cfg config.MySQL, l logger.Log, opts ...Option) (*MySQL, error) {
	dsn := cfg.URL()

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		l.Errorw("failed to open MySQL connection", zap.Error(err))
		return nil, fmt.Errorf("open mysql connection: %w", err)
	}

	// Apply pool configuration
	applyPoolConfig(env, db)

	// Apply custom options
	for _, opt := range opts {
		opt(db)
	}

	// Verify connection with ping
	if err := verifyConnection(ctx, db, l); err != nil {
		db.Close()
		return nil, err
	}

	mysql := &MySQL{
		DB: db,
	}

	l.Infow("MySQL connected successfully")

	return mysql, nil
}

// applyPoolConfig applies pool configuration with defaults.
func applyPoolConfig(env string, db *sql.DB) {
	if env == "production" || env == "PROD" {
		db.SetMaxOpenConns(maxOpenConnsProd)
		db.SetMaxIdleConns(maxIdleConnsProd)
	} else {
		db.SetMaxOpenConns(maxOpenConnsDev)
		db.SetMaxIdleConns(maxIdleConnsDev)
	}

	db.SetConnMaxLifetime(defaultConnMaxLifetime)
	db.SetConnMaxIdleTime(defaultConnMaxIdleTime)
}

// verifyConnection pings the database to ensure connectivity.
func verifyConnection(ctx context.Context, db *sql.DB, l logger.Log) error {
	pingCtx, cancel := context.WithTimeout(ctx, defaultPingTimeout)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		l.Errorw("failed to ping MySQL database", zap.Error(err))
		return fmt.Errorf("verify mysql connection: %w", err)
	}

	return nil
}

// Close gracefully closes the MySQL connection pool.
func (m *MySQL) Close() error {
	if m != nil && m.DB != nil {
		return m.DB.Close()
	}
	return nil
}

// Stats returns current pool statistics.
func (m *MySQL) Stats() sql.DBStats {
	if m == nil || m.DB == nil {
		return sql.DBStats{}
	}
	return m.DB.Stats()
}
