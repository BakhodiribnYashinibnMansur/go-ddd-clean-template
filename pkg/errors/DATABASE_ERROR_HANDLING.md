# Database Error Handling - Complete Guide

## Overview

Centralized error handling for all database and cache systems:
- ✅ **PostgreSQL** - Complete error handling with 300+ error codes
- ✅ **MySQL** - Complete error handling with 50+ error codes
- ✅ **Redis** - Complete error handling with 10+ error types

## Quick Links

- [PostgreSQL Guide](./postgres.go) - Full PostgreSQL error handling
- [MySQL Guide](./MYSQL_GUIDE.md) - MySQL error codes and examples
- [Redis Guide](./REDIS_GUIDE.md) - Redis error types and examples
- [AutoSource Guide](./AUTOSOURCE_GUIDE.md) - Automatic file/function tracking
- [Error Handling Guide](./ERROR_HANDLING_GUIDE.md) - Layered architecture
- [Logging Guide](./LOGGING_GUIDE.md) - Logging best practices

## Usage

### PostgreSQL

```go
import apperrors "github.com/evrone/go-clean-template/pkg/errors"

func (r *Repo) Create(ctx context.Context, user domain.User) error {
    _, err := r.pool.Exec(ctx, sql, args...)
    if err != nil {
        return apperrors.HandlePgError(ctx, err, "users", map[string]any{
            "username": user.Username,
            "phone":    user.Phone,
        })
    }
    return nil
}
```

**Handles:**
- Unique violations → `ErrRepoAlreadyExists`
- Foreign key violations → `ErrRepoConstraint`
- Not null violations → `ErrRepoConstraint`
- Connection errors → `ErrRepoConnection`
- Deadlocks → `ErrRepoDatabase`
- Timeouts → `ErrRepoTimeout`
- And 300+ other PostgreSQL errors!

### MySQL

```go
import apperrors "github.com/evrone/go-clean-template/pkg/errors"

func (r *Repo) Create(ctx context.Context, user domain.User) error {
    _, err := r.db.ExecContext(ctx, query, args...)
    if err != nil {
        return apperrors.HandleMySQLError(ctx, err, "users", map[string]any{
            "username": user.Username,
            "email":    user.Email,
        })
    }
    return nil
}
```

**Handles:**
- 1062 - Duplicate entry → `ErrRepoAlreadyExists`
- 1452 - Foreign key fails → `ErrRepoConstraint`
- 1048 - Column cannot be null → `ErrRepoConstraint`
- 1205 - Lock timeout → `ErrRepoTimeout`
- 1213 - Deadlock → `ErrRepoDatabase`
- And 50+ other MySQL error codes!

### Redis

```go
import apperrors "github.com/evrone/go-clean-template/pkg/errors"

func (r *CacheRepo) Get(ctx context.Context, key string) (string, error) {
    val, err := r.client.Get(ctx, key).Result()
    if err != nil {
        return "", apperrors.HandleRedisError(ctx, err, key, map[string]any{
            "operation": "get",
        })
    }
    return val, nil
}
```

**Handles:**
- `redis.Nil` → `ErrRepoNotFound`
- Connection errors → `ErrRepoConnection`
- Timeout errors → `ErrRepoTimeout`
- WRONGTYPE → `ErrRepoDatabase`
- OOM (out of memory) → `ErrRepoDatabase`
- And 10+ other Redis error types!

## Error Flow Architecture

```
┌─────────────────────┐
│  Repository Layer   │
│  (PostgreSQL/MySQL/ │
│      Redis)         │
└──────────┬──────────┘
           │
           │ Returns AppError (NO LOGGING)
           │
           ▼
┌─────────────────────┐
│   Service Layer     │
└──────────┬──────────┘
           │
           │ Maps to Service Error (NO LOGGING)
           │
           ▼
┌─────────────────────┐
│  Controller Layer   │
│                     │
│  ✅ LOGS HERE!     │
│  with zap.Errorw() │
└─────────────────────┘
```

## Error Codes

### Repository Layer (2xxx)
- `2001` - **ErrRepoNotFound** - Record/Key not found
- `2002` - **ErrRepoAlreadyExists** - Duplicate entry
- `2003` - **ErrRepoDatabase** - Generic database error
- `2004` - **ErrRepoTimeout** - Operation timeout
- `2005` - **ErrRepoConnection** - Connection error
- `2006` - **ErrRepoTransaction** - Transaction error
- `2007` - **ErrRepoConstraint** - Constraint violation

### Service Layer (3xxx)
- Automatically mapped from repository errors
- See [Error Handling Guide](./ERROR_HANDLING_GUIDE.md)

### Handler Layer (4xxx/5xxx)
- Automatically mapped from service errors
- See [Error Handling Guide](./ERROR_HANDLING_GUIDE.md)

## Features

### ✅ AutoSource - Automatic File/Function Tracking

No need to manually specify file and function names:

```go
// Before ❌
return apperrors.NewRepoError(ctx, code, msg).
    WithField("file", "internal/repo/.../get.go").
    WithField("function", "Repo.Get")

// After ✅  
return apperrors.AutoSource(
    apperrors.NewRepoError(ctx, code, msg))
```

File and function are extracted from runtime stack automatically!

### ✅ Centralized Error Handling

One function per database type handles ALL errors:

```go
// PostgreSQL
apperrors.HandlePgError(ctx, err, table, extraFields)

// MySQL
apperrors.HandleMySQLError(ctx, err, table, extraFields)

// Redis
apperrors.HandleRedisError(ctx, err, key, extraFields)
```

### ✅ Type-Safe Logging

Only in controller layer with zap:

```go
c.l.Errorw("operation failed",
    zap.Error(handlerErr),
    zap.String("error_code", handlerErr.Code),
    zap.Int("http_status", handlerErr.HTTPStatus),
    zap.Int64("user_id", userID),
)
```

### ✅ Rich Context

Every error carries context from all layers:

```json
{
  "level": "error",
  "msg": "failed to create user",
  
  "Repository": {
    "file": "internal/repo/postgres/user/create.go",
    "function": "Repo.Create",
    "table": "users",
    "username": "john",
    "mysql_code": 1062,
    "sql_state": "23000"
  },
  
  "Service": {
    "file": "internal/usecase/user/create.go",
    "function": "UseCase.Create",
    "operation": "create_user"
  },
  
  "Handler": {
    "file": "internal/controller/user/create.go",
    "function": "Controller.Create",
    "endpoint": "/api/v1/users",
    "method": "POST",
    "http_status": 409
  }
}
```

## Error Mapping

```
┌──────────────────────────────────────────────────┐
│           PostgreSQL / MySQL / Redis             │
└──────────────────┬───────────────────────────────┘
                   │
         HandlePgError() / HandleMySQLError() / HandleRedisError()
                   │
                   ▼
┌──────────────────────────────────────────────────┐
│          Repository Errors (2xxx)                │
│  - ErrRepoNotFound        → 2001                 │
│  - ErrRepoAlreadyExists   → 2002                 │
│  - ErrRepoDatabase        → 2003                 │
│  - ErrRepoTimeout         → 2004                 │
│  - ErrRepoConnection      → 2005                 │
│  - ErrRepoConstraint      → 2007                 │
└──────────────────┬───────────────────────────────┘
                   │
          MapRepoToServiceError()
                   │
                   ▼
┌──────────────────────────────────────────────────┐
│           Service Errors (3xxx)                  │
│  - ErrServiceNotFound     → 3003                 │
│  - ErrServiceAlreadyExists→ 3004                 │
│  - ErrServiceConflict     → 3007                 │
│  - ErrServiceDependency   → 3009                 │
└──────────────────┬───────────────────────────────┘
                   │
       MapServiceToHandlerError()
                   │
                   ▼
┌──────────────────────────────────────────────────┐
│          Handler Errors (4xxx/5xxx)              │
│  - ErrHandlerNotFound     → 404                  │
│  - ErrHandlerConflict     → 409                  │
│  - ErrHandlerInternal     → 500                  │
└──────────────────────────────────────────────────┘
```

## Complete Example

### Repository Layer (PostgreSQL)

```go
package postgres

func (r *UserRepo) Create(ctx context.Context, user domain.User) error {
    _, err := r.pool.Exec(ctx, insertSQL, user.Username, user.Email)
    if err != nil {
        // NO LOGGING - just return structured error
        return apperrors.HandlePgError(ctx, err, "users", map[string]any{
            "username": user.Username,
            "email":    user.Email,
        })
    }
    return nil
}
```

### Service Layer

```go
package usecase

func (uc *UserUseCase) Create(ctx context.Context, user domain.User) error {
    err := uc.repo.Create(ctx, user)
    if err != nil {
        // NO LOGGING - just map and return
        return apperrors.AutoSource(
            apperrors.MapRepoToServiceError(ctx, err)).
            WithField("operation", "create_user")
    }
    return nil
}
```

### Controller Layer

```go
package controller

func (c *UserController) Create(ctx *gin.Context) {
    var req CreateUserRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        // ... handle binding error
    }
    
    user := domain.User{Username: req.Username, Email: req.Email}
    err := c.usecase.Create(ctx.Request.Context(), user)
    
    if err != nil {
        // Map to handler error
        handlerErr := apperrors.AutoSource(
            apperrors.MapServiceToHandlerError(ctx.Request.Context(), err))
        
        // ✅ LOG HERE - only in controller!
        c.logger.Errorw("failed to create user",
            zap.Error(handlerErr),
            zap.String("error_code", handlerErr.Code),
            zap.Int("http_status", handlerErr.HTTPStatus),
            zap.String("username", req.Username),
            zap.String("endpoint", ctx.Request.URL.Path),
        )
        
        // Return HTTP response
        ctx.JSON(handlerErr.HTTPStatus, gin.H{
            "error": handlerErr.UserMsg,
            "code":  handlerErr.Code,
        })
        return
    }
    
    ctx.JSON(201, gin.H{"message": "user created"})
}
```

## Benefits Summary

| Feature | PostgreSQL | MySQL | Redis |
|---------|-----------|-------|-------|
| Centralized Handler | ✅ | ✅ | ✅ |
| Auto Error Detection | ✅ | ✅ | ✅ |
| Error Code Mapping | ✅ (300+) | ✅ (50+) | ✅ (10+) |
| AutoSource | ✅ | ✅ | ✅ |
| No Repo Logging | ✅ | ✅ | ✅ |
| Rich Context | ✅ | ✅ | ✅ |
| Type-Safe | ✅ | ✅ | ✅ |

## Documentation

- **PostgreSQL**: See `postgres.go` - 800+ lines with full error codes
- **MySQL**: See `MYSQL_GUIDE.md` - Complete guide with examples
- **Redis**: See `REDIS_GUIDE.md` - Complete guide with examples
- **AutoSource**: See `AUTOSOURCE_GUIDE.md` - Automatic source tracking
- **Architecture**: See `ERROR_HANDLING_GUIDE.md` - Layered error handling

## Production Ready! 🚀

All three database/cache error handlers are:
- ✅ **Battle-tested** patterns
- ✅ **Consistent** API across all databases
- ✅ **Comprehensive** error coverage
- ✅ **Well-documented** with guides and examples
- ✅ **Type-safe** with zap logging
- ✅ **Clean** separation of concerns

Start using today:
```go
// Just replace your error returns with:
return apperrors.HandlePgError(ctx, err, table, extraFields)
return apperrors.HandleMySQLError(ctx, err, table, extraFields)
return apperrors.HandleRedisError(ctx, err, key, extraFields)
```

**That's it!** ✨
