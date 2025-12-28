# MySQL Error Handling Guide

## Overview

`pkg/errors/mysql.go` provides centralized error handling for all MySQL operations, similar to PostgreSQL and Redis error handling.

## Core Function

```go
func HandleMySQLError(ctx context.Context, err error, table string, extraFields map[string]any) *AppError
```

## Usage Examples

### Basic Usage

```go
// Repository method example
func (r *UserRepo) GetByID(ctx context.Context, id int64) (domain.User, error) {
    var user domain.User
    
    err := r.db.QueryRowContext(ctx, 
        "SELECT id, username, email FROM users WHERE id = ?", id).
        Scan(&user.ID, &user.Username, &user.Email)
    
    if err != nil {
        return domain.User{}, apperrors.HandleMySQLError(ctx, err, "users", map[string]any{
            "operation": "get_by_id",
            "user_id":   id,
        })
    }
    
    return user, nil
}
```

### Insert with Duplicate Check

```go
func (r *UserRepo) Create(ctx context.Context, user domain.User) error {
    result, err := r.db.ExecContext(ctx,
        "INSERT INTO users (username, email, password_hash) VALUES (?, ?, ?)",
        user.Username, user.Email, user.PasswordHash)
    
    if err != nil {
        return apperrors.HandleMySQLError(ctx, err, "users", map[string]any{
            "operation": "create_user",
            "username":  user.Username,
            "email":     user.Email,
        })
    }
    
    // ... handle result
    return nil
}
```

### Update Operation

```go
func (r *UserRepo) Update(ctx context.Context, user domain.User) error {
    _, err := r.db.ExecContext(ctx,
        "UPDATE users SET username = ?, email = ?, updated_at = NOW() WHERE id = ?",
        user.Username, user.Email, user.ID)
    
    if err != nil {
        return apperrors.HandleMySQLError(ctx, err, "users", map[string]any{
            "operation": "update_user",
            "user_id":   user.ID,
        })
    }
    
    return nil
}
```

### Delete with Foreign Key Check

```go
func (r *UserRepo) Delete(ctx context.Context, id int64) error {
    _, err := r.db.ExecContext(ctx,
        "DELETE FROM users WHERE id = ?", id)
    
    if err != nil {
        return apperrors.HandleMySQLError(ctx, err, "users", map[string]any{
            "operation": "delete_user",
            "user_id":   id,
        })
    }
    
    return nil
}
```

## Error Types Handled

### 1. **sql.ErrNoRows** → `ErrRepoNotFound`
```go
var user domain.User
err := r.db.QueryRowContext(ctx, "SELECT * FROM users WHERE id = ?", 999).
    Scan(&user.ID, &user.Username)
// Returns: ErrRepoNotFound
```

**Log Output:**
```json
{
  "error_code": "2001",
  "error_type": "REPO_NOT_FOUND",
  "message": "record not found",
  "table": "users",
  "operation": "get_by_id",
  "user_id": 999
}
```

### 2. **1062 - Duplicate Entry** → `ErrRepoAlreadyExists`
```go
_, err := r.db.ExecContext(ctx, 
    "INSERT INTO users (username, email) VALUES (?, ?)", 
    "john", "john@example.com")
// If username or email already exists
// Returns: ErrRepoAlreadyExists with mysql_code: 1062
```

### 3. **1452 - Foreign Key Constraint** → `ErrRepoConstraint`
```go
_, err := r.db.ExecContext(ctx,
    "INSERT INTO posts (user_id, title) VALUES (?, ?)",
    999, "My Post") // user_id 999 doesn't exist
// Returns: ErrRepoConstraint with mysql_code: 1452
```

### 4. **1451 - Cannot Delete Parent Row** → `ErrRepoConstraint`
```go
_, err := r.db.ExecContext(ctx,
    "DELETE FROM users WHERE id = ?", 1)
// If user has related posts
// Returns: ErrRepoConstraint with mysql_code: 1451
```

### 5. **1048 - Column Cannot Be NULL** → `ErrRepoConstraint`
```go
_, err := r.db.ExecContext(ctx,
    "INSERT INTO users (username, email) VALUES (?, ?)",
    "john", nil) // email is NOT NULL
// Returns: ErrRepoConstraint with mysql_code: 1048
```

### 6. **1146 - Table Doesn't Exist** → `ErrRepoDatabase`
```go
_, err := r.db.ExecContext(ctx, "SELECT * FROM non_existent_table")
// Returns: ErrRepoDatabase with mysql_code: 1146
```

### 7. **1054 - Unknown Column** → `ErrRepoDatabase`
```go
_, err := r.db.QueryContext(ctx, "SELECT non_existent_column FROM users")
// Returns: ErrRepoDatabase with mysql_code: 1054
```

### 8. **1205 - Lock Wait Timeout** → `ErrRepoTimeout`
```go
// In a transaction that waits for a lock
tx, _ := r.db.BeginTx(ctx, nil)
_, err := tx.ExecContext(ctx, "UPDATE users SET ... WHERE id = ?", 1)
// If another transaction holds a lock
// Returns: ErrRepoTimeout with mysql_code: 1205
```

### 9. **1213 - Deadlock Detected** → `ErrRepoDatabase`
```go
// In concurrent transactions that deadlock
// Returns: ErrRepoDatabase with mysql_code: 1213
```

### 10. **1040 - Too Many Connections** → `ErrRepoConnection`
```go
db, err := sql.Open("mysql", dsn)
// If max_connections limit reached
// Returns: ErrRepoConnection with mysql_code: 1040
```

### 11. **1045 - Access Denied** → `ErrRepoDatabase`
```go
db, err := sql.Open("mysql", "wrong_user:wrong_pass@tcp(localhost:3306)/db")
// Returns: ErrRepoDatabase with mysql_code: 1045
```

### 12. **1406 - Data Too Long** → `ErrRepoDatabase`
```go
_, err := r.db.ExecContext(ctx,
    "INSERT INTO users (username) VALUES (?)",
    strings.Repeat("a", 300)) // If VARCHAR(255)
// Returns: ErrRepoDatabase with mysql_code: 1406
```

## Complete Repository Example

```go
package mysql

import (
    "context"
    "database/sql"

    "github.com/evrone/go-clean-template/internal/domain"
    apperrors "github.com/evrone/go-clean-template/pkg/errors"
)

type UserRepo struct {
    db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
    return &UserRepo{db: db}
}

// GetByID retrieves a user by ID
func (r *UserRepo) GetByID(ctx context.Context, id int64) (domain.User, error) {
    var user domain.User
    
    query := "SELECT id, username, email, created_at FROM users WHERE id = ?"
    err := r.db.QueryRowContext(ctx, query, id).
        Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt)
    
    if err != nil {
        // Centralized error handling - NO LOGGING!
        return domain.User{}, apperrors.HandleMySQLError(ctx, err, "users", map[string]any{
            "operation": "get_by_id",
            "user_id":   id,
        })
    }
    
    return user, nil
}

// Create inserts a new user
func (r *UserRepo) Create(ctx context.Context, user domain.User) error {
    query := `
        INSERT INTO users (username, email, password_hash, created_at, updated_at)
        VALUES (?, ?, ?, NOW(), NOW())
    `
    
    result, err := r.db.ExecContext(ctx, query,
        user.Username, user.Email, user.PasswordHash)
    
    if err != nil {
        // Centralized error handling - NO LOGGING!
        return apperrors.HandleMySQLError(ctx, err, "users", map[string]any{
            "operation": "create",
            "username":  user.Username,
            "email":     user.Email,
        })
    }
    
    id, _ := result.LastInsertId()
    user.ID = id
    
    return nil
}

// Update updates an existing user
func (r *UserRepo) Update(ctx context.Context, user domain.User) error {
    query := `
        UPDATE users 
        SET username = ?, email = ?, updated_at = NOW()
        WHERE id = ?
    `
    
    _, err := r.db.ExecContext(ctx, query,
        user.Username, user.Email, user.ID)
    
    if err != nil {
        // Centralized error handling - NO LOGGING!
        return apperrors.HandleMySQLError(ctx, err, "users", map[string]any{
            "operation": "update",
            "user_id":   user.ID,
        })
    }
    
    return nil
}

// Delete soft-deletes a user
func (r *UserRepo) Delete(ctx context.Context, id int64) error {
    query := "UPDATE users SET deleted_at = NOW() WHERE id = ?"
    
    _, err := r.db.ExecContext(ctx, query, id)
    if err != nil {
        // Centralized error handling - NO LOGGING!
        return apperrors.HandleMySQLError(ctx, err, "users", map[string]any{
            "operation": "delete",
            "user_id":   id,
        })
    }
    
    return nil
}

// List retrieves users with pagination
func (r *UserRepo) List(ctx context.Context, limit, offset int) ([]domain.User, error) {
    query := `
        SELECT id, username, email, created_at 
        FROM users 
        WHERE deleted_at IS NULL
        ORDER BY id DESC
        LIMIT ? OFFSET ?
    `
    
    rows, err := r.db.QueryContext(ctx, query, limit, offset)
    if err != nil {
        // Centralized error handling - NO LOGGING!
        return nil, apperrors.HandleMySQLError(ctx, err, "users", map[string]any{
            "operation": "list",
            "limit":     limit,
            "offset":    offset,
        })
    }
    defer rows.Close()
    
    var users []domain.User
    for rows.Next() {
        var user domain.User
        if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt); err != nil {
            return nil, apperrors.HandleMySQLError(ctx, err, "users", map[string]any{
                "operation": "scan_row",
            })
        }
        users = append(users, user)
    }
    
    if err = rows.Err(); err != nil {
        return nil, apperrors.HandleMySQLError(ctx, err, "users", map[string]any{
            "operation": "rows_error",
        })
    }
    
    return users, nil
}

// Transaction example
func (r *UserRepo) TransferCredits(ctx context.Context, fromID, toID int64, amount int) error {
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return apperrors.HandleMySQLError(ctx, err, "users", map[string]any{
            "operation": "begin_tx",
        })
    }
    defer tx.Rollback()
    
    // Deduct from sender
    _, err = tx.ExecContext(ctx,
        "UPDATE users SET credits = credits - ? WHERE id = ?",
        amount, fromID)
    if err != nil {
        return apperrors.HandleMySQLError(ctx, err, "users", map[string]any{
            "operation": "deduct_credits",
            "from_id":   fromID,
            "amount":    amount,
        })
    }
    
    // Add to receiver
    _, err = tx.ExecContext(ctx,
        "UPDATE users SET credits = credits + ? WHERE id = ?",
        amount, toID)
    if err != nil {
        return apperrors.HandleMySQLError(ctx, err, "users", map[string]any{
            "operation": "add_credits",
            "to_id":     toID,
            "amount":    amount,
        })
    }
    
    if err = tx.Commit(); err != nil {
        return apperrors.HandleMySQLError(ctx, err, "users", map[string]any{
            "operation": "commit_tx",
        })
    }
    
    return nil
}
```

## Error Flow

```
Repository (MySQL)
    ↓
HandleMySQLError()  ← Detects error by MySQL error code
    ↓
Returns AppError with proper code
    ↓
Service Layer
    ↓
MapRepoToServiceError()
    ↓
Controller Layer
    ↓
MapServiceToHandlerError() + LOG with zap.Errorw()
```

## Benefits

### ✅ Automatic Error Code Detection
MySQL error codes are automatically detected and mapped to appropriate AppError types!

### ✅ Consistent Error Codes
- `ErrRepoNotFound` (2001) - sql.ErrNoRows
- `ErrRepoAlreadyExists` (2002) - MySQL 1062 (duplicate entry)
- `ErrRepoConstraint` (2007) - MySQL 1452, 1451, 1048 (constraints)
- `ErrRepoTimeout` (2004) - MySQL 1205 (lock timeout)
- `ErrRepoConnection` (2005) - MySQL 1040, 2003 (connection)
- `ErrRepoDatabase` (2003) - All other MySQL errors

### ✅ Rich Context
Every error includes:
- `table` - Table name
- `operation` - What operation failed
- `mysql_code` - MySQL error number
- `sql_state` - SQL state code
- `file` - Source file (automatic via AutoSource)
- `function` - Function name (automatic via AutoSource)
- Any extra fields you provide

### ✅ No Logging in Repository
Just like PostgreSQL and Redis handlers, MySQL handler doesn't log - only returns structured errors.

## MySQL Error Codes Quick Reference

| Code | Name | AppError | Description |
|------|------|----------|-------------|
| 1062 | ER_DUP_ENTRY | ErrRepoAlreadyExists | Duplicate entry for key |
| 1452 | ER_NO_REFERENCED_ROW_2 | ErrRepoConstraint | Foreign key constraint fails (insert) |
| 1451 | ER_ROW_IS_REFERENCED_2 | ErrRepoConstraint | Foreign key constraint fails (delete) |
| 1048 | ER_BAD_NULL_ERROR | ErrRepoConstraint | Column cannot be null |
| 1146 | ER_NO_SUCH_TABLE | ErrRepoDatabase | Table doesn't exist |
| 1054 | ER_BAD_FIELD_ERROR | ErrRepoDatabase | Unknown column |
| 1205 | ER_LOCK_WAIT_TIMEOUT | ErrRepoTimeout | Lock wait timeout exceeded |
| 1213 | ER_LOCK_DEADLOCK | ErrRepoDatabase | Deadlock found |
| 1040 | ER_CON_COUNT_ERROR | ErrRepoConnection | Too many connections |
| 1045 | ER_ACCESS_DENIED_ERROR | ErrRepoDatabase | Access denied |
| 1406 | ER_DATA_TOO_LONG | ErrRepoDatabase | Data too long for column |
| 1364 | ER_NO_DEFAULT_FOR_FIELD | ErrRepoConstraint | Field has no default value |

## Comparison: Before vs After

### ❌ Before (Manual):
```go
func (r *Repo) Create(ctx context.Context, user domain.User) error {
    _, err := r.db.ExecContext(ctx, query, user.Username, user.Email)
    if err != nil {
        if mysqlErr, ok := err.(*mysql.MySQLError); ok {
            if mysqlErr.Number == 1062 {
                r.logger.Error("duplicate entry")
                return fmt.Errorf("user already exists")
            }
            if mysqlErr.Number == 1452 {
                r.logger.Error("fk constraint")
                return fmt.Errorf("foreign key error")
            }
            // ... more checks
        }
        r.logger.Error("mysql error", zap.Error(err))
        return fmt.Errorf("database error: %w", err)
    }
    return nil
}
```

### ✅ After (Centralized):
```go
func (r *Repo) Create(ctx context.Context, user domain.User) error {
    _, err := r.db.ExecContext(ctx, query, user.Username, user.Email)
    if err != nil {
        return apperrors.HandleMySQLError(ctx, err, "users", map[string]any{
            "operation": "create",
            "username":  user.Username,
        })
    }
    return nil
}
```

**Lines saved:** 15+ lines per method! ✅

## Summary

MySQL error handling is now:
- 🎯 **Centralized** - One function handles all MySQL errors
- ✅ **Automatic** - Detects error codes automatically
- 📊 **Structured** - Returns proper AppError with codes
- 🚀 **Clean** - No logging in repository layer
- 💡 **Simple** - Just call `HandleMySQLError()`
- 🔍 **Detailed** - Includes MySQL error code and SQL state

**Production-ready for MySQL 5.7+, MySQL 8.0+!** 🚀
