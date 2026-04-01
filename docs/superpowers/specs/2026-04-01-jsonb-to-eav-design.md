# JSONB → Universal EAV Migration Design

## Summary

Replace all 9 JSONB columns across the codebase with a single universal `entity_metadata` EAV table. Each BC maintains DDD boundary by defining its own domain interface, while sharing a reusable `GenericMetadataRepo` implementation via composition.

## Motivation

- JSONB columns violate project convention (no JSONB policy)
- A single EAV table provides a consistent, queryable, normalized pattern
- Aligns with existing `site_settings` EAV pattern already in the codebase

---

## Database Schema

### New Table: `entity_metadata`

```sql
CREATE TABLE IF NOT EXISTS entity_metadata (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type VARCHAR(64)  NOT NULL,
    entity_id   UUID         NOT NULL,
    key         VARCHAR(128) NOT NULL,
    value       TEXT         NOT NULL DEFAULT '',
    value_type  VARCHAR(16)  NOT NULL DEFAULT 'string',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (entity_type, entity_id, key)
);
CREATE INDEX idx_entity_metadata_lookup ON entity_metadata(entity_type, entity_id);
CREATE INDEX idx_entity_metadata_type ON entity_metadata(entity_type);
```

### Entity Type Constants

| Constant               | Value                    | Source JSONB Column       |
|------------------------|--------------------------|---------------------------|
| `UserAttributes`       | `user_attributes`        | `users.attributes`        |
| `SessionData`          | `session_data`           | `session.data`            |
| `PolicyConditions`     | `policy_conditions`      | `policy.conditions`       |
| `AuditLogMetadata`     | `audit_log_metadata`     | `audit_log.metadata`      |
| `SystemErrorMetadata`  | `system_error_metadata`  | `system_errors.metadata`  |
| `IntegrationConfig`    | `integration_config`     | `integrations.config`     |
| `JobPayload`           | `job_payload`            | `jobs.payload`            |
| `TranslationData`      | `translation_data`       | `translations.data`       |
| `DataExportFilters`    | `data_export_filters`    | `data_exports.filters`    |

### Value Types

| value_type   | Go parse method          | Example value                          |
|-------------|--------------------------|----------------------------------------|
| `string`    | direct                   | `"John"`                               |
| `integer`   | `strconv.Atoi`           | `"42"`                                 |
| `boolean`   | `strconv.ParseBool`      | `"true"`                               |
| `float`     | `strconv.ParseFloat`     | `"3.14"`                               |
| `json_array`| `json.Unmarshal`         | `["127.0.0.1","192.168.1.1"]`          |

---

## Go Architecture

### Shared Implementation

```
internal/shared/infrastructure/metadata/
├── postgres.go        -- GenericMetadataRepo struct + methods
└── entity_types.go    -- entity_type string constants
```

**`GenericMetadataRepo`** — reusable, not exposed as a shared interface. Each BC composes it internally.

```go
type GenericMetadataRepo struct {
    pool    *pgxpool.Pool
    builder squirrel.StatementBuilderType
}

func NewGenericMetadataRepo(pool *pgxpool.Pool) *GenericMetadataRepo

// Methods:
func (r *GenericMetadataRepo) SetMany(ctx context.Context, entityType string, entityID uuid.UUID, entries map[string]string) error
func (r *GenericMetadataRepo) GetAll(ctx context.Context, entityType string, entityID uuid.UUID) (map[string]string, error)
func (r *GenericMetadataRepo) DeleteAll(ctx context.Context, entityType string, entityID uuid.UUID) error
func (r *GenericMetadataRepo) DeleteKeys(ctx context.Context, entityType string, entityID uuid.UUID, keys []string) error
```

**`SetMany` behavior:** UPSERT — `INSERT ... ON CONFLICT (entity_type, entity_id, key) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()`

### BC Integration Pattern (DDD-compliant)

Each BC owns its domain interface. The infrastructure layer composes `GenericMetadataRepo` as a private field.

**Example: User BC**

Domain layer (`internal/user/domain/repository.go`):
```go
// No change to existing UserRepository interface.
// Attributes are part of the User aggregate — the repo handles persistence internally.
```

Infrastructure layer (`internal/user/infrastructure/postgres/write_repo.go`):
```go
type UserWriteRepo struct {
    pool     *pgxpool.Pool
    builder  squirrel.StatementBuilderType
    metadata *metadata.GenericMetadataRepo  // composed, private
}

func NewUserWriteRepo(pool *pgxpool.Pool) *UserWriteRepo {
    return &UserWriteRepo{
        pool:     pool,
        builder:  squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
        metadata: metadata.NewGenericMetadataRepo(pool),
    }
}

// In Save():
//   OLD: configJSON, _ := json.Marshal(user.Attributes()) → insert JSONB
//   NEW: r.metadata.SetMany(ctx, metadata.UserAttributes, user.ID(), attrsAsStringMap)

// In FindByID() / List():
//   OLD: scan attrsJSON []byte → json.Unmarshal
//   NEW: attrs, _ := r.metadata.GetAll(ctx, metadata.UserAttributes, userID)

// In Delete():
//   OLD: (attributes deleted with row)
//   NEW: r.metadata.DeleteAll(ctx, metadata.UserAttributes, userID) + delete row
```

---

## Per-BC Change Summary

### Active BCs (6) — migrate JSONB → EAV

| BC | JSONB Column | Domain Type Change | Repo Change |
|---|---|---|---|
| `user` | `users.attributes` | `map[string]any` → `map[string]string` | `json.Marshal` → `metadata.SetMany` |
| `authz` | `policy.conditions` | `map[string]any` → `map[string]string` | `json.Marshal` → `metadata.SetMany` |
| `audit` | `audit_log.metadata` | `map[string]any` → `map[string]string` | `json.Marshal` → `metadata.SetMany` |
| `systemerror` | `system_errors.metadata` | `map[string]any` → `map[string]string` | `json.Marshal` → `metadata.SetMany` |
| `integration` | `integrations.config` | `map[string]any` → `map[string]string` | `json.Marshal` → `metadata.SetMany` |
| `dataexport` | `data_exports.filters` | Remove unused `filtersJSON` scan | `metadata.SetMany` (when filters are used) |

### Unused BCs (3) — DROP COLUMN only

| BC | JSONB Column | Reason |
|---|---|---|
| `session` | `session.data` | Not used in Go code |
| `job` | `jobs.payload` | Not used in Go code |
| `translation` | `translations.data` | Read as `::text`, not JSONB |

### Seed Data

`seed_authz.sql` — JSONB insert values become `entity_metadata` INSERT rows:

```sql
-- OLD:
-- INSERT INTO policy (..., conditions) VALUES (..., '{"user.role_name": "auditor"}'::jsonb);

-- NEW:
-- INSERT INTO policy (...) VALUES (...);  -- no conditions column
-- INSERT INTO entity_metadata (entity_type, entity_id, key, value, value_type)
-- VALUES ('policy_conditions', <policy_id>, 'user.role_name', 'auditor', 'string');
```

For array values like `{"env.ip_not_in": ["127.0.0.1", "192.168.1.1"]}`:
```sql
INSERT INTO entity_metadata (entity_type, entity_id, key, value, value_type)
VALUES ('policy_conditions', <policy_id>, 'env.ip_not_in', '["127.0.0.1","192.168.1.1"]', 'json_array');
```

---

## Migration Strategy

Single migration file: `migrations/postgres/YYYYMMDD_jsonb_to_eav.sql`

### Steps:
1. **CREATE** `entity_metadata` table with indexes
2. **MIGRATE** data from each JSONB column into `entity_metadata` rows (using `jsonb_each_text`)
3. **DROP** all 9 JSONB columns
4. **UPDATE** `seed_authz.sql` to use `entity_metadata` inserts

### Data migration SQL pattern:
```sql
-- Example: users.attributes → entity_metadata
INSERT INTO entity_metadata (entity_type, entity_id, key, value, value_type)
SELECT 'user_attributes', id, kv.key, kv.value, 'string'
FROM users, jsonb_each_text(attributes) AS kv
WHERE attributes IS NOT NULL AND attributes != '{}'::jsonb
ON CONFLICT DO NOTHING;
```

---

## Files Changed

### New files:
- `internal/shared/infrastructure/metadata/postgres.go`
- `internal/shared/infrastructure/metadata/entity_types.go`
- `migrations/postgres/YYYYMMDD_jsonb_to_eav.sql`

### Modified files:

**Domain layer (type changes):**
- `internal/user/domain/entity.go` — `map[string]any` → `map[string]string`
- `internal/authz/domain/policy.go` — `map[string]any` → `map[string]string`
- `internal/audit/domain/entity.go` — `map[string]any` → `map[string]string`
- `internal/systemerror/domain/entity.go` — `map[string]any` → `map[string]string`
- `internal/integration/domain/entity.go` — `map[string]any` → `map[string]string`

**Infrastructure layer (repo changes):**
- `internal/user/infrastructure/postgres/write_repo.go`
- `internal/user/infrastructure/postgres/read_repo.go`
- `internal/authz/infrastructure/postgres/write_repo.go`
- `internal/authz/infrastructure/postgres/read_repo.go`
- `internal/audit/infrastructure/postgres/write_repo.go`
- `internal/audit/infrastructure/postgres/read_repo.go`
- `internal/systemerror/infrastructure/postgres/write_repo.go`
- `internal/systemerror/infrastructure/postgres/read_repo.go`
- `internal/integration/infrastructure/postgres/write_repo.go`
- `internal/dataexport/infrastructure/postgres/write_repo.go`

**Application layer (DTO type changes):**
- `internal/user/application/dto.go` — `Attributes map[string]any` → `map[string]string`
- `internal/audit/application/dto.go` (if exists)
- `internal/systemerror/application/dto.go` (if exists)
- `internal/integration/application/dto.go` (if exists)

**HTTP layer (request/response type changes):**
- Handlers that accept/return attributes, conditions, metadata, config

**Migrations:**
- `migrations/postgres/20260208000000_seed_authz.sql` — update inserts
- `migrations/postgres/20260207000001_fix_missing_columns.sql` — remove (adds JSONB columns)

**Docs:**
- `migrations/postgres/MIGRATIONS.md` — update schema docs
- `docs/swagger/` — update swagger annotations

**Backup file:**
- `migrations/postgres/20260101000000_init_schema.sql.bak` — update or leave as-is (backup)

---

## Testing

- Existing integration tests that create users/policies/audit logs with attributes/conditions/metadata will need updating to use `map[string]string`
- `GenericMetadataRepo` gets its own unit tests
- Seed data migration verified by running full migration suite on clean DB
