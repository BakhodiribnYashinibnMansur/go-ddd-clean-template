// Package main provides a standalone CLI tool for managing database schema migrations.
// It applies incremental SQL changes to the Postgres database to keep the schema in sync with the codebase.
package main

import (
	"log"

	"gct/config"
	"gct/internal/shared/infrastructure/db/postgres"
	"gct/internal/shared/infrastructure/logger"

	"go.uber.org/zap"
)

// main initializes the migration engine and applies any pending updates.
func main() {
	// 1. Load application configuration to retrieve database connection strings.
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config initialization error: %s", err)
	}

	// 2. Initialize a minimal logger for the migration process.
	l := logger.New(cfg.Log.Level)

	// 3. Initialize the Postgres migrator using SQL scripts from the "migrations" directory.
	migrator, err := postgres.NewMigration(cfg.Database.Postgres, "migrations")
	if err != nil {
		l.Fatalw("Failed to initialize migrator", zap.Error(err))
	}
	defer migrator.Close()

	// 4. Execute "Up" migrations.
	// This will identify all pending .sql files and execute them sequentially.
	err = migrator.Up()
	if err != nil {
		l.Fatalw("Failed to apply migrations", zap.Error(err))
	}

	l.Infow("Database migrations applied successfully")
}
