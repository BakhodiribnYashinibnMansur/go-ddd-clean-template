package container

import (
	"context"
	"database/sql"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // pgx stdlib driver
	"github.com/pressly/goose"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"gct/config"
)

// RunPostgresTestContainer is a function that runs a postgres test container
func RunPostgresTestContainer(cfg config.Database, schemaPath string) *pgxpool.Pool {
	ctx := context.Background()

	postgresContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
	)
	if err != nil {
		log.Fatalf("failed to start postgres container: %v", err)
	}

	dbURL, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("failed to get connection string: %v", err)
	}

	log.Printf("Database URL: %s", dbURL)

	// Create pgx connection pool
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("failed to create connection pool: %v", err)
	}

	// Test connection
	err = pool.Ping(ctx)
	if err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	// Run migrations with Goose if schema path provided
	if schemaPath != "" {
		log.Printf("Running migrations from: %s", schemaPath)

		// Create sql.DB for Goose compatibility using pgx stdlib driver
		db, err := sql.Open("pgx", dbURL)
		if err != nil {
			log.Fatalf("failed to open database for migrations: %v", err)
		}
		defer db.Close()

		if err := goose.SetDialect("postgres"); err != nil {
			log.Fatalf("failed to set goose dialect: %v", err)
		}
		if err := goose.Up(db, schemaPath); err != nil {
			log.Fatalf("failed to run migrations: %v", err)
		}
		log.Printf("Migrations completed successfully")
	}

	log.Printf("PostgreSQL test container ready")

	return pool
}
