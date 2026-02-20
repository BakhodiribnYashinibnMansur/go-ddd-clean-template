# System Error Logging

This package provides functionality to log errors to the database for monitoring and debugging purposes.

## Features

- **Database Storage**: All errors are stored in the `system_errors` table
- **Severity Levels**: ERROR, FATAL, PANIC, WARN
- **Context Information**: Request ID, User ID, IP Address, Path, Method
- **Stack Traces**: Automatic stack trace capture
- **Metadata**: Store additional error context as JSONB
- **Resolution Tracking**: Mark errors as resolved with timestamp and resolver

## Usage

### Basic Error Logging

```go
import (
    "context"
    "gct/pkg/errorx"
)

// Initialize error logger
errorLogger := errorx.NewErrorLogger(repo.Persistent.Postgres.SystemError, logger)

// Log a simple error
err := errorLogger.LogErrorSimple(ctx, "USER_NOT_FOUND", "User not found", err)

// Log with full context
err := errorLogger.LogError(ctx, errorx.LogErrorInput{
    Code:        "AUTH_FAILED",
    Message:     "Authentication failed",
    Err:         err,
    Severity:    "ERROR",
    ServiceName: "auth-service",
    RequestID:   &requestID,
    UserID:      &userID,
    IPAddress:   &ipAddr,
    Path:        &path,
    Method:      &method,
    Metadata: map[string]any{
        "attempt": 3,
        "reason":  "invalid_password",
    },
})
```

### Severity Levels

```go
// Log error (default)
errorLogger.LogErrorSimple(ctx, "CODE", "message", err)

// Log fatal error
errorLogger.LogFatal(ctx, "CODE", "message", err)

// Log panic
errorLogger.LogPanic(ctx, "CODE", "message", err)

// Log warning
errorLogger.LogWarn(ctx, "CODE", "message", err)
```

### Repository Operations

```go
// Get error by ID
systemErr, err := repo.Persistent.Postgres.SystemError.GetByID(ctx, errorID)

// List errors with filters
errors, err := repo.Persistent.Postgres.SystemError.List(ctx, systemerror.ListFilter{
    Code:       stringPtr("AUTH_FAILED"),
    Severity:   stringPtr("ERROR"),
    IsResolved: boolPtr(false),
    Limit:      100,
    Offset:     0,
})

// Mark error as resolved
err := repo.Persistent.Postgres.SystemError.MarkAsResolved(ctx, errorID, resolverUserID)
```

## Database Schema

The `system_errors` table includes:

- `id` - UUID primary key
- `code` - Error code (indexed)
- `message` - Error message
- `stack_trace` - Stack trace at error time
- `metadata` - Additional context as JSONB
- `severity` - ERROR, FATAL, PANIC, WARN (indexed)
- `service_name` - Service that generated the error
- `request_id` - Associated request ID (indexed)
- `user_id` - Associated user ID
- `ip_address` - Client IP address
- `path` - Request path
- `method` - HTTP method
- `is_resolved` - Resolution status (indexed)
- `resolved_at` - Resolution timestamp
- `resolved_by` - User who resolved the error
- `created_at` - Error timestamp (indexed)

## Best Practices

1. **Use Descriptive Error Codes**: Use UPPER_SNAKE_CASE codes like `USER_NOT_FOUND`, `AUTH_FAILED`
2. **Include Context**: Always include request ID, user ID when available
3. **Add Metadata**: Store relevant debugging information in metadata
4. **Choose Appropriate Severity**: 
   - `WARN` - Recoverable issues
   - `ERROR` - Standard errors
   - `FATAL` - Critical errors affecting service
   - `PANIC` - System-level panics
5. **Resolve Errors**: Mark errors as resolved after fixing to track resolution
