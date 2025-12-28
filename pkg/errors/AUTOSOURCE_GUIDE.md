# AutoSource() - Automatic Source Tracking

## Muammo

Avval har safar error yaratganda file path va function nomini qo'lda yozish kerak edi:

```go
return domain.User{}, apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase,
    "failed to build SQL query").
    WithField("file", "internal/repo/persistent/postgres/user/client/get.go"). // ❌ Qo'lda yozish kerak
    WithField("function", "Get").  // ❌ Qo'lda yozish kerak
    WithField("operation", "build_query")
```

## Yechim - AutoSource()

Endi `AutoSource()` funksiyasi runtime stackdan avtomatik file va function nomini oladi:

```go
return domain.User{}, apperrors.AutoSource(
    apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase,
        "failed to build SQL query")).  // ✅ file va function avtomatik!
    WithField("operation", "build_query")
```

## Qanday Ishlaydi?

`AutoSource()` funksiyasi Go'ning `runtime` paketidan foydalanib:

1. **Stack trace**ni oladi
2. **File path**dan faqat project ichidagi qismini ajratadi
3. **Function name**ni short formatda qaytaradi

### Runtime Magic

```go
func AutoSource(err *AppError) *AppError {
    // skip = 1 means we skip AutoSource itself and get the actual caller
    file, function := GetCaller(1)
    return err.
        WithField("file", file).
        WithField("function", function)
}

func GetCaller(skip int) (file string, function string) {
    pc, fullPath, _, ok := runtime.Caller(skip + 1)
    
    // Extract relative path: /Users/.../project/internal/repo/... → internal/repo/...
    if idx := strings.Index(fullPath, "/internal/"); idx != -1 {
        file = fullPath[idx+1:]  // "internal/repo/persistent/postgres/user/client/get.go"
    }
    
    // Extract function name: ...client.(*Repo).Get → Repo.Get
    fn := runtime.FuncForPC(pc)
    funcName = extractShortName(fn.Name())  // "Repo.Get"
    
    return file, funcName
}
```

## Ishlatish - Repository Layer

### Avvalgi usul (Manual):
```go
func (r *Repo) Get(ctx context.Context, filter UserFilter) (domain.User, error) {
    // ...
    if err == pgx.ErrNoRows {
        return domain.User{}, apperrors.NewRepoError(ctx, apperrors.ErrRepoNotFound,
            "user not found in database").
            WithField("file", "internal/repo/persistent/postgres/user/client/get.go").  // ❌
            WithField("function", "Get").  // ❌
            WithField("table", "users")
    }
}
```

### Yangi usul (AutoSource):
```go
func (r *Repo) Get(ctx context.Context, filter UserFilter) (domain.User, error) {
    // ...
    if err == pgx.ErrNoRows {
        return domain.User{}, apperrors.AutoSource(
            apperrors.NewRepoError(ctx, apperrors.ErrRepoNotFound,
                "user not found in database")).  // ✅ Avtomatik!
            WithField("table", "users")
    }
}
```

## Ishlatish - Use Case Layer

### Avvalgi usul:
```go
func (uc *UseCase) GetByID(ctx context.Context, id int64) (domain.User, error) {
    user, err := uc.repo.User.Client.GetByID(ctx, id)
    if err != nil {
        return domain.User{}, apperrors.MapRepoToServiceError(ctx, err).
            WithField("file", "internal/usecase/user/client/get_by_id.go").  // ❌
            WithField("function", "GetByID").  // ❌
            WithField("operation", "get_user_by_id")
    }
    return user, nil
}
```

### Yangi usul:
```go
func (uc *UseCase) GetByID(ctx context.Context, id int64) (domain.User, error) {
    user, err := uc.repo.User.Client.GetByID(ctx, id)
    if err != nil {
        return domain.User{}, apperrors.AutoSource(
            apperrors.MapRepoToServiceError(ctx, err)).  // ✅ Avtomatik!
            WithField("operation", "get_user_by_id")
    }
    return user, nil
}
```

## Ishlatish - Controller Layer

Controller layerda ham ishlatish mumkin, lekin odatda qo'lda yozganingiz ma'qul, chunki endpoint va method qo'shish kerak:

```go
func (c *Controller) Get(ctx *gin.Context) {
    // ...
    handlerErr := apperrors.AutoSource(
        apperrors.MapServiceToHandlerError(ctx.Request.Context(), err))
    
    // Qo'shimcha HTTP context qo'shamiz
    handlerErr.WithField("endpoint", ctx.Request.URL.Path).
        WithField("method", ctx.Request.Method)
    
    // LOG
    c.l.Errorw("failed to get user", zap.Error(handlerErr), ...)
}
```

## Output Formati

### File Path:
- Full path: `/Users/mrb/Downloads/go-clean-template-master/internal/repo/persistent/postgres/user/client/get.go`
- AutoSource: `internal/repo/persistent/postgres/user/client/get.go` ✅

### Function Name:
- Full name: `github.com/evrone/go-clean-template/internal/repo/persistent/postgres/user/client.(*Repo).Get`
- AutoSource: `Repo.Get` ✅

## Log Output Misoli

```json
{
  "level": "error",
  "msg": "failed to get user",
  "error_code": "4004",
  
  "Repository Layer": {
    "file": "internal/repo/persistent/postgres/user/client/get.go",  // ✅ Avtomatik!
    "function": "Repo.Get",  // ✅ Avtomatik!
    "table": "users",
    "filter_id": 12345
  },
  
  "Service Layer": {
    "file": "internal/usecase/user/client/get_by_id.go",  // ✅ Avtomatik!
    "function": "UseCase.GetByID",  // ✅ Avtomatik!
    "operation": "get_user_by_id"
  },
  
  "Handler Layer": {
    "file": "internal/controller/restapi/v1/user/client/get.go",  // ✅ Avtomatik!
    "function": "Controller.Get",  // ✅ Avtomatik!
    "endpoint": "/api/v1/users/12345",
    "method": "GET"
  }
}
```

## Boshqa Helper Funksiyalar

### WithCaller()
Skip levelni o'zingiz ko'rsatasiz:

```go
// Skip 0 = WithCaller o'zidan keyin kelgan funksiya
err := apperrors.WithCaller(
    apperrors.NewRepoError(ctx, code, msg), 
    0  // skip level
)
```

### GetCaller()
Faqat ma'lumot olish uchun:

```go
file, function := apperrors.GetCaller(0)
fmt.Printf("Current location: %s::%s\n", file, function)
// Output: internal/repo/.../get.go::Repo.Get
```

## Afzalliklari

### ✅ DRY (Don't Repeat Yourself)
File path va function nomini har safar yozmaslikka to'g'ri keladi.

### ✅ Xatosiz
Qo'lda yozishda typo bo'lishi mumkin, AutoSource() doim to'g'ri.

### ✅ Refactoring-Friendly
File yoki function nomini o'zgartirsangiz, kod avtomatik yangilanadi.

### ✅ Qisqa va Tushunarli
Kod ancha qisqaradi va o'qish osonroq bo'ladi.

## Performance

`runtime.Caller()` ancha tez ishlaydi va error flow'da (error scenario'da) performance muammo emas. Error flow odatda rare case bo'lgani uchun, bir necha mikrosekund overhead muammo emas.

## Xulosa

`AutoSource()` - bu magic helper funksiya bo'lib, file path va function nomini stackdan avtomatik oladi. Bu sizning kodingizni:

- ✅ Qisqaroq qiladi
- ✅ Xatosiz qiladi  
- ✅ Refactoring-friendly qiladi
- ✅ Ancha tushunarli qiladi

**Hammayerda `AutoSource()` dan foydalaning!** 🚀
