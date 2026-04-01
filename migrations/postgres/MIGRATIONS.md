# Postgres Migrations

## Tables

### role

| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK, default gen_random_uuid() |
| name | VARCHAR | NOT NULL, UNIQUE |
| created_at | TIMESTAMP | default CURRENT_TIMESTAMP |

### permission

| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK, default gen_random_uuid() |
| parent_id | UUID | FK -> permission(id), nullable |
| name | VARCHAR | NOT NULL, UNIQUE(parent_id, name) |
| created_at | TIMESTAMP | default CURRENT_TIMESTAMP |

### role_permission

| Column | Type | Constraints |
|--------|------|-------------|
| role_id | UUID | PK, FK -> role(id) ON DELETE CASCADE |
| permission_id | UUID | PK, FK -> permission(id) ON DELETE CASCADE |
| created_at | TIMESTAMP | default CURRENT_TIMESTAMP |

### scope

| Column | Type | Constraints |
|--------|------|-------------|
| path | VARCHAR | PK |
| method | VARCHAR | PK |
| created_at | TIMESTAMP | default CURRENT_TIMESTAMP |

### permission_scope

| Column | Type | Constraints |
|--------|------|-------------|
| permission_id | UUID | PK, FK -> permission(id) ON DELETE CASCADE |
| path | VARCHAR | PK, FK -> scope(path, method) |
| method | VARCHAR | PK, FK -> scope(path, method) |
| created_at | TIMESTAMP | default CURRENT_TIMESTAMP |

### users

| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK, default gen_random_uuid() |
| role_id | UUID | FK -> role(id) |
| username | VARCHAR | UNIQUE |
| email | VARCHAR | UNIQUE |
| phone | VARCHAR | UNIQUE |
| password_hash | TEXT | |
| salt | VARCHAR | |
| active | BOOLEAN | default TRUE |
| is_approved | BOOLEAN | default FALSE |
| last_seen | TIMESTAMP | |
| deleted_at | BIGINT | default 0 |
| created_at | TIMESTAMP | default CURRENT_TIMESTAMP |
| updated_at | TIMESTAMP | default CURRENT_TIMESTAMP |

Indexes: `idx_users_phone`, `idx_users_username`, `idx_users_deleted_at`

### session

| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK, default gen_random_uuid() |
| user_id | UUID | NOT NULL, FK -> users(id) ON DELETE CASCADE |
| device_id | UUID | NOT NULL |
| device_name | VARCHAR(255) | |
| device_type | session_device_type | |
| ip_address | INET | |
| user_agent | VARCHAR(512) | |
| fcm_token | VARCHAR(512) | |
| refresh_token_hash | VARCHAR(512) | |
| expires_at | TIMESTAMP | NOT NULL |
| last_activity | TIMESTAMP | NOT NULL, default NOW() |
| revoked | BOOLEAN | NOT NULL, default FALSE |
| os | VARCHAR(100) | |
| os_version | VARCHAR(50) | |
| browser | VARCHAR(100) | |
| browser_version | VARCHAR(50) | |
| created_at | TIMESTAMP | NOT NULL, default NOW() |
| updated_at | TIMESTAMP | NOT NULL, default NOW() |

Indexes: `idx_session_user_id`, `idx_session_device_id`, `idx_session_expires_at`, `idx_session_last_activity`, `idx_session_revoked` (partial: WHERE revoked = FALSE), `idx_session_os`, `idx_session_browser`

### relation

| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK, default gen_random_uuid() |
| type | relation_types | NOT NULL |
| name | VARCHAR | NOT NULL, UNIQUE(type, name) |
| created_at | TIMESTAMP | default CURRENT_TIMESTAMP |

### user_relation

| Column | Type | Constraints |
|--------|------|-------------|
| user_id | UUID | PK, FK -> users(id) ON DELETE CASCADE |
| relation_id | UUID | PK, FK -> relation(id) ON DELETE CASCADE |
| created_at | TIMESTAMP | default CURRENT_TIMESTAMP |

### policy

| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK, default gen_random_uuid() |
| permission_id | UUID | FK -> permission(id) ON DELETE CASCADE |
| effect | policy_effect | NOT NULL |
| priority | INT | default 100 |
| active | BOOLEAN | default TRUE |
| created_at | TIMESTAMP | default CURRENT_TIMESTAMP |

Indexes: `idx_policy_permission_id`

### audit_log

| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK, default gen_random_uuid() |
| user_id | UUID | FK -> users(id) |
| session_id | UUID | FK -> session(id) |
| action | audit_action_type | NOT NULL |
| resource_type | VARCHAR(64) | |
| resource_id | UUID | |
| platform | VARCHAR(16) | |
| ip_address | INET | |
| user_agent | VARCHAR(512) | |
| permission | VARCHAR(128) | |
| policy_id | UUID | FK -> policy(id) |
| decision | VARCHAR(16) | |
| success | BOOLEAN | NOT NULL |
| error_message | TEXT | |
| created_at | TIMESTAMP | NOT NULL, default NOW() |

Indexes: `idx_audit_user_id`, `idx_audit_session_id`, `idx_audit_action`, `idx_audit_resource`, `idx_audit_created_at`, `idx_audit_decision` (partial), `idx_audit_policy_id` (partial), `idx_audit_failed_attempts` (partial: WHERE success = FALSE)

### endpoint_history

| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK, default gen_random_uuid() |
| user_id | UUID | FK -> users(id) |
| session_id | UUID | FK -> session(id) |
| method | VARCHAR(8) | NOT NULL |
| path | VARCHAR(255) | NOT NULL |
| status_code | SMALLINT | NOT NULL |
| duration_ms | INTEGER | NOT NULL |
| platform | VARCHAR(16) | |
| ip_address | INET | |
| user_agent | VARCHAR(512) | |
| permission | VARCHAR(128) | |
| decision | VARCHAR(16) | |
| request_id | UUID | |
| rate_limited | BOOLEAN | default FALSE |
| response_size | INTEGER | |
| error_message | TEXT | |
| created_at | TIMESTAMP | NOT NULL, default NOW() |

Indexes: `idx_eh_user_id`, `idx_eh_session_id`, `idx_eh_path`, `idx_eh_method`, `idx_eh_status`, `idx_eh_created_at`, `idx_eh_user_created`, `idx_eh_path_status`, `idx_eh_decision` (partial), `idx_eh_errors` (partial: WHERE status_code >= 500), `idx_eh_slow_requests` (partial: WHERE duration_ms > 1000)

### system_errors

| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK, default gen_random_uuid() |
| code | VARCHAR(64) | NOT NULL |
| message | TEXT | NOT NULL |
| stack_trace | TEXT | |
| severity | VARCHAR(16) | NOT NULL, default 'ERROR' |
| service_name | VARCHAR(64) | default 'api' |
| request_id | UUID | |
| user_id | UUID | |
| ip_address | INET | |
| path | VARCHAR(255) | |
| method | VARCHAR(8) | |
| is_resolved | BOOLEAN | default FALSE |
| resolved_at | TIMESTAMP | |
| resolved_by | UUID | |
| created_at | TIMESTAMP | NOT NULL, default NOW() |

Indexes: `idx_sys_err_code`, `idx_sys_err_severity`, `idx_sys_err_created_at`, `idx_sys_err_req_id` (partial), `idx_sys_err_resolved`

### function_metrics

| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK, default gen_random_uuid() |
| name | VARCHAR(255) | NOT NULL |
| latency_ms | INTEGER | NOT NULL |
| is_panic | BOOLEAN | default FALSE |
| panic_error | TEXT | |
| created_at | TIMESTAMP | NOT NULL, default NOW() |

Indexes: `idx_func_metrics_name`, `idx_func_metrics_created_at`, `idx_func_metrics_panic`

### site_settings

| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK, default gen_random_uuid() |
| key | VARCHAR(64) | NOT NULL, UNIQUE |
| value | TEXT | |
| value_type | VARCHAR(16) | NOT NULL, default 'string' |
| category | VARCHAR(32) | NOT NULL, default 'general' |
| description | TEXT | |
| is_public | BOOLEAN | default FALSE |
| created_at | TIMESTAMP | NOT NULL, default NOW() |
| updated_at | TIMESTAMP | NOT NULL, default NOW() |

Indexes: `idx_site_settings_key`, `idx_site_settings_category`, `idx_site_settings_public`

### error_code

| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK, default gen_random_uuid() |
| code | VARCHAR(255) | NOT NULL, UNIQUE |
| message | TEXT | NOT NULL |
| http_status | INT | NOT NULL |
| category | error_category_enum | default 'UNKNOWN' |
| severity | error_severity_enum | default 'MEDIUM' |
| retryable | BOOLEAN | default FALSE |
| retry_after | INT | default 0 |
| suggestion | TEXT | |
| created_at | TIMESTAMP | NOT NULL, default NOW() |
| updated_at | TIMESTAMP | NOT NULL, default NOW() |

Indexes: `idx_error_code_code`, `idx_error_code_category`

### integrations

| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK, default gen_random_uuid() |
| name | VARCHAR(100) | NOT NULL, UNIQUE |
| description | TEXT | |
| base_url | VARCHAR(500) | NOT NULL |
| is_active | BOOLEAN | NOT NULL, default true |
| created_at | TIMESTAMPTZ | NOT NULL, default NOW() |
| updated_at | TIMESTAMPTZ | NOT NULL, default NOW() |
| deleted_at | TIMESTAMPTZ | |

Indexes: `idx_integrations_name` (partial: WHERE deleted_at IS NULL), `idx_integrations_is_active` (partial)

### api_keys

| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK, default gen_random_uuid() |
| integration_id | UUID | NOT NULL, FK -> integrations(id) ON DELETE CASCADE |
| name | VARCHAR(100) | NOT NULL |
| key | VARCHAR(255) | NOT NULL, UNIQUE (hashed) |
| key_prefix | VARCHAR(20) | NOT NULL |
| is_active | BOOLEAN | NOT NULL, default true |
| expires_at | TIMESTAMPTZ | |
| last_used_at | TIMESTAMPTZ | |
| created_at | TIMESTAMPTZ | NOT NULL, default NOW() |
| updated_at | TIMESTAMPTZ | NOT NULL, default NOW() |
| deleted_at | TIMESTAMPTZ | |

Indexes: `idx_api_keys_integration_id` (partial), `idx_api_keys_key` (partial: WHERE deleted_at IS NULL AND is_active = true), `idx_api_keys_key_prefix` (partial)

### entity_metadata

Universal EAV (Entity-Attribute-Value) table that replaces all former JSONB columns (attributes, data, conditions, metadata, config, etc.).

| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK, default gen_random_uuid() |
| entity_type | VARCHAR(64) | NOT NULL |
| entity_id | UUID | NOT NULL |
| key | VARCHAR(128) | NOT NULL |
| value | TEXT | NOT NULL, default '' |
| value_type | VARCHAR(16) | NOT NULL, default 'string' |
| created_at | TIMESTAMPTZ | NOT NULL, default NOW() |
| updated_at | TIMESTAMPTZ | NOT NULL, default NOW() |

Unique constraint: `(entity_type, entity_id, key)`

Indexes: `idx_entity_metadata_lookup` (entity_type, entity_id), `idx_entity_metadata_type` (entity_type)

## Enums

| Type | Values |
|------|--------|
| `relation_types` | UNREVEALED, BRANCH, REGION |
| `policy_effect` | ALLOW, DENY |
| `session_device_type` | DESKTOP, MOBILE, TABLET, BOT, TV |
| `audit_action_type` | LOGIN, LOGOUT, SESSION_REVOKE, PASSWORD_CHANGE, MFA_VERIFY_FAIL, ACCESS_GRANTED, ACCESS_DENIED, POLICY_MATCHED, POLICY_DENIED, USER_CREATE, USER_UPDATE, USER_DELETE, ROLE_ASSIGN, ROLE_REMOVE, ORDER_APPROVE, ORDER_CANCEL, PAYMENT_PROCESS, PAYMENT_CANCEL, POLICY_EVALUATED, ADMIN_CHANGE |
| `error_severity_enum` | LOW, MEDIUM, HIGH, CRITICAL |
| `error_category_enum` | DATA, AUTH, SYSTEM, VALIDATION, BUSINESS, UNKNOWN |

## Triggers

| Trigger | Table | Event | Function |
|---------|-------|-------|----------|
| `invalidate_cache_*` | users, role, permission, policy, session, relation, integrations, api_keys | INSERT/UPDATE/DELETE | `notify_cache_invalidation()` — pg_notify on `cache_invalidation` channel |
| `trigger_integrations_updated_at` | integrations | UPDATE | `update_integrations_updated_at()` |
| `trigger_api_keys_updated_at` | api_keys | UPDATE | `update_api_keys_updated_at()` |

## Seed Data

**Roles:** super_admin, admin, manager, user, auditor, hr, support, developer, viewer

**Default Users:**
- `admin` (admin@test.com) — role: admin
- `viewer_demo` (viewer@example.com) — role: viewer

**Relations:** Tashkent, Samarkand, Fergana (REGION) / Chilonzor, Yunusobod, Mirzo Ulugbek (BRANCH)

**Site Settings:** site_name, site_description, maintenance_mode, maintenance_message, allow_registration, max_upload_size, session_timeout, admin_email, smtp_enabled, items_per_page

## Migration Files

| File | Description |
|------|-------------|
| `20260101000000_init_schema.sql` | Base schema: roles, permissions, users, sessions, relations, policies |
| `20260101010000_seed_data.sql` | Seed roles, permissions, relations |
| `20260101020000_create_audit_log.sql` | Audit log table + audit_action_type enum |
| `20260101030000_create_endpoint_history.sql` | HTTP request history table |
| `20260101040000_create_system_errors.sql` | System errors table |
| `20260101050000_create_function_metrics.sql` | Function metrics table |
| `20260104170000_cache_invalidation.sql` | Cache invalidation triggers (pg_notify) |
| `20260104190000_add_admin_change_audit_type.sql` | Add ADMIN_CHANGE to audit_action_type |
| `20260112100000_add_is_approved.sql` | Add is_approved column to users |
| `20260112130000_create_site_settings.sql` | Site settings table + defaults |
| `20260114210000_add_session_device_details.sql` | Add os, browser columns to session |
| `20260124235500_create_error_codes.sql` | Error codes table + enums |
| `20260128000000_ci_seed.sql` | CI admin user seed |
| `20260207000000_create_integrations.sql` | Integrations + API keys tables |
| `20260207000001_fix_missing_columns.sql` | Fix missing columns |
| `20260207999999_create_entity_metadata.sql` | Create entity_metadata EAV table |
| `20260207000002_add_trig_integration.sql` | Cache triggers for integrations |
| `20260208000000_seed_authz.sql` | Full authz seed (scopes, permissions, roles, policies) |
| `20260401000000_jsonb_to_eav.sql` | Migrate all JSONB columns to entity_metadata EAV table |
