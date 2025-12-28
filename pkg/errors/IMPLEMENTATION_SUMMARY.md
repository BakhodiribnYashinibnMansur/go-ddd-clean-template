# Layered Error Handling - Summary

## Qanday ishlaydi?

Sizning proyektingizda uch qatlamli (layered) error handling tizimi joriy etilgan:

### 1. Repository Layer (internal/repo/persistent/postgres/user/client/)
**Fayl:** get.go, create.go, update.go, delete.go

**Vazifasi:**
- ❌ **LOG YOZMAYDI** - Loglarni controllerda yozamiz
- ✅ Database errorlarni to'g'ri aniqlaydi
- ✅ Context qo'shadi (file, function, table, operation)
- ✅ Postgres errorlarni handle qiladi (pgx.ErrNoRows, constraint violations, etc)

**Errorlar:**
- `apperrors.ErrRepoNotFound` - Record topilmadi
- `apperrors.ErrRepoAlreadyExists` - Unique constraint violation
- `apperrors.ErrRepoDatabase` - Umumiy database errorlari
- `apperrors.ErrRepoConstraint` - Boshqa constraint violations

**Misol:**
```go
if err == pgx.ErrNoRows {
    return domain.User{}, apperrors.NewRepoError(ctx, apperrors.ErrRepoNotFound,
        "user not found in database").
        WithField("file", "...").
        WithField("function", "Get").
        WithField("table", "users")
}
```

### 2. Use Case/Service Layer (internal/usecase/user/client/)
**Fayl:** get_by_id.go, create.go, update.go, delete.go

**Vazifasi:**
- ❌ **LOG YOZMAYDI** - Loglarni controllerda yozamiz
- ✅ Repository errorlarni Service errorlarga map qiladi
- ✅ Business logic contextini qo'shadi
- ✅ Validation qiladi (kerak bo'lsa)

**Error Mapping:**
```go
err := uc.repo.User.Client.GetByID(ctx, id)
if err != nil {
    return domain.User{}, apperrors.MapRepoToServiceError(ctx, err).
        WithField("file", "...").
        WithField("function", "GetByID").
        WithField("operation", "get_user_by_id").
        WithField("user_id", id)
}
```

**Mapping qoidalari:**
- `ErrRepoNotFound` → `ErrServiceNotFound`
- `ErrRepoAlreadyExists` → `ErrServiceAlreadyExists`
- `ErrRepoConstraint` → `ErrServiceConflict`
- Boshqalar → `ErrServiceDependency`

### 3. Controller/Handler Layer (internal/controller/restapi/v1/user/client/)
**Fayl:** get.go, create.go, update.go, delete.go

**Vazifasi:**
- ✅ **BU YERDA LOG YOZILADI!** - `zap.Error()`, `zap.String()`, `zap.Int64()` bilan
- ✅ Service errorlarni Handler errorlarga map qiladi
- ✅ HTTP contextini qo'shadi (endpoint, method, request_id)
- ✅ To'g'ri HTTP status code va user message qaytaradi

**Logging Pattern:**
```go
c.l.Errorw("failed to get user",
    zap.Error(handlerErr),                    // Error obyekti
    zap.String("error_code", handlerErr.Code), // Code
    zap.String("error_type", handlerErr.Type), // Type
    zap.Int("http_status", handlerErr.HTTPStatus), // HTTP status
    zap.String("user_message", handlerErr.UserMsg), // User uchun xabar
    zap.Int64("user_id", id),                  // ID
    zap.String("endpoint", ctx.Request.URL.Path), // Endpoint
)
```

**Mapping qoidalari:**
- `ErrServiceNotFound` → `ErrHandlerNotFound` (404)
- `ErrServiceInvalidInput/Validation` → `ErrHandlerBadRequest` (400)
- `ErrServiceUnauthorized` → `ErrHandlerUnauthorized` (401)
- `ErrServiceForbidden` → `ErrHandlerForbidden` (403)
- `ErrServiceConflict/AlreadyExists` → `ErrHandlerConflict` (409)
- Boshqalar → `ErrHandlerInternal` (500)

## Error Code Struktura

### Repository Layer (2xxx)
- `2001` - Not Found
- `2002` - Already Exists
- `2003` - Database Error
- `2004` - Timeout
- `2005` - Connection Error
- `2006` - Transaction Error
- `2007` - Constraint Violation
- `2099` - Unknown Error

### Service Layer (3xxx)
- `3001` - Invalid Input
- `3002` - Validation Error
- `3003` - Not Found
- `3004` - Already Exists
- `3005` - Unauthorized
- `3006` - Forbidden
- `3007` - Conflict
- `3008` - Business Rule Violation
- `3009` - Dependency Error
- `3099` - Unknown Error

### Handler Layer (4xxx/5xxx)
- `4000` - Bad Request
- `4001` - Unauthorized
- `4003` - Forbidden
- `4004` - Not Found
- `4009` - Conflict
- `5000` - Internal Error
- `5099` - Unknown Error

## Afzalliklari

### ✅ Bitta Log Entry
Bir xil error bir marta log qilinadi, controller layerda:
```
Repository → Service → Handler
   (silent)   (silent)   (LOG!)
```

### ✅ To'liq Trace
Bitta log barcha layerlardan contextni ko'rsatadi:
```json
{
  "level": "error",
  "msg": "failed to get user",
  
  "Repository Layer": {
    "file": "internal/repo/persistent/postgres/user/client/get.go",
    "function": "Get",
    "table": "users",
    "user_id": "12345"
  },
  
  "Service Layer": {
    "file": "internal/usecase/user/client/get_by_id.go",
    "function": "GetByID",
    "operation": "get_user_by_id"
  },
  
  "Handler Layer": {
    "file": "internal/controller/restapi/v1/user/client/get.go",
    "function": "Get",
    "endpoint": "/api/v1/users/12345",
    "method": "GET",
    "error_code": "4004",
    "http_status": 404
  }
}
```

### ✅ Type-Safe Logging
Zap logger doim type-safe funksiyalar bilan:
- `zap.Error(err)` - error type
- `zap.String(key, val)` - string type
- `zap.Int(key, val)` - int type
- `zap.Int64(key, val)` - int64 type
- `zap.Bool(key, val)` - bool type
- `zap.Any(key, val)` - boshqa typelar

### ✅ To'g'ri HTTP Response
Error typeiga qarab to'g'ri HTTP status code va user message:
- 404 - "Foydalanuvchi topilmadi"
- 409 - "Allaqachon mavjud"
- 400 - "Noto'g'ri ma'lumot"
- 500 - "Ichki xatolik"

## Qo'llanmalar

Batafsil ma'lumot uchun quyidagi fayllarni o'qing:

1. **ERROR_HANDLING_GUIDE.md** - Har bir layer uchun to'liq qo'llanma
2. **LOGGING_GUIDE.md** - Logging best practices
3. **README.md** - pkg/errors paketi haqida
4. **examples/** - Amaliy misollar

## Umumiy Qoidalar

1. ❌ Repository layerda LOG YOZMANG
2. ❌ Service layerda LOG YOZMANG
3. ✅ Faqat Controller layerda LOG YOZING
4. ✅ Doim `zap.TYPE()` funksiyalardan foydalaning
5. ✅ Har bir layerda context qo'shing (`WithField()`)
6. ✅ Error type va HTTP status to'g'ri mapping qiling
7. ✅ User-friendly messagelar qaytaring

## Joriy etilgan fayllar

### Repository Layer
- ✅ internal/repo/persistent/postgres/user/client/get.go
- ✅ internal/repo/persistent/postgres/user/client/create.go
- ✅ internal/repo/persistent/postgres/user/client/update.go
- ✅ internal/repo/persistent/postgres/user/client/delete.go

### Use Case Layer
- ✅ internal/usecase/user/client/get_by_id.go
- ✅ internal/usecase/user/client/create.go
- ✅ internal/usecase/user/client/update.go
- ✅ internal/usecase/user/client/delete.go

### Controller Layer
- ✅ internal/controller/restapi/v1/user/client/get.go
- ✅ internal/controller/restapi/v1/user/client/create.go
- ✅ internal/controller/restapi/v1/user/client/update.go
- ✅ internal/controller/restapi/v1/user/client/delete.go

## Keyingi Qadamlar

Qolgan metodlarni ham shu patternda yangilash kerak:
- [ ] User session metodlari
- [ ] Auth metodlari (login, logout, etc)
- [ ] Boshqa entitylar (agar mavjud bo'lsa)

Va esda tuting: **Faqat controllerda log yozing, zap.TYPE() funksiyalardan foydalaning!**
