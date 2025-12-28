# Error Handling Guide

## Umumiy Qoidalar

### Layerlar bo'yicha Javobgarlik

1. **Repository Layer** - Faqat errorni qaytaradi, LOG YOZMAYDI!
2. **Use Case/Service Layer** - Errorni map qiladi va contextni qo'shadi, LOG YOZMAYDI!
3. **Controller/Handler Layer** - Errorni LOG qiladi va response qaytaradi

## Repository Layer

### Qoidalar:
- ❌ LOG yozmaslik
- ✅ Error type to'g'ri aniqlash (ErrRepoNotFound, ErrRepoDatabase, etc)
- ✅ Context ma'lumotlari qo'shish (file, function, table, etc)
- ✅ pgx errorlarni to'g'ri handle qilish

### Misol - Get metodi:

```go
func (r *Repo) Get(ctx context.Context, filter UserFilter) (domain.User, error) {
	// ... query building ...
	
	err = r.pool.QueryRow(ctx, sql, args...).Scan(...)
	if err != nil {
		// Check for "no rows" error
		if err == pgx.ErrNoRows {
			return domain.User{}, apperrors.NewRepoError(ctx, apperrors.ErrRepoNotFound,
				"user not found in database").
				WithField("file", "internal/repo/persistent/postgres/user/client/get.go").
				WithField("function", "Get").
				WithField("table", "users").
				WithField("filter_id", filter.ID).
				WithDetails("No user record exists with the given filter criteria")
		}
		
		// Other database errors - NO LOGGING
		return domain.User{}, apperrors.WrapRepoError(ctx, err, apperrors.ErrRepoDatabase,
			"failed to query user from database").
			WithField("file", "internal/repo/persistent/postgres/user/client/get.go").
			WithField("function", "Get").
			WithField("table", "users")
	}
	
	return u, nil
}
```

### Misol - Create metodi (Constraint handling):

```go
func (r *Repo) Create(ctx context.Context, u domain.User) error {
	_, err = r.pool.Exec(ctx, sql, args...)
	if err != nil {
		// Check for constraint violation
		if pgErr, ok := err.(*pgconn.PgError); ok {
			// 23505 = unique_violation
			if pgErr.Code == "23505" {
				return apperrors.NewRepoError(ctx, apperrors.ErrRepoAlreadyExists,
					"user already exists").
					WithField("file", "internal/repo/persistent/postgres/user/client/create.go").
					WithField("function", "Create").
					WithField("table", "users").
					WithField("constraint", pgErr.ConstraintName).
					WithDetails("A user with this phone or username already exists")
			}
			
			// Other constraint violations
			if strings.HasPrefix(pgErr.Code, "23") {
				return apperrors.NewRepoError(ctx, apperrors.ErrRepoConstraint,
					"database constraint violation").
					WithField("constraint", pgErr.ConstraintName).
					WithField("pg_code", pgErr.Code).
					WithDetails(pgErr.Message)
			}
		}
		
		// Generic database error
		return apperrors.WrapRepoError(ctx, err, apperrors.ErrRepoDatabase,
			"failed to insert user into database").
			WithField("file", "...").
			WithField("function", "Create")
	}
	return nil
}
```

## Use Case/Service Layer

### Qoidalar:
- ❌ LOG yozmayslik
- ✅ Repository errorni Service errorga map qilish
- ✅ Business logic contextni qo'shish
- ✅ Validation errorlarni handle qilish

### Misol - GetByID:

```go
func (uc *UseCase) GetByID(ctx context.Context, id int64) (domain.User, error) {
	user, err := uc.repo.User.Client.GetByID(ctx, id)
	if err != nil {
		// Map repository error to service error - NO LOGGING
		return domain.User{}, apperrors.MapRepoToServiceError(ctx, err).
			WithField("file", "internal/usecase/user/client/get_by_id.go").
			WithField("function", "GetByID").
			WithField("operation", "get_user_by_id").
			WithField("user_id", id)
	}
	return user, nil
}
```

### Misol - Business Logic Validation:

```go
func (uc *UseCase) UpdateUser(ctx context.Context, u domain.User) error {
	// Validation
	if u.Phone == "" {
		return apperrors.NewServiceError(ctx, apperrors.ErrServiceValidation,
			"phone number is required").
			WithField("file", "internal/usecase/user/client/update.go").
			WithField("function", "UpdateUser").
			WithField("operation", "validate_phone").
			WithDetails("Phone number cannot be empty")
	}
	
	// Call repository
	err := uc.repo.User.Client.Update(ctx, u)
	if err != nil {
		return apperrors.MapRepoToServiceError(ctx, err).
			WithField("file", "internal/usecase/user/client/update.go").
			WithField("function", "UpdateUser").
			WithField("operation", "update_user").
			WithField("user_id", u.ID)
	}
	
	return nil
}
```

## Controller/Handler Layer

### Qoidalar:
- ✅ BU YERDA LOG YOZILADI! (zap.Error(), zap.String(), zap.Int64(), etc)
- ✅ Service errorni Handler errorga map qilish
- ✅ HTTP contextni qo'shish (endpoint, method, request_id)
- ✅ To'g'ri HTTP status code va user message qaytarish

### Misol - GET endpoint:

```go
func (c *Controller) Get(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		// Invalid input from client - LOG HERE
		handlerErr := apperrors.NewHandlerError(ctx.Request.Context(), 
			apperrors.ErrHandlerBadRequest, "invalid user id format").
			WithField("file", "internal/controller/restapi/v1/user/client/get.go").
			WithField("function", "Get").
			WithField("endpoint", ctx.Request.URL.Path).
			WithField("method", ctx.Request.Method).
			WithField("param_id", ctx.Param("id"))

		c.l.Errorw("failed to parse user ID",
			zap.Error(handlerErr),
			zap.String("error_code", handlerErr.Code),
			zap.String("error_type", handlerErr.Type),
			zap.Int("http_status", handlerErr.HTTPStatus),
			zap.String("param_id", ctx.Param("id")),
		)

		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid user id", nil, false)
		return
	}

	user, err := c.u.User.Client.GetByID(ctx.Request.Context(), id)
	if err != nil {
		// Map service error to handler error
		handlerErr := apperrors.MapServiceToHandlerError(ctx.Request.Context(), err)

		// Add handler layer context
		handlerErr.WithField("file", "internal/controller/restapi/v1/user/client/get.go").
			WithField("function", "Get").
			WithField("endpoint", ctx.Request.URL.Path).
			WithField("method", ctx.Request.Method).
			WithField("user_id", id)

		// LOG ONCE - with all layer information
		c.l.Errorw("failed to get user",
			zap.Error(handlerErr),
			zap.String("error_code", handlerErr.Code),
			zap.String("error_type", handlerErr.Type),
			zap.Int("http_status", handlerErr.HTTPStatus),
			zap.String("user_message", handlerErr.UserMsg),
			zap.Int64("user_id", id),
			zap.String("endpoint", ctx.Request.URL.Path),
			zap.String("method", ctx.Request.Method),
		)

		response.ControllerResponse(ctx, handlerErr.HTTPStatus, handlerErr.UserMsg, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, user, nil, true)
}
```

### Misol - POST endpoint:

```go
func (c *Controller) Create(ctx *gin.Context) {
	var body domain.User
	if err := ctx.ShouldBindJSON(&body); err != nil {
		// Invalid JSON - LOG HERE
		handlerErr := apperrors.NewHandlerError(ctx.Request.Context(), 
			apperrors.ErrHandlerBadRequest, "invalid request body format").
			WithField("file", "internal/controller/restapi/v1/user/client/create.go").
			WithField("function", "Create").
			WithField("endpoint", ctx.Request.URL.Path).
			WithField("method", ctx.Request.Method)

		c.l.Errorw("failed to bind request body",
			zap.Error(handlerErr),
			zap.String("error_code", handlerErr.Code),
			zap.Error(err),
		)

		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid request body", nil, false)
		return
	}

	err := c.u.User.Client.Create(ctx.Request.Context(), body)
	if err != nil {
		handlerErr := apperrors.MapServiceToHandlerError(ctx.Request.Context(), err)
		
		handlerErr.WithField("file", "internal/controller/restapi/v1/user/client/create.go").
			WithField("function", "Create").
			WithField("endpoint", ctx.Request.URL.Path).
			WithField("method", ctx.Request.Method)

		// LOG ONCE
		c.l.Errorw("failed to create user",
			zap.Error(handlerErr),
			zap.String("error_code", handlerErr.Code),
			zap.String("error_type", handlerErr.Type),
			zap.Int("http_status", handlerErr.HTTPStatus),
			zap.String("user_message", handlerErr.UserMsg),
		)

		response.ControllerResponse(ctx, handlerErr.HTTPStatus, handlerErr.UserMsg, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusCreated, nil, nil, true)
}
```

## Zap Logger Uchun Type Safety

**ESDA TUTING:** Loglarni doim zap.TYPE() funksiyalari bilan yozing:

```go
c.l.Errorw("message",
	zap.Error(err),           // error type uchun
	zap.String("key", val),   // string uchun
	zap.Int("key", val),      // int uchun
	zap.Int64("key", val),    // int64 uchun
	zap.Bool("key", val),     // bool uchun
	zap.Any("key", val),      // boshqa typelar uchun
)
```

## Error Code Structure

- **2xxx** - Repository Layer errors
  - 2001 - Not Found
  - 2002 - Already Exists
  - 2003 - Database Error
  - 2004 - Timeout
  - 2005 - Connection Error
  - 2006 - Transaction Error
  - 2007 - Constraint Violation

- **3xxx** - Service Layer errors
  - 3001 - Invalid Input
  - 3002 - Validation Error
  - 3003 - Not Found
  - 3004 - Already Exists
  - 3005 - Unauthorized
  - 3006 - Forbidden
  - 3007 - Conflict
  - 3008 - Business Rule Violation

- **4xxx/5xxx** - Handler Layer errors
  - 4000 - Bad Request
  - 4001 - Unauthorized
  - 4003 - Forbidden
  - 4004 - Not Found
  - 4009 - Conflict
  - 5000 - Internal Error

## Error Flow Diagrami

```
┌─────────────────────────────────────────────┐
│  Client Request                             │
└─────────────────┬───────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────┐
│  Controller/Handler Layer                   │
│  - Validate input                           │
│  - Call service                             │
│  - ✅ LOG ERROR HERE (zap.Errorw)          │
│  - Map to HTTP response                     │
└─────────────────┬───────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────┐
│  Use Case/Service Layer                     │
│  - Business logic                           │
│  - Map repo → service error                 │
│  - ❌ NO LOGGING                            │
│  - Add business context                     │
└─────────────────┬───────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────┐
│  Repository Layer                           │
│  - Database operations                      │
│  - Create typed errors                      │
│  - ❌ NO LOGGING                            │
│  - Add DB context (table, constraint, etc)  │
└─────────────────┬───────────────────────────┘
                  │
                  ▼
           Database/Cache
```

## Xulosa

1. **Repository** - errorni yaratib qaytaradi, LOG YOZMAYDI
2. **Service** - errorni map qiladi va contextni qo'shadi, LOG YOZMAYDI
3. **Controller** - errorni LOG qiladi (zap.TYPE() bilan) va response qaytaradi

Bu yondashuv:
- ✅ Bir xil errorni bir marta log qiladi
- ✅ Har bir layerdan to'liq contextni to'playdi
- ✅ Type-safe logging (zap.TYPE())
- ✅ To'g'ri HTTP status code va user messagelarni qaytaradi
