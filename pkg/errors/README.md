# Layer-Based Error Handling System

Professional error handling system with separate error codes for Repository, Service, and Handler layers.

## Architecture

```
┌─────────────────┐
│  Handler Layer  │  4xxx, 5xxx codes → Maps to HTTP responses
├─────────────────┤
│  Service Layer  │  3xxx codes → Business logic errors
├─────────────────┤
│Repository Layer │  2xxx codes → Database/data source errors
└─────────────────┘
```

## Error Structure

```go
type AppError struct {
    Type        string         // Error type (e.g., "USER_NOT_FOUND")
    Code        string         // Numeric code (e.g., "4041")
    Message     string         // Developer message
    HTTPStatus  int            // HTTP status code
    UserMsg     string         // User-friendly message
    Details     string         // Detailed explanation
    Fields      map[string]any // Additional context
    Err         error          // Wrapped error
    Stack       []uintptr      // Stack trace
}
```

## Response Format

```json
{
  "status": "error",
  "statusCode": 404,
  "error": {
    "code": "4004",
    "message": "Resource not found",
    "type": "HANDLER_NOT_FOUND",
    "details": "The requested resource does not exist",
    "timestamp": "2023-12-08T12:30:45Z",
    "path": "/api/v1/users/12345",
    "method": "GET"
  }
}
```

## Layer-Specific Errors

### Repository Layer (2xxx)

Used for database and data source errors.

```go
import "github.com/evrone/go-clean-template/pkg/errors"

// Creating repository errors
err := errors.NewRepoError(ctx, errors.ErrRepoNotFound, "user not found in database").
    WithField("user_id", userID).
    WithDetails("The user record does not exist")

// Wrapping database errors
err := errors.WrapRepoError(ctx, dbErr, errors.ErrRepoDatabase, "failed to query user")
```

**Available Codes:**
- `REPO_NOT_FOUND` (2001) → HTTP 404
- `REPO_ALREADY_EXISTS` (2002) → HTTP 409
- `REPO_DATABASE_ERROR` (2003) → HTTP 500
- `REPO_TIMEOUT` (2004) → HTTP 504
- `REPO_CONNECTION` (2005) → HTTP 500
- `REPO_TRANSACTION` (2006) → HTTP 500
- `REPO_CONSTRAINT` (2007) → HTTP 409

### Service Layer (3xxx)

Used for business logic errors.

```go
// Creating service errors
err := errors.NewServiceError(ctx, errors.ErrServiceValidation, "validation failed").
    WithField("field", "email").
    WithDetails("Email format is invalid")

// Mapping repository errors
serviceErr := errors.MapRepoToServiceError(ctx, repoErr)
```

**Available Codes:**
- `SERVICE_INVALID_INPUT` (3001) → HTTP 400
- `SERVICE_VALIDATION` (3002) → HTTP 400
- `SERVICE_NOT_FOUND` (3003) → HTTP 404
- `SERVICE_ALREADY_EXISTS` (3004) → HTTP 409
- `SERVICE_UNAUTHORIZED` (3005) → HTTP 401
- `SERVICE_FORBIDDEN` (3006) → HTTP 403
- `SERVICE_CONFLICT` (3007) → HTTP 409
- `SERVICE_BUSINESS_RULE` (3008) → HTTP 400
- `SERVICE_DEPENDENCY` (3009) → HTTP 500

### Handler Layer (4xxx, 5xxx)

Used for HTTP handler errors.

```go
// Creating handler errors
err := errors.NewHandlerError(ctx, errors.ErrHandlerUnauthorized, "authentication required")

// Mapping service errors
handlerErr := errors.MapServiceToHandlerError(ctx, serviceErr)
```

**Available Codes:**
- `HANDLER_BAD_REQUEST` (4000) → HTTP 400
- `HANDLER_UNAUTHORIZED` (4001) → HTTP 401
- `HANDLER_FORBIDDEN` (4003) → HTTP 403
- `HANDLER_NOT_FOUND` (4004) → HTTP 404
- `HANDLER_CONFLICT` (4009) → HTTP 409
- `HANDLER_INTERNAL_ERROR` (5000) → HTTP 500

## Usage Examples

### 1. Repository Layer

```go
func (r *UserRepo) GetByID(ctx context.Context, id string) (*User, error) {
    var user User
    err := r.db.Where("id = ?", id).First(&user).Error
    
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, errors.NewRepoError(ctx, errors.ErrRepoNotFound, "user not found").
            WithField("id", id).
            WithField("table", "users").
            WithDetails(fmt.Sprintf("User with ID '%s' does not exist", id))
    }
    
    if err != nil {
        return nil, errors.WrapRepoError(ctx, err, errors.ErrRepoDatabase, "failed to query user")
    }
    
    return &user, nil
}
```

### 2. Service Layer

```go
func (s *UserService) GetUser(ctx context.Context, id string) (*User, error) {
    // Call repository
    user, err := s.repo.GetByID(ctx, id)
    if err != nil {
        // Map repository error to service error
        return nil, errors.MapRepoToServiceError(ctx, err)
    }
    
    // Business logic validation
    if user.Status != "active" {
        return nil, errors.NewServiceError(ctx, errors.ErrServiceForbidden, "user account is not active").
            WithField("user_id", id).
            WithField("status", user.Status).
            WithDetails("The user account has been deactivated")
    }
    
    return user, nil
}
```

### 3. Handler Layer

```go
func (h *UserHandler) GetUser(c *gin.Context) {
    userID := c.Param("id")
    
    // Call service
    user, err := h.service.GetUser(c.Request.Context(), userID)
    if err != nil {
        // Map service error to handler error
        handlerErr := errors.MapServiceToHandlerError(c.Request.Context(), err)
        response.Error(c, handlerErr)
        return
    }
    
    response.Success(c, user)
}
```

## Error Flow

```
Repository Error (REPO_NOT_FOUND, 2001)
         ↓
Service Error (SERVICE_NOT_FOUND, 3003)
         ↓
Handler Error (HANDLER_NOT_FOUND, 4004)
         ↓
HTTP Response (404 JSON)
```

## Helper Functions

### Error Checking

```go
if errors.Is(err, errors.ErrRepoNotFound) {
    // Handle not found
}

if errors.Is(err, errors.ErrServiceUnauthorized) {
    // Handle unauthorized
}
```

### Getting Error Code

```go
code := errors.GetCode(err)
if code == errors.ErrHandlerNotFound {
    // Handle 404
}
```

### Adding Context

```go
err := errors.NewRepoError(ctx, errors.ErrRepoNotFound, "record not found").
    WithField("id", recordID).
    WithField("table", "users").
    WithDetails("The specified record does not exist in the database")
```

## Response Integration

The `pkg/response` package automatically handles error mapping:

```go
func (c *Controller) UpdateUser(ctx *gin.Context) {
    user, err := c.service.UpdateUser(ctx, userID, data)
    if err != nil {
        response.Error(ctx, err)  // Automatic mapping!
        return
    }
    response.Success(ctx, user)
}
```

Response will be:
```json
{
  "status": "error",
  "statusCode": 404,
  "error": {
    "code": "4004",
    "message": "Resource not found",
    "type": "HANDLER_NOT_FOUND",
    "details": "...",
    "timestamp": "2023-12-08T12:30:45Z",
    "path": "/api/v1/users/123",
    "method": "PUT"
  }
}
```

## Logging

### Using Zap Logger

The package provides helper functions for logging errors with Zap:

```go
import (
    "github.com/evrone/go-clean-template/pkg/errors"
    "go.uber.org/zap"
)

// Log error with all fields
errors.LogError(logger, err)

// Log as warning
errors.LogWarn(logger, err)

// Log info
errors.LogInfo(logger, err, "operation failed")
```

### Log Output Examples

**Production (JSON):**
```json
{
  "level": "error",
  "timestamp": "2023-12-08T12:30:45Z",
  "error_type": "REPO_NOT_FOUND",
  "error_code": "2001",
  "http_status": 404,
  "message": "user not found in database",
  "user_message": "Record not found in database",
  "details": "The user with ID '12345' does not exist",
  "user_id": "12345",
  "table": "users"
}
```

**Development (Console):**
```
ERROR  user not found in database
  error_type: REPO_NOT_FOUND
  error_code: 2001
  http_status: 404
  user_id: 12345
  table: users
```

### Manual Logging

You can also log manually:

```go
if appErr, ok := err.(*errors.AppError); ok {
    logger.Error(appErr.Message,
        zap.String("error_type", appErr.Type),
        zap.String("error_code", appErr.Code),
        zap.Int("http_status", appErr.HTTPStatus),
        zap.String("user_message", appErr.UserMsg),
        zap.String("details", appErr.Details),
        zap.Any("fields", appErr.Fields),
        zap.Error(appErr.Err),
    )
}
```

## Examples

Run examples:
```bash
# Layer-based error handling
cd pkg/errors/examples/layers
go run main.go

# Response format examples
cd pkg/errors/examples/response_format
go run main.go

# Basic usage
cd pkg/errors
go test -v
```

## Best Practices

1. **Use layer-specific errors** - Use `REPO_*` in repositories, `SERVICE_*` in services, `HANDLER_*` in handlers
2. **Map between layers** - Use `MapRepoToServiceError()` and `MapServiceToHandlerError()`
3. **Add context** - Use `WithField()` and `WithDetails()` for debugging
4. **Wrap, don't replace** - Use `Wrap()` to preserve error chain
5. **Check errors properly** - Use `errors.Is()` instead of string comparison

## Adding New Error Codes

1. Add constants to appropriate layer file (`repository.go`, `service.go`, or `handler.go`)
2. Add to messages map
3. Update `getNumericCode()` in `errors.go`
4. Update `MapToHTTPStatus()` if needed

Example:
```go
// In repository.go
const (
    ErrRepoLocked    = "REPO_LOCKED"
    CodeRepoLocked   = "2008"
)

var repoMessages = map[string]string{
    ...
    ErrRepoLocked: "Resource is locked",
}
```
