# Feature Flags

Bu loyiha [GoFeatureFlag](https://gofeatureflag.org/) yordamida feature flaglarni qo'llab-quvvatlaydi. Bu sizga kodni o'zgartirmasdan funksiyalarni yoqish/o'chirish imkonini beradi.

## Konfiguratsiya

### Environment Variables

`.env` faylida quyidagi o'zgaruvchilarni sozlang:

```bash
# Feature Flags
FEATURE_FLAG_ENABLED=true                        # Feature flaglarni yoqish/o'chirish
FEATURE_FLAG_CONFIG_PATH=./config/flags.yaml     # Flag konfiguratsiya fayli yo'li
FEATURE_FLAG_POLLING_INTERVAL=60                 # Yangilanishlarni tekshirish intervali (soniyalarda)
FEATURE_FLAG_USE_FILE=true                       # Fayl retrieverdan foydalanish
FEATURE_FLAG_USE_REDIS=false                     # Redis retrieverdan foydalanish
FEATURE_FLAG_REDIS_KEY=feature_flags             # Redis key nomi
```

### Flag Konfiguratsiya Fayli

`config/flags.yaml` faylida flaglarni aniqlang:

```yaml
enable-new-feature:
  variations:
    enabled: true
    disabled: false
  defaultRule:
    variation: disabled
  targeting:
    - name: "Beta foydalanuvchilar uchun yoqish"
      query: beta eq true
      variation: enabled
```

## Ishlatish

### 1. Oddiy Boolean Flag

```go
import (
    "gct/pkg/featureflag"
)

func (h *Handler) SomeHandler(c *gin.Context) {
    ctx := c.Request.Context()
    
    // Flagni tekshirish
    if featureflag.IsEnabled(ctx, "enable-new-feature", false) {
        // Yangi funksiya
        h.newFeatureLogic(ctx)
    } else {
        // Eski funksiya
        h.oldFeatureLogic(ctx)
    }
}
```

### 2. Foydalanuvchi Konteksti bilan

```go
func (h *Handler) UserHandler(c *gin.Context) {
    ctx := c.Request.Context()
    
    // Foydalanuvchi ma'lumotlarini olish
    userID := c.GetString("user_id")
    email := c.GetString("email")
    
    // Feature flag foydalanuvchisini yaratish
    user := featureflag.NewUser(userID).
        WithEmail(email).
        WithCustom("plan", "premium").
        WithCustom("beta", true)
    
    // Kontekstga qo'shish
    ctx = featureflag.WithUser(ctx, user)
    
    // Flagni tekshirish
    if featureflag.IsEnabled(ctx, "enable-premium-feature", false) {
        // Premium funksiya
    }
}
```

### 3. String Variation (A/B Testing)

```go
func (h *Handler) Homepage(c *gin.Context) {
    ctx := c.Request.Context()
    
    variant := featureflag.GetStringVariation(ctx, "homepage-variant", "variant-a")
    
    switch variant {
    case "variant-a":
        c.HTML(200, "homepage-a.html", nil)
    case "variant-b":
        c.HTML(200, "homepage-b.html", nil)
    default:
        c.HTML(200, "homepage-default.html", nil)
    }
}
```

### 4. Integer Variation (Rate Limiting)

```go
func (h *Handler) APIHandler(c *gin.Context) {
    ctx := c.Request.Context()
    
    rateLimit := featureflag.GetIntVariation(ctx, "api-rate-limit", 100)
    
    // Rate limit qo'llash
    if h.checkRateLimit(ctx, rateLimit) {
        // Request davom etadi
    } else {
        c.JSON(429, gin.H{"error": "Rate limit exceeded"})
        return
    }
}
```

### 5. JSON Variation (Murakkab Konfiguratsiya)

```go
type FeatureConfig struct {
    MaxItems    int  `json:"maxItems"`
    EnableCache bool `json:"enableCache"`
    Timeout     int  `json:"timeout"`
}

func (h *Handler) ConfigHandler(c *gin.Context) {
    ctx := c.Request.Context()
    
    defaultConfig := FeatureConfig{
        MaxItems:    10,
        EnableCache: true,
        Timeout:     30,
    }
    
    configData := featureflag.GetJSONVariation(ctx, "feature-config", defaultConfig)
    
    // Type assertion
    config, ok := configData.(map[string]any)
    if ok {
        // Konfiguratsiyadan foydalanish
    }
}
```

### 6. Middleware bilan Ishlatish

Agar feature flag clientni barcha handlerlar uchun mavjud qilmoqchi bo'lsangiz:

```go
// internal/controller/restapi/router.go
import (
    "gct/pkg/featureflag"
)

func NewRouter(handler *gin.Engine, cfg *config.Config, useCases *usecase.UseCase, l logger.Log) {
    // Feature flag clientni yaratish
    ffClient, err := featureflag.New(
        context.Background(),
        cfg.FeatureFlag,
        useCases.Redis, // Redis client
        l,
    )
    if err != nil {
        l.WithContext(context.Background()).Errorw("failed to init feature flags", zap.Error(err))
    }
    
    // Middleware qo'shish
    if ffClient != nil {
        handler.Use(featureflag.Middleware(ffClient))
    }
    
    // ... qolgan router konfiguratsiyasi
}
```

## Targeting Rules

### Foydalanuvchi Atributlari

```yaml
enable-feature:
  targeting:
    - name: "Email bo'yicha"
      query: email in ["user1@example.com", "user2@example.com"]
      variation: enabled
    
    - name: "Plan bo'yicha"
      query: plan eq "premium"
      variation: enabled
    
    - name: "Mamlakat bo'yicha"
      query: country eq "US"
      variation: enabled
```

### Foiz Rollout

```yaml
enable-feature:
  defaultRule:
    variation: disabled
    progressiveRollout:
      initial:
        variation: enabled
        percentage: 10  # 10% foydalanuvchilar
      end:
        variation: enabled
        percentage: 100  # 100% foydalanuvchilar
```

### Vaqt bo'yicha Scheduling

```yaml
enable-holiday-theme:
  scheduling:
    - start: 2026-12-01T00:00:00Z
      end: 2026-12-31T23:59:59Z
      variation: enabled
```

## Redis bilan Ishlatish

Redis orqali flaglarni dinamik yangilash:

```bash
# Redis'ga flag konfiguratsiyasini yuklash
redis-cli SET feature_flags "$(cat config/flags.yaml)"
```

`.env` faylida:

```bash
FEATURE_FLAG_USE_REDIS=true
FEATURE_FLAG_REDIS_KEY=feature_flags
```

## Best Practices

1. **Default Qiymatlar**: Har doim default qiymat bering, agar flag mavjud bo'lmasa
2. **Logging**: Flag evaluationlarini log qiling (debug level)
3. **Cleanup**: Eski flaglarni o'chiring, agar ular ishlatilmasa
4. **Testing**: Har bir flag variantini test qiling
5. **Documentation**: Har bir flagning maqsadini hujjatlang
6. **Monitoring**: Flag evaluationlarini monitoring qiling

## Misollar

### Use Case 1: Yangi API Endpoint

```go
func (h *Handler) NewEndpoint(c *gin.Context) {
    ctx := c.Request.Context()
    
    if !featureflag.IsEnabled(ctx, "enable-new-endpoint", false) {
        c.JSON(404, gin.H{"error": "Not found"})
        return
    }
    
    // Yangi endpoint logikasi
}
```

### Use Case 2: Database Migration

```go
func (r *Repository) GetUser(ctx context.Context, id string) (*User, error) {
    if featureflag.IsEnabled(ctx, "use-new-user-table", false) {
        return r.getUserFromNewTable(ctx, id)
    }
    return r.getUserFromOldTable(ctx, id)
}
```

### Use Case 3: External Service

```go
func (s *Service) SendEmail(ctx context.Context, email string) error {
    provider := featureflag.GetStringVariation(ctx, "email-provider", "sendgrid")
    
    switch provider {
    case "sendgrid":
        return s.sendgridClient.Send(ctx, email)
    case "ses":
        return s.sesClient.Send(ctx, email)
    default:
        return s.defaultClient.Send(ctx, email)
    }
}
```

## Troubleshooting

### Flag ishlamayapti

1. `FEATURE_FLAG_ENABLED=true` ekanligini tekshiring
2. Flag nomi to'g'ri yozilganligini tekshiring
3. `config/flags.yaml` fayli mavjudligini tekshiring
4. Loglarni tekshiring: `LOG_LEVEL=debug`

### Redis bilan muammolar

1. Redis ulanishi to'g'riligini tekshiring
2. Redis'da flag konfiguratsiyasi mavjudligini tekshiring: `redis-cli GET feature_flags`
3. `FEATURE_FLAG_USE_REDIS=true` ekanligini tekshiring

## Qo'shimcha Ma'lumot

- [GoFeatureFlag Documentation](https://gofeatureflag.org/)
- [Flag Configuration Format](https://gofeatureflag.org/docs/configure_flag/flag_format)
- [Targeting Rules](https://gofeatureflag.org/docs/configure_flag/rule_format)
