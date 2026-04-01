# Error Logging Package

Complete error logging system for tracking and monitoring application errors in the database.

## 📁 Structure

```
├── consts/errors.go                                    # Error code constants
├── internal/repo/persistent/postgres/system_error/     # Database repository
│   ├── README.md                                       # Detailed documentation
│   ├── repo.go                                         # Repository initialization
│   ├── system_error.go                                 # CRUD operations
│   └── system_error_test.go                            # Integration tests
├── pkg/errorx/                                         # Error logging utilities
│   ├── logger.go                                       # Base error logger
│   ├── http.go                                         # HTTP & use case loggers
│   └── logger_test.go                                  # Usage examples
└── migrations/postgres/20260101040000_create_system_errors.sql  # Database migration
```

## 🚀 Quick Start

### 1. Database Migration

The `system_errors` table is already created via migration:
```bash
# Migration is already in: migrations/postgres/20260101040000_create_system_errors.sql
# Run migrations if not already applied
goose -dir migrations/postgres postgres "your-connection-string" up
```

### 2. Basic Usage

```go
import (
    "gct/consts"
    "gct/pkg/errorx"
)

// Initialize error logger (in your app setup)
errorLogger := errorx.NewErrorLogger(
    repo.Persistent.Postgres.SystemError,
    logger,
)

// Log a simple error
err := errorLogger.LogErrorSimple(ctx, 
    consts.ErrCodeUserNotFound, 
    "User not found", 
    err,
)

// Log with full context
err = errorLogger.LogError(ctx, errorx.LogErrorInput{
    Code:        consts.ErrCodeAuthFailed,
    Message:     "Authentication failed",
    Err:         err,
    Severity:    consts.SeverityError,
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

### 3. HTTP Handler Usage

```go
import "gct/pkg/errorx"

type Handler struct {
    errorLogger *errorx.HTTPErrorLogger
}

func NewHandler(repo *repo.Repo, logger logger.Log) *Handler {
    return &Handler{
        errorLogger: errorx.NewHTTPErrorLogger(
            repo.Persistent.Postgres.SystemError,
            logger,
        ),
    }
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    httpCtx := errorx.ExtractHTTPContext(r)
    
    user, err := h.useCase.Authenticate(ctx, username, password)
    if err != nil {
        // Log authentication error with full HTTP context
        h.errorLogger.LogAuthError(ctx, err, httpCtx, username)
        http.Error(w, "Authentication failed", http.StatusUnauthorized)
        return
    }
    
    // Success...
}
```

### 4. Use Case Usage

```go
import "gct/pkg/errorx"

type UserUseCase struct {
    errorLogger *errorx.UseCaseErrorLogger
}

func NewUserUseCase(repo *repo.Repo, logger logger.Log) *UserUseCase {
    return &UserUseCase{
        errorLogger: errorx.NewUseCaseErrorLogger(
            repo.Persistent.Postgres.SystemError,
            logger,
            "user-service",
        ),
    }
}

func (uc *UserUseCase) CreateUser(ctx context.Context, input CreateUserInput) (*User, error) {
    user, err := uc.repo.Persistent.Postgres.User.Create(ctx, input)
    if err != nil {
        // Log database error
        uc.errorLogger.LogDatabaseError(ctx, err, "CREATE", "user")
        return nil, fmt.Errorf("failed to create user: %w", err)
    }
    
    return user, nil
}
```

## 📊 Error Codes

All error codes are defined in `consts/errors.go`:

### Categories:
- **Authentication & Authorization**: `AUTH_FAILED`, `INVALID_TOKEN`, `TOKEN_EXPIRED`, etc.
- **User Errors**: `USER_NOT_FOUND`, `USER_ALREADY_EXISTS`, `USER_NOT_APPROVED`, etc.
- **Database Errors**: `DATABASE_ERROR`, `QUERY_TIMEOUT`, `CONNECTION_LOST`, etc.
- **Validation Errors**: `VALIDATION_FAILED`, `INVALID_INPUT`, `MISSING_FIELD`, etc.
- **External Service Errors**: `EXTERNAL_SERVICE_ERROR`, `API_TIMEOUT`, `MYID_ERROR`, etc.
- **File & Storage Errors**: `FILE_NOT_FOUND`, `FILE_UPLOAD_FAILED`, `STORAGE_ERROR`, etc.
- **Cache Errors**: `CACHE_ERROR`, `CACHE_MISS`, `CACHE_SET_FAILED`, etc.
- **Business Logic Errors**: `INVALID_OPERATION`, `RESOURCE_NOT_FOUND`, `DUPLICATE_ENTRY`, etc.
- **System Errors**: `INTERNAL_ERROR`, `CONFIG_ERROR`, `PANIC`, etc.

### Severity Levels:
- `WARN` - Recoverable issues
- `ERROR` - Standard errors
- `FATAL` - Critical errors affecting service
- `PANIC` - System-level panics

## 🔍 Querying Errors

```go
// Get error by ID
systemErr, err := repo.Persistent.Postgres.SystemError.GetByID(ctx, errorID)

// List errors with filters
errors, err := repo.Persistent.Postgres.SystemError.List(ctx, systemerror.ListFilter{
    Code:       stringPtr(consts.ErrCodeAuthFailed),
    Severity:   stringPtr(consts.SeverityError),
    IsResolved: boolPtr(false),
    Limit:      100,
    Offset:     0,
})

// Mark error as resolved
err := repo.Persistent.Postgres.SystemError.MarkAsResolved(ctx, errorID, resolverUserID)
```

## 📝 Database Schema

The `system_errors` table includes:

| Column | Type | Description |
|--------|------|-------------|
| `id` | UUID | Primary key |
| `code` | VARCHAR(64) | Error code (indexed) |
| `message` | TEXT | Error message |
| `stack_trace` | TEXT | Stack trace at error time |
| `metadata` | — | Additional context (stored in `entity_metadata` table) |
| `severity` | VARCHAR(16) | ERROR, FATAL, PANIC, WARN (indexed) |
| `service_name` | VARCHAR(64) | Service that generated the error |
| `request_id` | UUID | Associated request ID (indexed) |
| `user_id` | UUID | Associated user ID |
| `ip_address` | INET | Client IP address |
| `path` | VARCHAR(255) | Request path |
| `method` | VARCHAR(8) | HTTP method |
| `is_resolved` | BOOLEAN | Resolution status (indexed) |
| `resolved_at` | TIMESTAMP | Resolution timestamp |
| `resolved_by` | UUID | User who resolved the error |
| `created_at` | TIMESTAMP | Error timestamp (indexed) |

## ✅ Best Practices

1. **Use Descriptive Error Codes**: Use constants from `consts/errors.go`
2. **Include Context**: Always include request ID, user ID when available
3. **Add Metadata**: Store relevant debugging information
4. **Choose Appropriate Severity**: 
   - Use `WARN` for recoverable issues
   - Use `ERROR` for standard errors
   - Use `FATAL` for critical errors
   - Use `PANIC` for system-level panics
5. **Resolve Errors**: Mark errors as resolved after fixing
6. **Monitor Regularly**: Query unresolved errors periodically
7. **Stack Traces**: Automatically captured for debugging

## 🎯 Integration Points

The error logging system is integrated into:
- ✅ Repository layer (`internal/repo/persistent/postgres/system_error/`)
- ✅ Main postgres repository (`internal/repo/persistent/postgres/repo.go`)
- ✅ Error constants (`consts/errors.go`)
- ✅ Error logging utilities (`pkg/errorx/`)
- ✅ Database migration (`migrations/postgres/20260101040000_create_system_errors.sql`)

## 📚 Additional Documentation

For more detailed information, see:
- `internal/repo/persistent/postgres/system_error/README.md` - Repository documentation
- `pkg/errorx/logger_test.go` - Usage examples
- `internal/repo/persistent/postgres/system_error/system_error_test.go` - Integration tests

## 🔧 Testing

```bash
# Run unit tests
go test ./pkg/errorx/...

# Run integration tests (requires database)
go test ./internal/repo/persistent/postgres/system_error/... -v

# Build verification
go build ./...
```

---

**Status**: ✅ Fully implemented and tested
**Build**: ✅ Passing
**Migration**: ✅ Available
