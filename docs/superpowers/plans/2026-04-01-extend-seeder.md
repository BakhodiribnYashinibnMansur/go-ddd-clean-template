# Extend Seeder Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Extend the existing seeder to populate dummy data for 12 additional bounded contexts (announcements, notifications, feature flags, integrations, translations, file metadata, site settings, error codes, IP rules, rate limits, audit logs, function metrics).

**Architecture:** Each BC gets its own `seed_xxx.go` file inside `internal/seeder/`. The main `seeder.go` orchestrates them in dependency order. Config and Asynq payload are extended with new count fields. All inserts use raw SQL via `pool.Exec()` with `gofakeit` for fake data — matching the existing pattern in `authz.go` and `users.go`.

**Tech Stack:** Go, pgx/v5, gofakeit/v7, Asynq

---

### Task 1: Extend config and Asynq payload

**Files:**
- Modify: `config/seeder.go`
- Modify: `internal/shared/infrastructure/asynq/handlers.go`

- [ ] **Step 1: Add new count fields to config/seeder.go**

```go
// config/seeder.go — add these fields to the Seeder struct after PoliciesCount:

	// Number of fake announcements to create
	AnnouncementsCount int `env:"SEEDER_ANNOUNCEMENTS_COUNT" envDefault:"10"`
	// Number of fake notifications to create
	NotificationsCount int `env:"SEEDER_NOTIFICATIONS_COUNT" envDefault:"30"`
	// Number of fake feature flags to create
	FeatureFlagsCount int `env:"SEEDER_FEATURE_FLAGS_COUNT" envDefault:"15"`
	// Number of fake integrations to create
	IntegrationsCount int `env:"SEEDER_INTEGRATIONS_COUNT" envDefault:"5"`
	// Number of fake translations to create
	TranslationsCount int `env:"SEEDER_TRANSLATIONS_COUNT" envDefault:"50"`
	// Number of fake file metadata records to create
	FileMetadataCount int `env:"SEEDER_FILE_METADATA_COUNT" envDefault:"20"`
	// Number of fake site settings to create
	SiteSettingsCount int `env:"SEEDER_SITE_SETTINGS_COUNT" envDefault:"15"`
	// Number of fake error codes to create
	ErrorCodesCount int `env:"SEEDER_ERROR_CODES_COUNT" envDefault:"20"`
	// Number of fake IP rules to create
	IPRulesCount int `env:"SEEDER_IP_RULES_COUNT" envDefault:"10"`
	// Number of fake rate limits to create
	RateLimitsCount int `env:"SEEDER_RATE_LIMITS_COUNT" envDefault:"8"`
	// Number of fake audit logs to create
	AuditLogsCount int `env:"SEEDER_AUDIT_LOGS_COUNT" envDefault:"50"`
	// Number of fake function metrics to create
	FunctionMetricsCount int `env:"SEEDER_FUNCTION_METRICS_COUNT" envDefault:"30"`
```

- [ ] **Step 2: Add new count fields to SeedPayload in handlers.go**

Add these fields to the `SeedPayload` struct in `internal/shared/infrastructure/asynq/handlers.go`:

```go
	AnnouncementsCount   int `json:"announcements_count"`
	NotificationsCount   int `json:"notifications_count"`
	FeatureFlagsCount    int `json:"feature_flags_count"`
	IntegrationsCount    int `json:"integrations_count"`
	TranslationsCount    int `json:"translations_count"`
	FileMetadataCount    int `json:"file_metadata_count"`
	SiteSettingsCount    int `json:"site_settings_count"`
	ErrorCodesCount      int `json:"error_codes_count"`
	IPRulesCount         int `json:"ip_rules_count"`
	RateLimitsCount      int `json:"rate_limits_count"`
	AuditLogsCount       int `json:"audit_logs_count"`
	FunctionMetricsCount int `json:"function_metrics_count"`
```

- [ ] **Step 3: Update init_asynq.go to pass new counts**

In `internal/app/init_asynq.go`, add these lines inside the seed handler after the existing `customCounts` assignments (after the `PoliciesCount` block, before `customCounts["clear_data"]`):

```go
		if payload.AnnouncementsCount > 0 {
			customCounts["announcements"] = payload.AnnouncementsCount
		}
		if payload.NotificationsCount > 0 {
			customCounts["notifications"] = payload.NotificationsCount
		}
		if payload.FeatureFlagsCount > 0 {
			customCounts["feature_flags"] = payload.FeatureFlagsCount
		}
		if payload.IntegrationsCount > 0 {
			customCounts["integrations"] = payload.IntegrationsCount
		}
		if payload.TranslationsCount > 0 {
			customCounts["translations"] = payload.TranslationsCount
		}
		if payload.FileMetadataCount > 0 {
			customCounts["file_metadata"] = payload.FileMetadataCount
		}
		if payload.SiteSettingsCount > 0 {
			customCounts["site_settings"] = payload.SiteSettingsCount
		}
		if payload.ErrorCodesCount > 0 {
			customCounts["error_codes"] = payload.ErrorCodesCount
		}
		if payload.IPRulesCount > 0 {
			customCounts["ip_rules"] = payload.IPRulesCount
		}
		if payload.RateLimitsCount > 0 {
			customCounts["rate_limits"] = payload.RateLimitsCount
		}
		if payload.AuditLogsCount > 0 {
			customCounts["audit_logs"] = payload.AuditLogsCount
		}
		if payload.FunctionMetricsCount > 0 {
			customCounts["function_metrics"] = payload.FunctionMetricsCount
		}
```

- [ ] **Step 4: Update cmd/seeder/main.go payload**

Add the new fields to the payload struct in `cmd/seeder/main.go`:

```go
	payload := asynq.SeedPayload{
		UsersCount:           100,
		RolesCount:           10,
		PermissionsCount:     20,
		PoliciesCount:        20,
		AnnouncementsCount:   10,
		NotificationsCount:   30,
		FeatureFlagsCount:    15,
		IntegrationsCount:    5,
		TranslationsCount:    50,
		FileMetadataCount:    20,
		SiteSettingsCount:    15,
		ErrorCodesCount:      20,
		IPRulesCount:         10,
		RateLimitsCount:      8,
		AuditLogsCount:       50,
		FunctionMetricsCount: 30,
		Seed:                 0,
		ClearData:            true,
	}
```

- [ ] **Step 5: Verify compilation**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go build ./...`
Expected: builds with no errors

- [ ] **Step 6: Commit**

```bash
git add config/seeder.go internal/shared/infrastructure/asynq/handlers.go internal/app/init_asynq.go cmd/seeder/main.go
git commit -m "feat(seeder): extend config and payload for all BCs"
```

---

### Task 2: Update seeder.go orchestration (clearData + Seed method)

**Files:**
- Modify: `internal/seeder/seeder.go`

- [ ] **Step 1: Add new tables to clearData**

In `internal/seeder/seeder.go`, replace the `tables` slice in `clearData()` with:

```go
	tables := []string{
		// Reverse FK order — dependent tables first
		"function_metrics",
		"endpoint_history",
		"audit_log",
		"notifications",
		"announcements",
		"translations",
		"file_metadata",
		"api_keys",
		"integrations",
		"ip_rules",
		"rate_limits",
		"feature_flags",
		"error_code",
		"site_settings",
		// Original tables
		"user_relation",
		"role_permission",
		"session",
		"policy",
		"users",
		"role",
		"permission",
	}
```

- [ ] **Step 2: Add new seed calls to Seed method**

In the `Seed()` method, after the existing `seedPolicies` call and before the `duration` calculation, add:

```go
	// Phase 1: Independent tables
	siteSettingsCount := s.getCount(customCounts, "site_settings", s.cfg.Seeder.SiteSettingsCount)
	if err := s.seedSiteSettings(ctx, siteSettingsCount); err != nil {
		return fmt.Errorf("failed to seed site settings: %w", err)
	}

	errorCodesCount := s.getCount(customCounts, "error_codes", s.cfg.Seeder.ErrorCodesCount)
	if err := s.seedErrorCodes(ctx, errorCodesCount); err != nil {
		return fmt.Errorf("failed to seed error codes: %w", err)
	}

	featureFlagsCount := s.getCount(customCounts, "feature_flags", s.cfg.Seeder.FeatureFlagsCount)
	if err := s.seedFeatureFlags(ctx, featureFlagsCount); err != nil {
		return fmt.Errorf("failed to seed feature flags: %w", err)
	}

	rateLimitsCount := s.getCount(customCounts, "rate_limits", s.cfg.Seeder.RateLimitsCount)
	if err := s.seedRateLimits(ctx, rateLimitsCount); err != nil {
		return fmt.Errorf("failed to seed rate limits: %w", err)
	}

	ipRulesCount := s.getCount(customCounts, "ip_rules", s.cfg.Seeder.IPRulesCount)
	if err := s.seedIPRules(ctx, ipRulesCount); err != nil {
		return fmt.Errorf("failed to seed ip rules: %w", err)
	}

	functionMetricsCount := s.getCount(customCounts, "function_metrics", s.cfg.Seeder.FunctionMetricsCount)
	if err := s.seedFunctionMetrics(ctx, functionMetricsCount); err != nil {
		return fmt.Errorf("failed to seed function metrics: %w", err)
	}

	// Phase 2: Depends on users
	integrationsCount := s.getCount(customCounts, "integrations", s.cfg.Seeder.IntegrationsCount)
	if err := s.seedIntegrations(ctx, integrationsCount); err != nil {
		return fmt.Errorf("failed to seed integrations: %w", err)
	}

	fileMetadataCount := s.getCount(customCounts, "file_metadata", s.cfg.Seeder.FileMetadataCount)
	if err := s.seedFileMetadata(ctx, fileMetadataCount); err != nil {
		return fmt.Errorf("failed to seed file metadata: %w", err)
	}

	translationsCount := s.getCount(customCounts, "translations", s.cfg.Seeder.TranslationsCount)
	if err := s.seedTranslations(ctx, translationsCount); err != nil {
		return fmt.Errorf("failed to seed translations: %w", err)
	}

	announcementsCount := s.getCount(customCounts, "announcements", s.cfg.Seeder.AnnouncementsCount)
	if err := s.seedAnnouncements(ctx, announcementsCount); err != nil {
		return fmt.Errorf("failed to seed announcements: %w", err)
	}

	notificationsCount := s.getCount(customCounts, "notifications", s.cfg.Seeder.NotificationsCount)
	if err := s.seedNotifications(ctx, notificationsCount); err != nil {
		return fmt.Errorf("failed to seed notifications: %w", err)
	}

	// Phase 3: Depends on users + sessions + policies
	auditLogsCount := s.getCount(customCounts, "audit_logs", s.cfg.Seeder.AuditLogsCount)
	if err := s.seedAuditLogs(ctx, auditLogsCount); err != nil {
		return fmt.Errorf("failed to seed audit logs: %w", err)
	}
```

- [ ] **Step 3: Verify compilation (will fail — seed methods not yet defined)**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go build ./...`
Expected: compilation errors for undefined seed methods — this is correct, we'll implement them in Tasks 3-14.

- [ ] **Step 4: Commit**

```bash
git add internal/seeder/seeder.go
git commit -m "feat(seeder): add orchestration for all BC seed methods"
```

---

### Task 3: Seed site settings

**Files:**
- Create: `internal/seeder/seed_site_settings.go`

- [ ] **Step 1: Create seed_site_settings.go**

```go
package seeder

import (
	"context"
	"fmt"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *Seeder) seedSiteSettings(ctx context.Context, count int) error {
	s.logger.Infoc(ctx, "Seeding site settings...", zap.Int("count", count))

	now := time.Now()

	predefined := []struct {
		key       string
		value     string
		valueType string
		category  string
		desc      string
		isPublic  bool
	}{
		{"site_name", "My Application", "string", "general", "The name of the site", true},
		{"site_description", "A modern web application", "string", "general", "Site description for SEO", true},
		{"maintenance_mode", "false", "bool", "general", "Enable maintenance mode", false},
		{"max_upload_size", "10485760", "int", "general", "Max file upload size in bytes", false},
		{"smtp_host", "smtp.example.com", "string", "email", "SMTP server hostname", false},
		{"smtp_port", "587", "int", "email", "SMTP server port", false},
		{"smtp_from", "noreply@example.com", "string", "email", "Default sender email", false},
		{"session_timeout", "3600", "int", "security", "Session timeout in seconds", false},
		{"rate_limit_enabled", "true", "bool", "security", "Enable global rate limiting", false},
		{"api_version", "v1", "string", "api", "Current API version", true},
	}

	for _, s2 := range predefined {
		_, err := s.pool.Exec(ctx,
			`INSERT INTO site_settings (id, key, value, value_type, category, description, is_public, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			uuid.New(), s2.key, s2.value, s2.valueType, s2.category, s2.desc, s2.isPublic, now, now,
		)
		if err != nil {
			s.logger.Warnc(ctx, "Failed to create predefined site setting", zap.Error(err), zap.String("key", s2.key))
		}
	}

	categories := []string{"general", "email", "security", "api"}
	valueTypes := []string{"string", "bool", "int"}

	for i := 0; i < count-len(predefined); i++ {
		if i+len(predefined) >= count {
			break
		}
		key := fmt.Sprintf("custom_%s_%d", gofakeit.Word(), i)
		_, err := s.pool.Exec(ctx,
			`INSERT INTO site_settings (id, key, value, value_type, category, description, is_public, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			uuid.New(), key, gofakeit.Word(), valueTypes[gofakeit.Number(0, len(valueTypes)-1)],
			categories[gofakeit.Number(0, len(categories)-1)], gofakeit.Sentence(5), gofakeit.Bool(), now, now,
		)
		if err != nil {
			s.logger.Warnc(ctx, "Failed to create random site setting", zap.Error(err), zap.String("key", key))
		}
	}

	return nil
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/seeder/seed_site_settings.go
git commit -m "feat(seeder): add site settings seed data"
```

---

### Task 4: Seed error codes

**Files:**
- Create: `internal/seeder/seed_error_codes.go`

- [ ] **Step 1: Create seed_error_codes.go**

```go
package seeder

import (
	"context"
	"fmt"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *Seeder) seedErrorCodes(ctx context.Context, count int) error {
	s.logger.Infoc(ctx, "Seeding error codes...", zap.Int("count", count))

	now := time.Now()

	predefined := []struct {
		code       string
		message    string
		httpStatus int
		category   string
		severity   string
		retryable  bool
		retryAfter int
		suggestion string
	}{
		{"AUTH_001", "Invalid credentials", 401, "AUTH", "MEDIUM", false, 0, "Check username and password"},
		{"AUTH_002", "Token expired", 401, "AUTH", "LOW", true, 0, "Refresh your access token"},
		{"AUTH_003", "Insufficient permissions", 403, "AUTH", "MEDIUM", false, 0, "Contact administrator for access"},
		{"DATA_001", "Resource not found", 404, "DATA", "LOW", false, 0, "Verify the resource ID"},
		{"DATA_002", "Duplicate entry", 409, "DATA", "LOW", false, 0, "Use a unique value"},
		{"DATA_003", "Invalid input format", 400, "VALIDATION", "LOW", false, 0, "Check the request body format"},
		{"SYS_001", "Internal server error", 500, "SYSTEM", "HIGH", true, 30, "Try again later"},
		{"SYS_002", "Service unavailable", 503, "SYSTEM", "CRITICAL", true, 60, "Service is under maintenance"},
		{"SYS_003", "Database connection failed", 500, "SYSTEM", "CRITICAL", true, 10, "Try again shortly"},
		{"BIZ_001", "Rate limit exceeded", 429, "BUSINESS", "LOW", true, 60, "Wait before retrying"},
		{"BIZ_002", "Account suspended", 403, "BUSINESS", "HIGH", false, 0, "Contact support"},
		{"VAL_001", "Required field missing", 400, "VALIDATION", "LOW", false, 0, "Provide all required fields"},
	}

	for _, ec := range predefined {
		_, err := s.pool.Exec(ctx,
			`INSERT INTO error_code (id, code, message, http_status, category, severity, retryable, retry_after, suggestion, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
			uuid.New(), ec.code, ec.message, ec.httpStatus, ec.category, ec.severity, ec.retryable, ec.retryAfter, ec.suggestion, now, now,
		)
		if err != nil {
			s.logger.Warnc(ctx, "Failed to create predefined error code", zap.Error(err), zap.String("code", ec.code))
		}
	}

	categories := []string{"DATA", "AUTH", "SYSTEM", "VALIDATION", "BUSINESS"}
	severities := []string{"LOW", "MEDIUM", "HIGH", "CRITICAL"}
	statuses := []int{400, 401, 403, 404, 409, 422, 429, 500, 502, 503}

	for i := 0; i < count-len(predefined); i++ {
		if i+len(predefined) >= count {
			break
		}
		cat := categories[gofakeit.Number(0, len(categories)-1)]
		code := fmt.Sprintf("%s_%03d", cat, i+100)
		_, err := s.pool.Exec(ctx,
			`INSERT INTO error_code (id, code, message, http_status, category, severity, retryable, retry_after, suggestion, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
			uuid.New(), code, gofakeit.Sentence(4), statuses[gofakeit.Number(0, len(statuses)-1)],
			cat, severities[gofakeit.Number(0, len(severities)-1)], gofakeit.Bool(), gofakeit.Number(0, 120),
			gofakeit.Sentence(6), now, now,
		)
		if err != nil {
			s.logger.Warnc(ctx, "Failed to create random error code", zap.Error(err), zap.String("code", code))
		}
	}

	return nil
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/seeder/seed_error_codes.go
git commit -m "feat(seeder): add error codes seed data"
```

---

### Task 5: Seed feature flags

**Files:**
- Create: `internal/seeder/seed_feature_flags.go`

- [ ] **Step 1: Create seed_feature_flags.go**

```go
package seeder

import (
	"context"
	"fmt"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *Seeder) seedFeatureFlags(ctx context.Context, count int) error {
	s.logger.Infoc(ctx, "Seeding feature flags...", zap.Int("count", count))

	now := time.Now()

	predefined := []struct {
		key   string
		name  string
		fType string
		value string
		desc  string
	}{
		{"dark_mode", "Dark Mode", "bool", "false", "Enable dark mode for the UI"},
		{"new_dashboard", "New Dashboard", "bool", "true", "Show the redesigned dashboard"},
		{"beta_api_v2", "Beta API v2", "bool", "false", "Enable API v2 beta endpoints"},
		{"max_upload_mb", "Max Upload Size (MB)", "int", "50", "Maximum file upload size in megabytes"},
		{"welcome_message", "Welcome Message", "string", "Welcome to our platform!", "Landing page welcome message"},
		{"maintenance_banner", "Maintenance Banner", "string", "", "Banner text during maintenance"},
		{"signup_enabled", "User Signup", "bool", "true", "Allow new user registrations"},
		{"export_formats", "Export Formats", "json", `["csv","xlsx","pdf"]`, "Available data export formats"},
	}

	for _, ff := range predefined {
		_, err := s.pool.Exec(ctx,
			`INSERT INTO feature_flags (id, key, name, type, value, description, is_active, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			uuid.New(), ff.key, ff.name, ff.fType, ff.value, ff.desc, true, now, now,
		)
		if err != nil {
			s.logger.Warnc(ctx, "Failed to create predefined feature flag", zap.Error(err), zap.String("key", ff.key))
		}
	}

	flagTypes := []string{"bool", "string", "int"}

	for i := 0; i < count-len(predefined); i++ {
		if i+len(predefined) >= count {
			break
		}
		key := fmt.Sprintf("%s_%s_%d", gofakeit.Word(), gofakeit.Word(), i)
		fType := flagTypes[gofakeit.Number(0, len(flagTypes)-1)]
		var value string
		switch fType {
		case "bool":
			value = fmt.Sprintf("%t", gofakeit.Bool())
		case "int":
			value = fmt.Sprintf("%d", gofakeit.Number(1, 1000))
		case "string":
			value = gofakeit.Word()
		}
		_, err := s.pool.Exec(ctx,
			`INSERT INTO feature_flags (id, key, name, type, value, description, is_active, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			uuid.New(), key, gofakeit.Sentence(3), fType, value, gofakeit.Sentence(5), gofakeit.Bool(), now, now,
		)
		if err != nil {
			s.logger.Warnc(ctx, "Failed to create random feature flag", zap.Error(err), zap.String("key", key))
		}
	}

	return nil
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/seeder/seed_feature_flags.go
git commit -m "feat(seeder): add feature flags seed data"
```

---

### Task 6: Seed rate limits

**Files:**
- Create: `internal/seeder/seed_rate_limits.go`

- [ ] **Step 1: Create seed_rate_limits.go**

```go
package seeder

import (
	"context"
	"fmt"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *Seeder) seedRateLimits(ctx context.Context, count int) error {
	s.logger.Infoc(ctx, "Seeding rate limits...", zap.Int("count", count))

	now := time.Now()

	predefined := []struct {
		name          string
		pathPattern   string
		method        string
		limitCount    int
		windowSeconds int
	}{
		{"Login Rate Limit", "/auth/login", "POST", 5, 60},
		{"Register Rate Limit", "/auth/register", "POST", 3, 300},
		{"API General Limit", "/api/v1/*", "ALL", 100, 60},
		{"File Upload Limit", "/api/v1/files/upload", "POST", 10, 300},
		{"Password Reset Limit", "/auth/reset-password", "POST", 3, 600},
		{"Export Rate Limit", "/api/v1/export/*", "POST", 5, 300},
	}

	for _, rl := range predefined {
		_, err := s.pool.Exec(ctx,
			`INSERT INTO rate_limits (id, name, path_pattern, method, limit_count, window_seconds, is_active, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			uuid.New(), rl.name, rl.pathPattern, rl.method, rl.limitCount, rl.windowSeconds, true, now, now,
		)
		if err != nil {
			s.logger.Warnc(ctx, "Failed to create predefined rate limit", zap.Error(err), zap.String("name", rl.name))
		}
	}

	methods := []string{"GET", "POST", "PUT", "DELETE", "ALL"}

	for i := 0; i < count-len(predefined); i++ {
		if i+len(predefined) >= count {
			break
		}
		name := fmt.Sprintf("%s Limit %d", gofakeit.Word(), i)
		path := fmt.Sprintf("/api/v1/%s/*", gofakeit.Word())
		_, err := s.pool.Exec(ctx,
			`INSERT INTO rate_limits (id, name, path_pattern, method, limit_count, window_seconds, is_active, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			uuid.New(), name, path, methods[gofakeit.Number(0, len(methods)-1)],
			gofakeit.Number(10, 200), gofakeit.Number(30, 600), gofakeit.Bool(), now, now,
		)
		if err != nil {
			s.logger.Warnc(ctx, "Failed to create random rate limit", zap.Error(err), zap.String("name", name))
		}
	}

	return nil
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/seeder/seed_rate_limits.go
git commit -m "feat(seeder): add rate limits seed data"
```

---

### Task 7: Seed IP rules

**Files:**
- Create: `internal/seeder/seed_ip_rules.go`

- [ ] **Step 1: Create seed_ip_rules.go**

```go
package seeder

import (
	"context"
	"fmt"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *Seeder) seedIPRules(ctx context.Context, count int) error {
	s.logger.Infoc(ctx, "Seeding IP rules...", zap.Int("count", count))

	now := time.Now()

	predefined := []struct {
		ip     string
		action string
		reason string
	}{
		{"127.0.0.1", "allow", "Localhost"},
		{"10.0.0.0/8", "allow", "Internal network"},
		{"192.168.0.0/16", "allow", "Private network"},
		{"203.0.113.50", "block", "Known malicious IP"},
		{"198.51.100.0/24", "block", "Suspicious subnet"},
	}

	for _, rule := range predefined {
		_, err := s.pool.Exec(ctx,
			`INSERT INTO ip_rules (id, ip_address, type, reason, is_active, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			uuid.New(), rule.ip, rule.action, rule.reason, true, now, now,
		)
		if err != nil {
			s.logger.Warnc(ctx, "Failed to create predefined IP rule", zap.Error(err), zap.String("ip", rule.ip))
		}
	}

	actions := []string{"allow", "block"}
	reasons := []string{"Automated scan detected", "Brute force attempt", "VPN access", "Office network", "Partner API"}

	for i := 0; i < count-len(predefined); i++ {
		if i+len(predefined) >= count {
			break
		}
		ip := fmt.Sprintf("%d.%d.%d.%d", gofakeit.Number(1, 223), gofakeit.Number(0, 255), gofakeit.Number(0, 255), gofakeit.Number(1, 254))
		_, err := s.pool.Exec(ctx,
			`INSERT INTO ip_rules (id, ip_address, type, reason, is_active, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			uuid.New(), ip, actions[gofakeit.Number(0, 1)], reasons[gofakeit.Number(0, len(reasons)-1)], gofakeit.Bool(), now, now,
		)
		if err != nil {
			s.logger.Warnc(ctx, "Failed to create random IP rule", zap.Error(err), zap.String("ip", ip))
		}
	}

	return nil
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/seeder/seed_ip_rules.go
git commit -m "feat(seeder): add IP rules seed data"
```

---

### Task 8: Seed function metrics

**Files:**
- Create: `internal/seeder/seed_function_metrics.go`

- [ ] **Step 1: Create seed_function_metrics.go**

```go
package seeder

import (
	"context"
	"fmt"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *Seeder) seedFunctionMetrics(ctx context.Context, count int) error {
	s.logger.Infoc(ctx, "Seeding function metrics...", zap.Int("count", count))

	functionNames := []string{
		"UserService.Create", "UserService.GetByID", "UserService.Update",
		"AuthService.Login", "AuthService.Logout", "AuthService.RefreshToken",
		"RoleService.AssignRole", "PolicyService.Evaluate",
		"FileService.Upload", "FileService.Download",
		"NotificationService.Send", "AuditService.Log",
		"ExportService.GenerateCSV", "DashboardService.GetStats",
		"IntegrationService.Sync", "TranslationService.GetAll",
	}

	for i := 0; i < count; i++ {
		name := functionNames[gofakeit.Number(0, len(functionNames)-1)]
		latency := gofakeit.Number(5, 500)
		isPanic := gofakeit.Float64Range(0, 1) < 0.05 // 5% panic rate
		var panicError *string
		if isPanic {
			err := fmt.Sprintf("panic: %s at %s:%d", gofakeit.ErrorRuntime().Error(), gofakeit.Word()+".go", gofakeit.Number(10, 500))
			panicError = &err
		}
		createdAt := gofakeit.DateRange(time.Now().AddDate(0, -1, 0), time.Now())

		_, err := s.pool.Exec(ctx,
			`INSERT INTO function_metrics (id, name, latency_ms, is_panic, panic_error, created_at)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			uuid.New(), name, latency, isPanic, panicError, createdAt,
		)
		if err != nil {
			s.logger.Warnc(ctx, "Failed to create function metric", zap.Error(err), zap.String("name", name))
		}
	}

	return nil
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/seeder/seed_function_metrics.go
git commit -m "feat(seeder): add function metrics seed data"
```

---

### Task 9: Seed integrations + API keys

**Files:**
- Create: `internal/seeder/seed_integrations.go`

- [ ] **Step 1: Create seed_integrations.go**

```go
package seeder

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *Seeder) seedIntegrations(ctx context.Context, count int) error {
	s.logger.Infoc(ctx, "Seeding integrations...", zap.Int("count", count))

	now := time.Now()

	predefined := []struct {
		name    string
		desc    string
		baseURL string
	}{
		{"Slack", "Slack notifications integration", "https://hooks.slack.com/services"},
		{"Telegram", "Telegram bot integration", "https://api.telegram.org"},
		{"Email SMTP", "Email notification service", "https://smtp.example.com"},
		{"Webhook Relay", "Generic webhook relay service", "https://webhook.example.com"},
		{"Monitoring API", "External monitoring integration", "https://monitor.example.com/api"},
	}

	for i, intg := range predefined {
		if i >= count {
			break
		}
		integrationID := uuid.New()
		_, err := s.pool.Exec(ctx,
			`INSERT INTO integrations (id, name, description, base_url, is_active, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			integrationID, intg.name, intg.desc, intg.baseURL, true, now, now,
		)
		if err != nil {
			s.logger.Warnc(ctx, "Failed to create integration", zap.Error(err), zap.String("name", intg.name))
			continue
		}

		// Create an API key for each integration
		apiKey := generateAPIKey()
		prefix := apiKey[:8]
		_, err = s.pool.Exec(ctx,
			`INSERT INTO api_keys (id, integration_id, name, key, key_prefix, is_active, expires_at, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			uuid.New(), integrationID, fmt.Sprintf("%s API Key", intg.name), apiKey, prefix, true, now.AddDate(1, 0, 0), now, now,
		)
		if err != nil {
			s.logger.Warnc(ctx, "Failed to create API key", zap.Error(err), zap.String("integration", intg.name))
		}
	}

	for i := 0; i < count-len(predefined); i++ {
		if i+len(predefined) >= count {
			break
		}
		integrationID := uuid.New()
		name := fmt.Sprintf("%s Integration %d", gofakeit.Company(), i)
		_, err := s.pool.Exec(ctx,
			`INSERT INTO integrations (id, name, description, base_url, is_active, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			integrationID, name, gofakeit.Sentence(5), fmt.Sprintf("https://%s.example.com/api", gofakeit.Word()),
			gofakeit.Bool(), now, now,
		)
		if err != nil {
			s.logger.Warnc(ctx, "Failed to create random integration", zap.Error(err), zap.String("name", name))
			continue
		}

		apiKey := generateAPIKey()
		_, _ = s.pool.Exec(ctx,
			`INSERT INTO api_keys (id, integration_id, name, key, key_prefix, is_active, expires_at, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			uuid.New(), integrationID, fmt.Sprintf("%s Key", name), apiKey, apiKey[:8], true, now.AddDate(1, 0, 0), now, now,
		)
	}

	return nil
}

func generateAPIKey() string {
	bytes := make([]byte, 32)
	_, _ = rand.Read(bytes)
	return "gct_" + hex.EncodeToString(bytes)
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/seeder/seed_integrations.go
git commit -m "feat(seeder): add integrations and API keys seed data"
```

---

### Task 10: Seed file metadata

**Files:**
- Create: `internal/seeder/seed_file_metadata.go`

- [ ] **Step 1: Create seed_file_metadata.go**

```go
package seeder

import (
	"context"
	"fmt"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *Seeder) seedFileMetadata(ctx context.Context, count int) error {
	s.logger.Infoc(ctx, "Seeding file metadata...", zap.Int("count", count))

	// Get user IDs for uploaded_by reference
	rows, err := s.pool.Query(ctx, "SELECT id FROM users LIMIT 50")
	if err != nil {
		return fmt.Errorf("failed to get users for file metadata: %w", err)
	}
	defer rows.Close()

	var userIDs []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return fmt.Errorf("failed to scan user id: %w", err)
		}
		userIDs = append(userIDs, id)
	}

	now := time.Now()

	fileTypes := []struct {
		ext      string
		mimeType string
		minSize  int64
		maxSize  int64
	}{
		{"pdf", "application/pdf", 50000, 10000000},
		{"png", "image/png", 10000, 5000000},
		{"jpg", "image/jpeg", 20000, 8000000},
		{"xlsx", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", 30000, 15000000},
		{"csv", "text/csv", 1000, 50000000},
		{"docx", "application/vnd.openxmlformats-officedocument.wordprocessingml.document", 20000, 10000000},
		{"json", "application/json", 100, 1000000},
		{"txt", "text/plain", 100, 500000},
	}

	for i := 0; i < count; i++ {
		ft := fileTypes[gofakeit.Number(0, len(fileTypes)-1)]
		originalName := fmt.Sprintf("%s_%s.%s", gofakeit.Word(), gofakeit.Word(), ft.ext)
		storedName := fmt.Sprintf("%s.%s", uuid.New().String(), ft.ext)
		size := gofakeit.Int64Range(ft.minSize, ft.maxSize)
		url := fmt.Sprintf("/files/%s/%s", "uploads", storedName)

		var uploadedBy *uuid.UUID
		if len(userIDs) > 0 {
			u := userIDs[gofakeit.Number(0, len(userIDs)-1)]
			uploadedBy = &u
		}

		_, err := s.pool.Exec(ctx,
			`INSERT INTO file_metadata (id, original_name, stored_name, bucket, url, size, mime_type, uploaded_by, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
			uuid.New(), originalName, storedName, "uploads", url, size, ft.mimeType, uploadedBy, now, now,
		)
		if err != nil {
			s.logger.Warnc(ctx, "Failed to create file metadata", zap.Error(err), zap.String("name", originalName))
		}
	}

	return nil
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/seeder/seed_file_metadata.go
git commit -m "feat(seeder): add file metadata seed data"
```

---

### Task 11: Seed translations

**Files:**
- Create: `internal/seeder/seed_translations.go`

- [ ] **Step 1: Create seed_translations.go**

```go
package seeder

import (
	"context"
	"fmt"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *Seeder) seedTranslations(ctx context.Context, count int) error {
	s.logger.Infoc(ctx, "Seeding translations...", zap.Int("count", count))

	now := time.Now()

	type translationSet struct {
		entityType string
		key        string
		uz         string
		ru         string
		en         string
	}

	predefined := []translationSet{
		{"ui", "auth.login", "Kirish", "Вход", "Login"},
		{"ui", "auth.logout", "Chiqish", "Выход", "Logout"},
		{"ui", "auth.register", "Ro'yxatdan o'tish", "Регистрация", "Register"},
		{"ui", "error.not_found", "Topilmadi", "Не найдено", "Not Found"},
		{"ui", "error.forbidden", "Ruxsat berilmagan", "Доступ запрещён", "Forbidden"},
		{"ui", "error.server_error", "Server xatosi", "Ошибка сервера", "Server Error"},
		{"ui", "nav.dashboard", "Boshqaruv paneli", "Панель управления", "Dashboard"},
		{"ui", "nav.settings", "Sozlamalar", "Настройки", "Settings"},
		{"ui", "nav.users", "Foydalanuvchilar", "Пользователи", "Users"},
		{"ui", "action.save", "Saqlash", "Сохранить", "Save"},
		{"ui", "action.cancel", "Bekor qilish", "Отмена", "Cancel"},
		{"ui", "action.delete", "O'chirish", "Удалить", "Delete"},
		{"ui", "action.search", "Qidirish", "Поиск", "Search"},
		{"ui", "message.success", "Muvaffaqiyatli", "Успешно", "Success"},
		{"ui", "message.confirm_delete", "O'chirishni tasdiqlaysizmi?", "Подтвердите удаление?", "Confirm deletion?"},
	}

	langs := []string{"uz", "ru", "en"}

	for _, t := range predefined {
		entityID := uuid.New()
		translations := map[string]string{"uz": t.uz, "ru": t.ru, "en": t.en}
		for _, lang := range langs {
			data := fmt.Sprintf(`{"%s":"%s"}`, t.key, translations[lang])
			_, err := s.pool.Exec(ctx,
				`INSERT INTO translations (id, entity_type, entity_id, lang_code, data, created_at, updated_at)
				 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
				uuid.New(), t.entityType, entityID, lang, data, now, now,
			)
			if err != nil {
				s.logger.Warnc(ctx, "Failed to create predefined translation", zap.Error(err), zap.String("key", t.key))
			}
		}
	}

	entityTypes := []string{"ui", "email", "notification", "error"}

	remaining := count - len(predefined)*len(langs)
	for i := 0; i < remaining; i++ {
		if i < 0 {
			break
		}
		entityID := uuid.New()
		entityType := entityTypes[gofakeit.Number(0, len(entityTypes)-1)]
		lang := langs[gofakeit.Number(0, len(langs)-1)]
		key := fmt.Sprintf("%s.%s_%d", entityType, gofakeit.Word(), i)
		data := fmt.Sprintf(`{"%s":"%s"}`, key, gofakeit.Sentence(3))

		_, err := s.pool.Exec(ctx,
			`INSERT INTO translations (id, entity_type, entity_id, lang_code, data, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			uuid.New(), entityType, entityID, lang, data, now, now,
		)
		if err != nil {
			s.logger.Warnc(ctx, "Failed to create random translation", zap.Error(err))
		}
	}

	return nil
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/seeder/seed_translations.go
git commit -m "feat(seeder): add translations seed data"
```

---

### Task 12: Seed announcements

**Files:**
- Create: `internal/seeder/seed_announcements.go`

- [ ] **Step 1: Create seed_announcements.go**

```go
package seeder

import (
	"context"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *Seeder) seedAnnouncements(ctx context.Context, count int) error {
	s.logger.Infoc(ctx, "Seeding announcements...", zap.Int("count", count))

	now := time.Now()

	predefined := []struct {
		title   string
		content string
		aType   string
	}{
		{"System Update v2.5", "We have released a major system update with performance improvements and new features.", "info"},
		{"Scheduled Maintenance", "The system will undergo maintenance on Saturday from 02:00 to 06:00 UTC.", "warning"},
		{"New Feature: Data Export", "You can now export your data in CSV, XLSX, and PDF formats from the dashboard.", "info"},
		{"Security Advisory", "Please update your passwords. We have enhanced our security policies.", "critical"},
		{"Welcome to the Platform", "Thank you for joining! Explore our features and let us know your feedback.", "info"},
	}

	types := []string{"info", "warning", "critical"}

	for i, ann := range predefined {
		if i >= count {
			break
		}
		startsAt := now.AddDate(0, 0, -gofakeit.Number(0, 30))
		endsAt := now.AddDate(0, 0, gofakeit.Number(1, 60))

		_, err := s.pool.Exec(ctx,
			`INSERT INTO announcements (id, title, content, type, is_active, starts_at, ends_at, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			uuid.New(), ann.title, ann.content, ann.aType, true, startsAt, endsAt, now, now,
		)
		if err != nil {
			s.logger.Warnc(ctx, "Failed to create predefined announcement", zap.Error(err), zap.String("title", ann.title))
		}
	}

	for i := 0; i < count-len(predefined); i++ {
		if i+len(predefined) >= count {
			break
		}
		startsAt := gofakeit.DateRange(now.AddDate(0, -1, 0), now)
		endsAt := gofakeit.DateRange(now, now.AddDate(0, 2, 0))

		_, err := s.pool.Exec(ctx,
			`INSERT INTO announcements (id, title, content, type, is_active, starts_at, ends_at, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			uuid.New(), gofakeit.Sentence(4), gofakeit.Paragraph(1, 3, 10, " "),
			types[gofakeit.Number(0, len(types)-1)], gofakeit.Bool(), startsAt, endsAt, now, now,
		)
		if err != nil {
			s.logger.Warnc(ctx, "Failed to create random announcement", zap.Error(err))
		}
	}

	return nil
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/seeder/seed_announcements.go
git commit -m "feat(seeder): add announcements seed data"
```

---

### Task 13: Seed notifications

**Files:**
- Create: `internal/seeder/seed_notifications.go`

- [ ] **Step 1: Create seed_notifications.go**

```go
package seeder

import (
	"context"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *Seeder) seedNotifications(ctx context.Context, count int) error {
	s.logger.Infoc(ctx, "Seeding notifications...", zap.Int("count", count))

	now := time.Now()

	types := []string{"info", "warning", "alert"}
	targetTypes := []string{"all", "admin", "user"}

	titles := []string{
		"New login detected", "Password changed successfully", "Your export is ready",
		"Account verification required", "Session expired", "New feature available",
		"Rate limit warning", "Maintenance scheduled", "Role updated",
		"File uploaded successfully", "API key expiring soon", "Security alert",
	}

	bodies := []string{
		"A new login was detected from a new device. If this wasn't you, please change your password.",
		"Your password has been updated successfully. You can now use it to log in.",
		"Your data export has been processed and is ready for download.",
		"Please verify your account to access all features.",
		"Your session has expired. Please log in again to continue.",
		"We have released a new feature. Check it out in your dashboard.",
		"You are approaching the rate limit for API requests.",
		"System maintenance is scheduled. Some services may be temporarily unavailable.",
		"Your role has been updated. You may have new permissions.",
		"Your file has been uploaded and is now available.",
		"One of your API keys is expiring soon. Please renew it.",
		"Unusual activity has been detected on your account.",
	}

	for i := 0; i < count; i++ {
		titleIdx := gofakeit.Number(0, len(titles)-1)

		_, err := s.pool.Exec(ctx,
			`INSERT INTO notifications (id, title, body, type, target_type, is_active, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			uuid.New(), titles[titleIdx], bodies[titleIdx%len(bodies)],
			types[gofakeit.Number(0, len(types)-1)], targetTypes[gofakeit.Number(0, len(targetTypes)-1)],
			true, now, now,
		)
		if err != nil {
			s.logger.Warnc(ctx, "Failed to create notification", zap.Error(err))
		}
	}

	return nil
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/seeder/seed_notifications.go
git commit -m "feat(seeder): add notifications seed data"
```

---

### Task 14: Seed audit logs

**Files:**
- Create: `internal/seeder/seed_audit_logs.go`

- [ ] **Step 1: Create seed_audit_logs.go**

```go
package seeder

import (
	"context"
	"fmt"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *Seeder) seedAuditLogs(ctx context.Context, count int) error {
	s.logger.Infoc(ctx, "Seeding audit logs...", zap.Int("count", count))

	// Get user IDs
	userRows, err := s.pool.Query(ctx, "SELECT id FROM users LIMIT 50")
	if err != nil {
		return fmt.Errorf("failed to get users for audit logs: %w", err)
	}
	defer userRows.Close()

	var userIDs []uuid.UUID
	for userRows.Next() {
		var id uuid.UUID
		if err := userRows.Scan(&id); err != nil {
			return fmt.Errorf("failed to scan user id: %w", err)
		}
		userIDs = append(userIDs, id)
	}

	// Get session IDs
	sessionRows, err := s.pool.Query(ctx, "SELECT id FROM session LIMIT 50")
	if err != nil {
		s.logger.Warnc(ctx, "No sessions found, audit logs will have nil session_id", zap.Error(err))
	}

	var sessionIDs []uuid.UUID
	if sessionRows != nil {
		for sessionRows.Next() {
			var id uuid.UUID
			if err := sessionRows.Scan(&id); err != nil {
				break
			}
			sessionIDs = append(sessionIDs, id)
		}
		sessionRows.Close()
	}

	// Get policy IDs
	policyRows, err := s.pool.Query(ctx, "SELECT id FROM policy LIMIT 50")
	if err != nil {
		s.logger.Warnc(ctx, "No policies found, audit logs will have nil policy_id", zap.Error(err))
	}

	var policyIDs []uuid.UUID
	if policyRows != nil {
		for policyRows.Next() {
			var id uuid.UUID
			if err := policyRows.Scan(&id); err != nil {
				break
			}
			policyIDs = append(policyIDs, id)
		}
		policyRows.Close()
	}

	actions := []string{
		"LOGIN", "LOGOUT", "SESSION_REVOKE",
		"PASSWORD_CHANGE", "ACCESS_GRANTED", "ACCESS_DENIED",
		"POLICY_MATCHED", "POLICY_DENIED",
		"USER_CREATE", "USER_UPDATE", "USER_DELETE",
		"ROLE_ASSIGN", "ROLE_REMOVE", "POLICY_EVALUATED", "ADMIN_CHANGE",
	}
	platforms := []string{"admin", "web", "mobile", "api"}
	resourceTypes := []string{"user", "role", "permission", "policy", "session", "file", "integration"}
	decisions := []string{"ALLOW", "DENY"}
	userAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15",
		"Mozilla/5.0 (Linux; Android 13) AppleWebKit/537.36",
		"curl/8.1.2",
		"PostmanRuntime/7.32.3",
	}

	for i := 0; i < count; i++ {
		var userID *uuid.UUID
		if len(userIDs) > 0 {
			u := userIDs[gofakeit.Number(0, len(userIDs)-1)]
			userID = &u
		}

		var sessionID *uuid.UUID
		if len(sessionIDs) > 0 {
			sid := sessionIDs[gofakeit.Number(0, len(sessionIDs)-1)]
			sessionID = &sid
		}

		var policyID *uuid.UUID
		if len(policyIDs) > 0 && gofakeit.Bool() {
			pid := policyIDs[gofakeit.Number(0, len(policyIDs)-1)]
			policyID = &pid
		}

		action := actions[gofakeit.Number(0, len(actions)-1)]
		resourceType := resourceTypes[gofakeit.Number(0, len(resourceTypes)-1)]
		resourceID := uuid.New()
		platform := platforms[gofakeit.Number(0, len(platforms)-1)]
		ip := fmt.Sprintf("%d.%d.%d.%d", gofakeit.Number(1, 223), gofakeit.Number(0, 255), gofakeit.Number(0, 255), gofakeit.Number(1, 254))
		ua := userAgents[gofakeit.Number(0, len(userAgents)-1)]
		decision := decisions[gofakeit.Number(0, 1)]
		success := decision == "ALLOW"
		var errorMsg *string
		if !success {
			msg := "Access denied: insufficient permissions"
			errorMsg = &msg
		}
		createdAt := gofakeit.DateRange(time.Now().AddDate(0, -1, 0), time.Now())

		_, err := s.pool.Exec(ctx,
			`INSERT INTO audit_log (id, user_id, session_id, action, resource_type, resource_id, platform, ip_address, user_agent, policy_id, decision, success, error_message, created_at)
			 VALUES ($1, $2, $3, $4::audit_action_type, $5, $6, $7, $8::inet, $9, $10, $11, $12, $13, $14)`,
			uuid.New(), userID, sessionID, action, resourceType, resourceID, platform, ip, ua, policyID, decision, success, errorMsg, createdAt,
		)
		if err != nil {
			s.logger.Warnc(ctx, "Failed to create audit log", zap.Error(err), zap.String("action", action))
		}
	}

	return nil
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/seeder/seed_audit_logs.go
git commit -m "feat(seeder): add audit logs seed data"
```

---

### Task 15: Build and verify

**Files:** None (verification only)

- [ ] **Step 1: Build the project**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go build ./...`
Expected: clean build, no errors

- [ ] **Step 2: Run go vet**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go vet ./internal/seeder/...`
Expected: no issues

- [ ] **Step 3: Final commit (if any fixes needed)**

If there were any compilation fixes, commit them:
```bash
git add -A
git commit -m "fix(seeder): resolve compilation issues"
```
