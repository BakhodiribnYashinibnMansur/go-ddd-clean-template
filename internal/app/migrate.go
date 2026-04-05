//go:build migrate

package app

import (
	"database/sql"
	"time"

	"gct/config"
	"gct/internal/kernel/infrastructure/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
)

const (
	_defaultAttempts = 20
	_defaultTimeout  = time.Second
)

func init() {
	l := logger.GetLogger()

	cfg, err := config.NewConfig()
	if err != nil {
		l.Fatalw("migrate: config error", zap.Error(err))
	}

	if cfg.IsProd() {
		l.Infow("Production environment detected. Waiting 1 minute before starting migrations...")
		for i := 60; i > 0; i-- {
			l.Infow("Migration countdown", zap.Int("seconds_left", i))
			time.Sleep(time.Second)
		}
	}

	databaseURL := cfg.Database.Postgres.URL()

	var (
		attempts = _defaultAttempts
		db       *sql.DB
	)

	for attempts > 0 {
		db, err = sql.Open("pgx", databaseURL)
		if err == nil {
			err = db.Ping()
			if err == nil {
				break
			}
		}

		l.Infow("Migrate: postgres is trying to connect", zap.Int("attempts_left", attempts))
		time.Sleep(_defaultTimeout)
		attempts--
	}

	if err != nil {
		l.Fatalw("Migrate: postgres connect error", zap.Error(err))
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		l.Fatalw("Migrate: goose set dialect error", zap.Error(err))
	}

	l.Infow("Migrate: running up migrations")
	if err := goose.Up(db, "migration/postgres"); err != nil {
		l.Fatalw("Migrate: up error", zap.Error(err))
	}

	l.Infow("Migrate: up success")
}
