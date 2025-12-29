# SQLite SQLC Configuration

This directory contains SQLC configuration and query files for SQLite.

## Setup

1. Install sqlc:
```bash
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

2. Generate Go code:
```bash
cd internal/repo/persistent/sqlite/sqlc
sqlc generate
```

## Files

- `sqlc.yaml` - SQLC configuration file for SQLite
- `query.sql` - SQLite queries with SQLC annotations (uses `?` placeholders)
- Generated files will appear here after running `sqlc generate`

## Usage

The generated code provides type-safe database operations:

```go
queries := sqlc.New(db)
user, err := queries.GetUser(ctx, userID)
```

## Schema Location

Database schema is located at: `migrations/sqlite/`

## Notes

- SQLite uses `?` for parameter placeholders
- SQLite supports RETURNING clause for INSERT operations
- Ideal for embedded, mobile, or lightweight applications
