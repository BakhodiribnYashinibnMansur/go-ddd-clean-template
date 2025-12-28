# Logging Best Practices - Handler Layer Only

## Overview

**Important:** Logs should only be written in the **Handler Layer**. Repository and Service layers should only return errors, not log them. The handler will log once with full context from all layers.

## Why Log Only in Handler?

1. **Avoid duplicate logs** - Same error logged multiple times
2. **Single source of truth** - One log entry per request
3. **Complete context** - Handler has full request context
4. **Performance** - Less I/O operations

## Architecture

```
Repository Layer  → Returns error (no logging)
       ↓
Service Layer     → Maps & returns error (no logging)
       ↓
Handler Layer     → Logs error once + sends response
```

## Code Example

### 1. Repository Layer (NO LOGGING)

```go
// internal/repo/persistent/postgres/user.go

func (r *UserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
    var user domain.User
    err := r.psql.Pool.QueryRow(ctx, 
        "SELECT id, name, email FROM users WHERE id = $1", id).
        Scan(&user.ID, &user.Name, &user.Email)
    
    if err == pgx.ErrNoRows {
        // Just return error, NO LOGGING
        return nil, errors.NewRepoError(ctx, errors.ErrRepoNotFound, 
            "user not found in database").
            WithField("user_id", id).
            WithField("table", "users").
            WithField("file", "internal/repo/persistent/postgres/user.go").
            WithField("function", "GetByID").
            WithDetails("No user record exists with the given ID")
    }
    
    if err != nil {
        return nil, errors.WrapRepoError(ctx, err, errors.ErrRepoDatabase, 
            "failed to query user").
            WithField("user_id", id).
            WithField("file", "internal/repo/persistent/postgres/user.go").
            WithField("function", "GetByID")
    }
    
    return &user, nil
}
```

### 2. Service Layer (NO LOGGING)

```go
// internal/usecase/user/service.go

func (s *UserService) GetUser(ctx context.Context, id string) (*domain.User, error) {
    // Call repository
    user, err := s.repo.GetByID(ctx, id)
    if err != nil {
        // Map error, NO LOGGING
        return nil, errors.MapRepoToServiceError(ctx, err).
            WithField("file", "internal/usecase/user/service.go").
            WithField("function", "GetUser").
            WithField("operation", "get_user")
    }
    
    // Business logic
    if user.Status != "active" {
        return nil, errors.NewServiceError(ctx, errors.ErrServiceForbidden, 
            "user account is not active").
            WithField("user_id", id).
            WithField("status", user.Status).
            WithField("file", "internal/usecase/user/service.go").
            WithField("function", "GetUser").
            WithDetails("User account has been deactivated or suspended")
    }
    
    return user, nil
}
```

### 3. Handler Layer (LOG HERE)

```go
// internal/controller/restapi/user/handler.go

func (h *UserHandler) GetUser(c *gin.Context) {
    userID := c.Param("id")
    
    // Call service
    user, err := h.service.GetUser(c.Request.Context(), userID)
    if err != nil {
        // Map to handler error
        handlerErr := errors.MapServiceToHandlerError(c.Request.Context(), err)
        
        // Add handler layer context
        handlerErr.WithField("file", "internal/controller/restapi/user/handler.go").
            WithField("function", "GetUser").
            WithField("endpoint", c.Request.URL.Path).
            WithField("method", c.Request.Method).
            WithField("request_id", c.GetString("request_id"))
        
        // LOG ONCE - with all layer information
        errors.LogError(h.logger, handlerErr)
        
        // Send response
        response.Error(c, handlerErr)
        return
    }
    
    // Success
    response.Success(c, user)
}
```

## Log Output Example

When the above error flows through all layers, **ONLY ONE LOG** is written in the handler:

```json
{
  "level": "error",
  "timestamp": "2025-12-28T18:53:00+0500",
  "caller": "user/handler.go:42",
  "msg": "user not found in database",
  "error_type": "HANDLER_NOT_FOUND",
  "error_code": "4004",
  "http_status": 404,
  "user_message": "Resource not found",
  "details": "No user record exists with the given ID",
  
  "Repository Layer": {
    "file": "internal/repo/persistent/postgres/user.go",
    "function": "GetByID",
    "table": "users",
    "user_id": "12345"
  },
  
  "Service Layer": {
    "file": "internal/usecase/user/service.go",
    "function": "GetUser",
    "operation": "get_user"
  },
  
  "Handler Layer": {
    "file": "internal/controller/restapi/user/handler.go",
    "function": "GetUser",
    "endpoint": "/api/v1/users/12345",
    "method": "GET",
    "request_id": "req-abc-123"
  },
  
  "stacktrace": "..."
}
```

## Formatted Log (Console Mode)

```
ERROR  user not found in database
├─ Repository: internal/repo/persistent/postgres/user.go::GetByID
│  ├─ table: users
│  └─ user_id: 12345
│
├─ Service: internal/usecase/user/service.go::GetUser
│  └─ operation: get_user
│
└─ Handler: internal/controller/restapi/user/handler.go::GetUser
   ├─ endpoint: /api/v1/users/12345
   ├─ method: GET
   ├─ error_code: 4004
   └─ http_status: 404
```

## Benefits

### ✅ Single Log Entry
```
Repository → Service → Handler
   (silent)   (silent)   (LOG!)
```

### ✅ Complete Trace

One log shows the entire error journey:
- Where it started (Repository)
- How it was processed (Service)
- Where it was handled (Handler)

### ✅ Request Context

Handler has access to:
- Request ID
- HTTP method
- Endpoint path
- Headers
- User info

## Implementation Pattern

### Always Add Context in Each Layer

**Repository:**
```go
.WithField("file", "internal/repo/.../file.go").
.WithField("function", "FunctionName").
.WithField("table", "table_name")
```

**Service:**
```go
.WithField("file", "internal/usecase/.../file.go").
.WithField("function", "FunctionName").
.WithField("operation", "operation_name")
```

**Handler:**
```go
.WithField("file", "internal/controller/.../file.go").
.WithField("function", "FunctionName").
.WithField("endpoint", path).
.WithField("method", method)
```

## Helper Function

You can create a helper to automatically add file/function info:

```go
// pkg/errors/context.go

func WithSource(err *AppError, file, function string) *AppError {
    return err.
        WithField("file", file).
        WithField("function", function)
}

// Usage in repository:
return nil, errors.WithSource(
    errors.NewRepoError(ctx, errors.ErrRepoNotFound, "not found"),
    "internal/repo/persistent/postgres/user.go",
    "GetByID",
)
```

## Anti-Patterns (DON'T DO THIS)

### ❌ Logging in Repository
```go
func (r *UserRepo) GetByID(ctx context.Context, id string) (*User, error) {
    err := r.db.Get(&user, id)
    if err != nil {
        r.logger.Error("user not found", zap.String("id", id))  // DON'T!
        return nil, err
    }
}
```

### ❌ Logging in Service
```go
func (s *UserService) GetUser(ctx context.Context, id string) (*User, error) {
    user, err := s.repo.GetByID(ctx, id)
    if err != nil {
        s.logger.Error("failed to get user", zap.Error(err))  // DON'T!
        return nil, err
    }
}
```

### ✅ Correct - Only Handler Logs
```go
func (h *UserHandler) GetUser(c *gin.Context) {
    user, err := h.service.GetUser(c, id)
    if err != nil {
        errors.LogError(h.logger, err)  // ✓ CORRECT!
        response.Error(c, err)
        return
    }
}
```

## Exception: Critical Errors

For critical errors that need immediate attention, you can log in lower layers, but use different log levels:

```go
// Repository - critical database issue
if err == ErrDatabaseDown {
    r.logger.Fatal("database connection lost", zap.Error(err))  // OK for fatal
}

// Service - security issue  
if suspiciousActivity {
    s.logger.Warn("suspicious activity detected", ...)  // OK for warnings
}
```

But regular business errors should only be logged in the handler.

## Summary

1. **Repository** - Return errors with context, NO logging
2. **Service** - Map errors and add context, NO logging  
3. **Handler** - Log ONCE with full trace + send response

This creates a single, comprehensive log entry that shows the complete error path through your application.
