# PostgreSQL Package

PostgreSQL connection pool with functional options pattern.

## Features

- Connection pooling with pgx/v5
- Functional options pattern for flexible configuration
- Production and development environment presets
- Connection health monitoring
- Zap logger integration with SQL tracing
- Squirrel query builder integration

## Usage

### Basic Usage

```go
import (
    "context"
    
    "gct/config"
    "gct/pkg/logger"
    "gct/pkg/postgres"
)

ctx := context.Background()
l := logger.New("info")

cfg := config.Postgres{
    Host:     "localhost",
    Port:     5432,
    User:     "user",
    Password: "password",
    Name:     "database",
    SSLMode:  "disable",
}

// Create connection with default settings
pg, err := postgres.New(ctx, "dev", cfg, l)
if err != nil {
    panic(err)
}
defer pg.Close()
```

### Using Custom Options

```go
import "time"

pg, err := postgres.New(ctx, "production", cfg, l,
    postgres.WithMaxConns(100),
    postgres.WithMinConns(20),
    postgres.WithMaxConnLifetime(1*time.Hour),
    postgres.WithMaxConnIdleTime(10*time.Minute),
    postgres.WithHealthCheckPeriod(1*time.Minute),
    postgres.WithConnectTimeout(5*time.Second),
    postgres.WithStatementTimeout(30*time.Second),
    postgres.WithApplicationName("my-service"),
)
```

## Available Options

### `WithMaxConns(maxConns int32)`
Sets the maximum number of connections in the pool.

**Default:**
- Production: 50
- Development: 8

**Example:**
```go
postgres.WithMaxConns(100)
```

### `WithMinConns(minConns int32)`
Sets the minimum number of idle connections in the pool.

**Default:**
- Production: 10
- Development: 3

**Example:**
```go
postgres.WithMinConns(20)
```

### `WithMaxConnLifetime(d time.Duration)`
Sets the maximum lifetime of a connection before it's closed and recreated.

**Default:** 10 hours

**Example:**
```go
postgres.WithMaxConnLifetime(1 * time.Hour)
```

### `WithMaxConnIdleTime(d time.Duration)`
Sets the maximum time a connection can remain idle before being closed.

**Default:** 30 minutes

**Example:**
```go
postgres.WithMaxConnIdleTime(10 * time.Minute)
```

### `WithHealthCheckPeriod(d time.Duration)`
Sets the interval for connection health checks.

**Default:** 5 minutes

**Example:**
```go
postgres.WithHealthCheckPeriod(1 * time.Minute)
```

### `WithConnectTimeout(d time.Duration)`
Sets the timeout for establishing new connections.

**Default:** 10 seconds

**Example:**
```go
postgres.WithConnectTimeout(5 * time.Second)
```

### `WithStatementTimeout(d time.Duration)`
Sets the maximum execution time for SQL statements.

**Default:** Not set

**Example:**
```go
postgres.WithStatementTimeout(30 * time.Second)
```

### `WithApplicationName(name string)`
Sets the application name visible in PostgreSQL's `pg_stat_activity`.

**Default:** Not set

**Example:**
```go
postgres.WithApplicationName("my-service")
```

### `WithTraceLogLevel(level tracelog.LogLevel)`
Sets the pgx trace log level.

**Available levels:**
- `tracelog.LogLevelNone`
- `tracelog.LogLevelError`
- `tracelog.LogLevelWarn`
- `tracelog.LogLevelInfo`
- `tracelog.LogLevelDebug`
- `tracelog.LogLevelTrace`

**Default:** `tracelog.LogLevelTrace`

**Example:**
```go
import "github.com/jackc/pgx/v5/tracelog"

postgres.WithTraceLogLevel(tracelog.LogLevelInfo)
```

## Environment Presets

The package includes optimized presets for different environments:

### Production
- MaxConns: 50
- MinConns: 10
- MaxConnLifetime: 10 hours
- MaxConnIdleTime: 30 minutes

### Development
- MaxConns: 8
- MinConns: 3
- MaxConnLifetime: 10 hours
- MaxConnIdleTime: 30 minutes

## Query Builder

The package includes Squirrel query builder with Dollar placeholder format:

```go
query, args, err := pg.Builder.
    Select("id", "name", "email").
    From("users").
    Where("deleted_at = ?", 0).
    ToSql()
```

## Pool Statistics

Get current pool statistics:

```go
stats := pg.Stats()
if stats != nil {
    fmt.Printf("Total connections: %d\n", stats.TotalConns())
    fmt.Printf("Idle connections: %d\n", stats.IdleConns())
    fmt.Printf("Acquired connections: %d\n", stats.AcquiredConns())
}
```

## Best Practices

1. **Always close connections**: Use `defer pg.Close()` after creating a connection
2. **Use context**: Always pass context for cancellation and timeout control
3. **Monitor pool stats**: Regularly check pool statistics in production
4. **Set statement timeout**: Prevent long-running queries from blocking resources
5. **Customize for your workload**: Adjust pool settings based on your application's needs
