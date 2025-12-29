# MySQL SQLC Configuration

This directory contains SQLC configuration and query files for MySQL.

## Setup

1. Install sqlc:
```bash
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

2. Generate Go code:
```bash
cd internal/repo/persistent/mysql/sqlc
sqlc generate
```

## Files

- `sqlc.yaml` - SQLC configuration file for MySQL
- `query.sql` - MySQL queries with SQLC annotations (uses `?` placeholders)
- Generated files will appear here after running `sqlc generate`

## Usage

The generated code provides type-safe database operations:

```go
queries := sqlc.New(db)
user, err := queries.GetUser(ctx, userID)
```

## Schema Location

Database schema is located at: `migrations/mysql/`

## Notes

- MySQL uses `?` for parameter placeholders instead of `$1, $2`
- `execresult` is used for INSERT operations to get LastInsertId
