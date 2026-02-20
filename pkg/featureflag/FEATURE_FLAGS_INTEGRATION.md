# Feature Flags Integration Example

This example demonstrates how to integrate feature flags into your application.

## Step 1: Initialize Feature Flag Client

In your `internal/app/app.go`:

```go
package app

import (
    "context"
    "gct/config"
    "gct/pkg/featureflag"
    "gct/pkg/logger"
)

func Run(cfg *config.Config) {
    l := logger.New(cfg.Log.Level)
    ctx := context.Background()
    
    // ... other initializations (postgres, redis, etc.)
    
    // Initialize Feature Flags
    ffClient, err := featureflag.New(ctx, cfg.FeatureFlag, redisClient, l)
    if err != nil {
        l.Fatalw("failed to init feature flags", zap.Error(err))
    }
    defer ffClient.Close(ctx)
    
    // Pass ffClient to your router/controllers
    handler := initRouter(cfg, useCases, ffClient, l)
    
    // ... rest of the application
}
```

## Step 2: Add Middleware (Optional)

In your router setup:

```go
func initRouter(cfg *config.Config, useCases *usecase.UseCase, ffClient *featureflag.Client, l logger.Log) *gin.Engine {
    handler := gin.New()
    
    // Add feature flag middleware
    handler.Use(featureflag.Middleware(ffClient))
    
    // ... other middleware and routes
    
    return handler
}
```

## Step 3: Use Feature Flags in Controllers

### Example 1: Simple Boolean Flag

```go
func (h *Handler) GetUsers(c *gin.Context) {
    ctx := c.Request.Context()
    
    // Check if new pagination is enabled
    if featureflag.IsEnabled(ctx, "enable-new-pagination", false) {
        h.getUsersWithNewPagination(c)
    } else {
        h.getUsersWithOldPagination(c)
    }
}
```

### Example 2: User-Specific Targeting

```go
func (h *Handler) GetDashboard(c *gin.Context) {
    ctx := c.Request.Context()
    
    // Get user from session
    userID := c.GetString("user_id")
    email := c.GetString("email")
    plan := c.GetString("plan")
    
    // Create feature flag user
    user := featureflag.NewUser(userID).
        WithEmail(email).
        WithCustom("plan", plan)
    
    // Add to context
    ctx = featureflag.WithUser(ctx, user)
    
    // Check premium features
    if featureflag.IsEnabled(ctx, "enable-premium-dashboard", false) {
        c.JSON(200, h.getPremiumDashboard(ctx))
    } else {
        c.JSON(200, h.getStandardDashboard(ctx))
    }
}
```

### Example 3: A/B Testing with String Variation

```go
func (h *Handler) GetHomepage(c *gin.Context) {
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

### Example 4: Dynamic Configuration with JSON Variation

```go
func (h *Handler) ProcessRequest(c *gin.Context) {
    ctx := c.Request.Context()
    
    defaultConfig := map[string]any{
        "maxItems":    10,
        "enableCache": true,
        "timeout":     30,
    }
    
    config := featureflag.GetJSONVariation(ctx, "api-config", defaultConfig)
    
    maxItems := int(config["maxItems"].(float64))
    enableCache := config["enableCache"].(bool)
    timeout := int(config["timeout"].(float64))
    
    // Use configuration
    h.processWithConfig(ctx, maxItems, enableCache, timeout)
}
```

## Step 4: Use in Usecases

```go
package usecase

import (
    "context"
    "gct/pkg/featureflag"
)

type UserUseCase struct {
    repo UserRepo
}

func (uc *UserUseCase) CreateUser(ctx context.Context, req CreateUserRequest) error {
    // Check if email verification is enabled
    if featureflag.IsEnabled(ctx, "enable-email-verification", false) {
        // Send verification email
        if err := uc.sendVerificationEmail(ctx, req.Email); err != nil {
            return err
        }
    }
    
    // Create user
    return uc.repo.Create(ctx, req)
}
```

## Step 5: Use in Middleware

```go
func RateLimitMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx := c.Request.Context()
        
        // Get dynamic rate limit from feature flags
        rateLimit := featureflag.GetIntVariation(ctx, "api-rate-limit", 100)
        
        // Apply rate limiting
        if !checkRateLimit(ctx, rateLimit) {
            c.JSON(429, gin.H{"error": "Rate limit exceeded"})
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

## Configuration Examples

### Basic Boolean Flag

```yaml
enable-new-feature:
  variations:
    enabled: true
    disabled: false
  defaultRule:
    variation: disabled
```

### Percentage Rollout

```yaml
enable-new-ui:
  variations:
    enabled: true
    disabled: false
  defaultRule:
    variation: disabled
    progressiveRollout:
      initial:
        variation: enabled
        percentage: 10  # Start with 10%
      end:
        variation: enabled
        percentage: 100  # Eventually 100%
```

### User Targeting

```yaml
enable-premium-feature:
  variations:
    enabled: true
    disabled: false
  defaultRule:
    variation: disabled
  targeting:
    - name: "Premium users"
      query: plan eq "premium"
      variation: enabled
    - name: "Beta testers"
      query: beta eq true
      variation: enabled
```

### Environment-Based

```yaml
enable-debug-mode:
  variations:
    enabled: true
    disabled: false
  defaultRule:
    variation: disabled
  targeting:
    - name: "Development environment"
      query: environment eq "development"
      variation: enabled
```

## Testing

```go
func TestFeatureFlag(t *testing.T) {
    // Create test user
    user := featureflag.NewUser("test-user").
        WithCustom("plan", "premium")
    
    // Create context with user
    ctx := context.Background()
    ctx = featureflag.WithUser(ctx, user)
    
    // Your test logic here
}
```

## Best Practices

1. **Always provide default values** - Your app should work even if flags fail
2. **Use meaningful names** - `enable-payment-gateway-v2` not `flag1`
3. **Clean up old flags** - Remove flags after full rollout
4. **Log evaluations** - Use debug logging to track flag usage
5. **Test all variations** - Test both enabled and disabled states
6. **Document flags** - Keep `config/flags.yaml` well-documented
7. **Use targeting wisely** - Target specific users/groups for gradual rollouts
8. **Monitor flag usage** - Track which flags are being evaluated

## Troubleshooting

### Flag not working

1. Check `FEATURE_FLAG_ENABLED=true` in `.env`
2. Verify flag name spelling
3. Check `config/flags.yaml` exists and is valid
4. Enable debug logging: `LOG_LEVEL=debug`

### Redis issues

1. Verify Redis connection
2. Check flag config in Redis: `redis-cli GET feature_flags`
3. Ensure `FEATURE_FLAG_USE_REDIS=true`

## Additional Resources

- [Full Documentation](docs/FEATURE_FLAGS.md)
- [Package README](pkg/featureflag/README.md)
- [Example Controller](internal/controller/restapi/v1/featureflag/featureflag_controller.go)
- [GoFeatureFlag Docs](https://gofeatureflag.org/)
