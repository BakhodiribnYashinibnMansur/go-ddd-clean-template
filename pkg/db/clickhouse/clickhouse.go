// Package clickhouse implements ClickHouse connection.
package clickhouse

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/evrone/go-clean-template/config"
	"github.com/evrone/go-clean-template/pkg/logger"
	"go.uber.org/zap"
)

const (
	// ENV PROD
	maxOpenConnsProd = 50
	maxIdleConnsProd = 10

	// ENV DEV
	maxOpenConnsDev = 8
	maxIdleConnsDev = 3

	defaultDialTimeout  = 10 * time.Second
	defaultReadTimeout  = 30 * time.Second
	defaultWriteTimeout = 30 * time.Second
	defaultPingTimeout  = 5 * time.Second
)

// ClickHouse struct wraps clickhouse connection.
type ClickHouse struct {
	Conn clickhouse.Conn
	DB   *sql.DB
}

// New creates a new ClickHouse connection with optimized settings.
func New(ctx context.Context, env string, cfg config.ClickHouse, l logger.Log, opts ...Option) (*ClickHouse, error) {
	options := &clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)},
		Auth: clickhouse.Auth{
			Database: cfg.Name,
			Username: cfg.User,
			Password: cfg.Password,
		},
		DialTimeout: defaultDialTimeout,
	}

	// Apply custom options
	for _, opt := range opts {
		opt(options)
	}

	conn, err := clickhouse.Open(options)
	if err != nil {
		l.Errorw("failed to open ClickHouse connection", zap.Error(err))
		return nil, fmt.Errorf("open clickhouse connection: %w", err)
	}

	// Verify connection with ping
	if err := verifyConnection(ctx, conn, l); err != nil {
		conn.Close()
		return nil, err
	}

	// Also create sql.DB for standard database/sql operations
	db := clickhouse.OpenDB(options)
	applyPoolConfig(env, db)

	ch := &ClickHouse{
		Conn: conn,
		DB:   db,
	}

	l.Infow("ClickHouse connected successfully")

	return ch, nil
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
}

// verifyConnection pings the ClickHouse server to ensure connectivity.
func verifyConnection(ctx context.Context, conn clickhouse.Conn, l logger.Log) error {
	pingCtx, cancel := context.WithTimeout(ctx, defaultPingTimeout)
	defer cancel()

	if err := conn.Ping(pingCtx); err != nil {
		l.Errorw("failed to ping ClickHouse server", zap.Error(err))
		return fmt.Errorf("verify clickhouse connection: %w", err)
	}

	return nil
}

// Close gracefully closes the ClickHouse connection.
func (ch *ClickHouse) Close() error {
	var errs []error

	if ch != nil {
		if ch.Conn != nil {
			if err := ch.Conn.Close(); err != nil {
				errs = append(errs, err)
			}
		}
		if ch.DB != nil {
			if err := ch.DB.Close(); err != nil {
				errs = append(errs, err)
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("close clickhouse: %v", errs)
	}
	return nil
}
