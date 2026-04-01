# Extend Seeder for All Essential Bounded Contexts

## Overview

Extend the existing seeder (`internal/seeder/`) to generate dummy data for 12 additional bounded contexts beyond the current users/roles/permissions/policies.

## Approach

**Option B: Separate seed files per BC** — each BC gets its own `seed_xxx.go` file inside `internal/seeder/`. The main `seeder.go` orchestrates them in dependency order.

## Files to Create

```
internal/seeder/
├── seeder.go                    (modify — add orchestration calls)
├── seed_site_settings.go        (new)
├── seed_error_codes.go          (new)
├── seed_feature_flags.go        (new)
├── seed_rate_limits.go          (new)
├── seed_ip_rules.go             (new)
├── seed_integrations.go         (new)
├── seed_translations.go         (new)
├── seed_file_metadata.go        (new)
├── seed_announcements.go        (new)
├── seed_notifications.go        (new)
├── seed_audit_logs.go           (new)
├── seed_function_metrics.go     (new)
```

## Files to Modify

- `config/seeder.go` — add count fields for each new BC
- `internal/shared/infrastructure/asynq/handlers.go` — extend `SeedPayload` struct
- `cmd/seeder/main.go` — pass new counts to payload
- `internal/seeder/seeder.go` — call new seed functions, clear new tables

## Seed Order (FK dependency)

**Phase 1 — Independent tables:**
1. Site Settings (15)
2. Error Codes (20)
3. Feature Flags (15)
4. Rate Limits (8)
5. IP Rules (10)
6. Function Metrics (30)

**Phase 2 — Dependent on users:**
7. Integrations + API Keys (5)
8. File Metadata (20) — references `users.id` via `uploaded_by`
9. Translations (50) — references entity types
10. Announcements (10)
11. Notifications (30)

**Phase 3 — Dependent on users + sessions:**
12. Audit Logs (50) — references `users.id`, `session.id`, `policy.id`

## Data Clearing Order

Reverse of seed order. Add to existing `clearData()`:

```
function_metrics, audit_log (already cleared), notifications, announcements,
translations, file_metadata, integrations, api_keys, ip_rules, rate_limits,
feature_flags, error_code, site_settings
```

## Insert Pattern

All seed functions follow the existing pattern — direct SQL via Squirrel builder + `pool.Exec()`. No domain layer involvement.

Each `seedXXX()` method:
1. Receives `ctx`, count, and any dependent IDs (e.g., userIDs)
2. Uses `gofakeit` for random data
3. Inserts via Squirrel `Insert().Columns().Values().ToSql()`
4. Returns generated IDs if needed by downstream seeders
5. Logs progress via `s.logger`

## Sample Data Strategy

| BC | Data Description |
|---|---|
| **Site Settings** | Realistic keys: `site_name`, `maintenance_mode`, `smtp_host`, etc. |
| **Error Codes** | Structured codes: `AUTH_001`, `DATA_001`, `SYS_001` with proper HTTP statuses |
| **Feature Flags** | Dev flags: `dark_mode`, `new_dashboard`, `beta_api_v2`, etc. |
| **Rate Limits** | Path patterns: `/api/v1/*`, `/auth/login`, `/upload/*` |
| **IP Rules** | Mix of allow/block with realistic IPs and CIDRs |
| **Integrations** | Named services: Slack, Telegram, Email, Webhook, Custom |
| **Translations** | Keys like `auth.login_success`, `error.not_found` in uz/ru/en |
| **File Metadata** | Realistic filenames with proper MIME types and sizes |
| **Announcements** | System updates, maintenance notices, feature releases |
| **Notifications** | INFO/WARNING/ALERT with realistic messages |
| **Audit Logs** | LOGIN, LOGOUT, ACCESS_GRANTED actions tied to real user/session IDs |
| **Function Metrics** | Realistic latencies (5-500ms), occasional panics |

## Config Additions

```go
// config/seeder.go additions
AnnouncementsCount    int  // SEEDER_ANNOUNCEMENTS_COUNT (default: 10)
NotificationsCount    int  // SEEDER_NOTIFICATIONS_COUNT (default: 30)
FeatureFlagsCount     int  // SEEDER_FEATURE_FLAGS_COUNT (default: 15)
IntegrationsCount     int  // SEEDER_INTEGRATIONS_COUNT (default: 5)
TranslationsCount     int  // SEEDER_TRANSLATIONS_COUNT (default: 50)
FileMetadataCount     int  // SEEDER_FILE_METADATA_COUNT (default: 20)
SiteSettingsCount     int  // SEEDER_SITE_SETTINGS_COUNT (default: 15)
ErrorCodesCount       int  // SEEDER_ERROR_CODES_COUNT (default: 20)
IPRulesCount          int  // SEEDER_IP_RULES_COUNT (default: 10)
RateLimitsCount       int  // SEEDER_RATE_LIMITS_COUNT (default: 8)
AuditLogsCount        int  // SEEDER_AUDIT_LOGS_COUNT (default: 50)
FunctionMetricsCount  int  // SEEDER_FUNCTION_METRICS_COUNT (default: 30)
```
