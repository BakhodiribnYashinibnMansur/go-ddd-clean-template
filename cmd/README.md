# Command Line Interface (`cmd`)

This directory contains the main entry points for the application. Each subdirectory represents a compilable binary.

## Directory Structure

### `app/`
Contains the `main.go` file for the primary web application server.
- **Purpose**: Initializes configuration, logger, and the application instance.
- **Usage**: `go run ./cmd/app`

### `migration/`
Contains the entry point for running database migrations.
- **Purpose**: Handles schema updates for the PostgreSQL database.
- **Usage**: Typically invoked via Makefile or CI/CD pipelines to apply migrations.

### `seeder/`
Contains the entry point for seeding the database with initial or test data.
- **Purpose**: Populates the database with default roles, permissions, and dummy users for development/testing.
- **Usage**: `go run ./cmd/seeder`

## Best Practices
- Keep `main` functions minimal. They should primarily focus on wiring up dependencies and starting the application (handling signals for graceful shutdown).
- Avoid putting business logic here; delegate to `internal/app` or specific `usecase` layers.
