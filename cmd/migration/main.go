package main

import (
	"log"
	"os"

	"github.com/evrone/go-clean-template/config"
	"github.com/evrone/go-clean-template/pkg/logger"
	"github.com/evrone/go-clean-template/pkg/postgres"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	l := logger.New(cfg.Log.Level)

	migrator, err := postgres.NewMigration(cfg.Database.Postgres, "migrations")
	if err != nil {
		l.Fatalw("Migration init error", zap.Error(err))
	}
	defer migrator.Close()

	// Default to status, or run Up if configured (though migrations often depend on CLI flags)
	// For this template, we'll just run Up to fulfill the requirement.
	err = migrator.Up()
	if err != nil {
		l.Fatalw("Migration Up error", zap.Error(err))
	}

	l.Infow("Migrations applied successfully")
	os.Exit(0)
}
