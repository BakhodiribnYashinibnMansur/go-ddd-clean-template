package container

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"gct/config"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // pgx stdlib driver
	"github.com/pressly/goose"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

// RunPostgresTestContainer is a function that runs a postgres test container
// RunPostgresTestContainer runs a postgres test container
func RunPostgresTestContainer(cfg config.Database, schemaPath string) (*pgxpool.Pool, testcontainers.Container, error) {
	ctx := context.Background()

	postgresContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage(PostgresqlImage),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start postgres container: %w", err)
	}

	dbURL, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, postgresContainer, fmt.Errorf("failed to get connection string: %w", err)
	}

	log.Printf("Database URL: %s", dbURL)

	// Create pgx connection pool
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, postgresContainer, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection with retry
	var pingErr error
	for range 5 {
		pingErr = pool.Ping(ctx)
		if pingErr == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}
	if pingErr != nil {
		return nil, postgresContainer, fmt.Errorf("failed to ping database after retries: %w", pingErr)
	}

	// Run migrations with Goose if schema path provided
	if schemaPath != "" {
		log.Printf("Running migrations from: %s", schemaPath)

		// Create sql.DB for Goose compatibility using pgx stdlib driver
		db, err := sql.Open("pgx", dbURL)
		if err != nil {
			return nil, postgresContainer, fmt.Errorf("failed to open database for migrations: %w", err)
		}
		defer db.Close()

		if err := goose.SetDialect("postgres"); err != nil {
			return nil, postgresContainer, fmt.Errorf("failed to set goose dialect: %w", err)
		}
		if err := goose.Up(db, schemaPath); err != nil {
			return nil, postgresContainer, fmt.Errorf("failed to run migrations: %w", err)
		}
		log.Printf("Migrations completed successfully")
	}

	log.Printf("PostgreSQL test container ready")

	return pool, postgresContainer, nil
}
