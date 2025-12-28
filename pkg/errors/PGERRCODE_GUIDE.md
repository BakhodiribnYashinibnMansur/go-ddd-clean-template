# PostgreSQL Error Codes - Usage Guide

## Overview

`pgerrcode.go` contains all PostgreSQL error code constants (SQLSTATE codes). These constants can be used to check specific PostgreSQL errors in your code.

## Constants

All PostgreSQL error codes are available as constants:

```go
import apperrors "github.com/evrone/go-clean-template/pkg/errors"

// Most common error codes:
apperrors.UniqueViolation              // "23505"
apperrors.ForeignKeyViolation          // "23503"
apperrors.NotNullViolation             // "23502"
apperrors.DeadlockDetected             // "40P01"
apperrors.QueryCanceled                // "57014"
apperrors.LockNotAvailable             // "55P03"
```

## Usage in HandlePgError

The `HandlePgError` function already uses these codes internally, but you can also use them for custom checks:

```go
func (r *Repo) Create(ctx context.Context, user domain.User) error {
    _, err := r.pool.Exec(ctx, sql, args...)
    if err != nil {
        var pgErr *pgconn.PgError
        if errors.As(err, &pgErr) {
            // Check specific error codes
            switch pgErr.Code {
            case apperrors.UniqueViolation:
                // Handle duplicate specially
                return apperrors.HandlePgError(ctx, err, "users", map[string]any{
                    "username": user.Username,
                    "conflict": "username_already_exists",
                })
            case apperrors.ForeignKeyViolation:
                // Handle FK violation
                return apperrors.HandlePgError(ctx, err, "users", map[string]any{
                    "detail": "referenced record not found",
                })
            default:
                // Use standard handler
                return apperrors.HandlePgError(ctx, err, "users", nil)
            }
        }
        return apperrors.HandlePgError(ctx, err, "users", nil)
    }
    return nil
}
```

## Helper Functions

Check error code classes quickly:

```go
var pgErr *pgconn.PgError
if errors.As(err, &pgErr) {
    // Check by class
    if apperrors.IsIntegrityConstraintViolation(pgErr.Code) {
        // Handle any constraint violation (23xxx)
    }
    
    if apperrors.IsConnectionException(pgErr.Code) {
        // Handle connection errors (08xxx)
    }
    
    if apperrors.IsTransactionRollback(pgErr.Code) {
        // Handle transaction rollback (40xxx)
    }
    
    if apperrors.IsDeadlock(pgErr.Code) {
        // Retry logic for deadlocks
    }
}
```

## Common Error Code Classes

### Class 23 — Integrity Constraint Violations
```go
apperrors.UniqueViolation         // 23505 - Duplicate entry
apperrors.ForeignKeyViolation     // 23503 - FK constraint fails
apperrors.NotNullViolation        // 23502 - NULL value not allowed
apperrors.CheckViolation          // 23514 - CHECK constraint fails
apperrors.ExclusionViolation      // 23P01 - Exclusion constraint fails
```

### Class 08 — Connection Exceptions
```go
apperrors.ConnectionException     // 08000 - General connection error
apperrors.ConnectionFailure       // 08006 - Connection failure
apperrors.ConnectionDoesNotExist  // 08003 - Connection doesn't exist
```

### Class 40 — Transaction Rollback
```go
apperrors.DeadlockDetected        // 40P01 - Deadlock detected
apperrors.SerializationFailure    // 40001 - Serialization failure
apperrors.TransactionRollback     // 40000 - Transaction rollback
```

### Class 42 — Syntax/Access Errors
```go
apperrors.UndefinedTable          // 42P01 - Table doesn't exist
apperrors.UndefinedColumn         // 42703 - Column doesn't exist
apperrors.InsufficientPrivilege   // 42501 - Permission denied
apperrors.SyntaxError             // 42601 - SQL syntax error
```

### Class 55 — Object State Errors
```go
apperrors.LockNotAvailable        // 55P03 - Lock timeout
apperrors.ObjectInUse             // 55006 - Object is being used
```

### Class 57 — Operator Intervention
```go
apperrors.QueryCanceled           // 57014 - Query was canceled
apperrors.AdminShutdown           // 57P01 - Database shutting down
apperrors.DatabaseDropped         // 57P04 - Database was dropped
```

## Advanced Example: Retry Logic

```go
func (r *Repo) UpdateWithRetry(ctx context.Context, user domain.User) error {
    maxRetries := 3
    
    for attempt := 0; attempt < maxRetries; attempt++ {
        err := r.Update(ctx, user)
        if err == nil {
            return nil
        }
        
        // Check if it's a deadlock
        var pgErr *pgconn.PgError
        if errors.As(err, &pgErr) {
            if pgErr.Code == apperrors.DeadlockDetected {
                // Retry after a short delay
                time.Sleep(time.Millisecond * 100 * time.Duration(attempt+1))
                continue
            }
        }
        
        // Not a deadlock, return error
        return apperrors.HandlePgError(ctx, err, "users", map[string]any{
            "user_id": user.ID,
            "attempt": attempt + 1,
        })
    }
    
    return apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase,
        "max retries exceeded for deadlock")
}
```

## Example: Custom Validation

```go
func (r *Repo) Delete(ctx context.Context, id int64) error {
    _, err := r.pool.Exec(ctx, "DELETE FROM users WHERE id = $1", id)
    if err != nil {
        var pgErr *pgconn.PgError
        if errors.As(err, &pgErr) {
            // Check if deletion failed due to FK constraint
            if pgErr.Code == apperrors.ForeignKeyViolation {
                return apperrors.AutoSource(
                    apperrors.NewRepoError(ctx, apperrors.ErrRepoConstraint,
                        "cannot delete user with existing dependencies")).
                    WithField("user_id", id).
                    WithField("constraint_type", "foreign_key").
                    WithDetails("User has related records that must be deleted first")
            }
        }
        
        return apperrors.HandlePgError(ctx, err, "users", map[string]any{
            "user_id": id,
        })
    }
    return nil
}
```

## All Error Classes

| Class | Name | Helper Function |
|-------|------|----------------|
| 00 | Successful Completion | - |
| 01 | Warning | - |
| 02 | No Data | - |
| 08 | Connection Exception | `IsConnectionException()` |
| 22 | Data Exception | `IsDataException()` |
| 23 | Integrity Constraint | `IsIntegrityConstraintViolation()` |
| 25 | Invalid Transaction State | `IsInvalidTransactionState()` |
| 40 | Transaction Rollback | `IsTransactionRollback()` |
| 42 | Syntax/Access Error | `IsSyntaxErrorOrAccessRuleViolation()` |
| 53 | Insufficient Resources | `IsInsufficientResources()` |
| 54 | Program Limit Exceeded | `IsProgramLimitExceeded()` |
| 55 | Object Not In State | `IsObjectNotInPrerequisiteState()` |
| 57 | Operator Intervention | `IsOperatorIntervention()` |
| 58 | System Error | `IsSystemError()` |
| HV | Foreign Data Wrapper | `IsForeignDataWrapperError()` |
| P0 | PL/pgSQL Error | `IsPLpgSQLError()` |
| XX | Internal Error | `IsInternalError()` |

## When to Use Constants vs HandlePgError

### ✅ Use Constants When:
- You need special logic for specific error codes
- Implementing retry logic for deadlocks
- Custom error messages for specific violations
- Business logic depends on error type

### ✅ Use HandlePgError When:
- Standard error handling is sufficient
- You want automatic error mapping
- No special logic needed for specific codes
- Keeping code simple and clean

## Example: Combined Approach

```go
func (r *Repo) Create(ctx context.Context, order domain.Order) error {
    _, err := r.pool.Exec(ctx, insertSQL, order.UserID, order.Total)
    if err != nil {
        var pgErr *pgconn.PgError
        if errors.As(err, &pgErr) {
            // Special handling for specific codes
            switch pgErr.Code {
            case apperrors.ForeignKeyViolation:
                // User doesn't exist
                return apperrors.AutoSource(
                    apperrors.NewRepoError(ctx, apperrors.ErrRepoConstraint,
                        "user not found")).
                    WithField("user_id", order.UserID).
                    WithDetails("Cannot create order for non-existent user")
                    
            case apperrors.UniqueViolation:
                // Order already exists
                return apperrors.AutoSource(
                    apperrors.NewRepoError(ctx, apperrors.ErrRepoAlreadyExists,
                        "order already exists")).
                    WithField("order_id", order.ID).
                    WithDetails("Duplicate order detected")
            }
        }
        
        // For all other errors, use standard handler
        return apperrors.HandlePgError(ctx, err, "orders", map[string]any{
            "user_id": order.UserID,
            "total":   order.Total,
        })
    }
    return nil
}
```

## Benefits

- ✅ **Type-Safe**: All error codes are constants
- ✅ **Discoverable**: IDE autocomplete shows all codes
- ✅ **Documented**: Each constant has a comment
- ✅ **Helper Functions**: Quick class checks
- ✅ **Centralized**: Single source of truth
- ✅ **Compatible**: Works with HandlePgError

## Summary

PostgreSQL error codes are now available as typed constants in your error package. Use them when you need specific error handling logic, or rely on `HandlePgError` for automatic handling.

**Best of both worlds!** 🚀
