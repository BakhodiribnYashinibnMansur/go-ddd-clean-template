# JSONB → Universal EAV Migration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace all 9 JSONB columns with a single `entity_metadata` EAV table, maintaining DDD boundary isolation.

**Architecture:** Shared `GenericMetadataRepo` implementation composed into each BC's infrastructure layer. Domain types change from `map[string]any` to `map[string]string`. Migration moves existing JSONB data via `jsonb_each_text`.

**Tech Stack:** Go, PostgreSQL, pgx/v5, Squirrel, goose migrations

---

### Task 1: Create shared `GenericMetadataRepo` and entity type constants

**Files:**
- Create: `internal/shared/infrastructure/metadata/entity_types.go`
- Create: `internal/shared/infrastructure/metadata/postgres.go`
- Modify: `internal/shared/domain/consts/tables.go`

- [ ] **Step 1: Add table constant to consts**

In `internal/shared/domain/consts/tables.go`, add `TableEntityMetadata`:

```go
TableEntityMetadata  = "entity_metadata"
```

Add it after the existing `TableUserSettings` line (before the closing parenthesis).

- [ ] **Step 2: Create entity type constants file**

Create `internal/shared/infrastructure/metadata/entity_types.go`:

```go
package metadata

// Entity type constants for the entity_metadata EAV table.
const (
	EntityTypeUserAttributes     = "user_attributes"
	EntityTypeSessionData        = "session_data"
	EntityTypePolicyConditions   = "policy_conditions"
	EntityTypeAuditLogMetadata   = "audit_log_metadata"
	EntityTypeSystemErrorMeta    = "system_error_metadata"
	EntityTypeIntegrationConfig  = "integration_config"
	EntityTypeJobPayload         = "job_payload"
	EntityTypeTranslationData    = "translation_data"
	EntityTypeDataExportFilters  = "data_export_filters"
)
```

- [ ] **Step 3: Create GenericMetadataRepo implementation**

Create `internal/shared/infrastructure/metadata/postgres.go`:

```go
package metadata

import (
	"context"
	"time"

	"gct/internal/shared/domain/consts"
	apperrors "gct/internal/shared/infrastructure/errors"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GenericMetadataRepo provides CRUD for the entity_metadata EAV table.
// It is not a domain interface — BCs compose it as a private field in their infra repos.
type GenericMetadataRepo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

// NewGenericMetadataRepo creates a new GenericMetadataRepo.
func NewGenericMetadataRepo(pool *pgxpool.Pool) *GenericMetadataRepo {
	return &GenericMetadataRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// SetMany upserts multiple key-value pairs for a given entity.
func (r *GenericMetadataRepo) SetMany(ctx context.Context, entityType string, entityID uuid.UUID, entries map[string]string) error {
	if len(entries) == 0 {
		return nil
	}

	now := time.Now()
	qb := r.builder.
		Insert(consts.TableEntityMetadata).
		Columns("entity_type", "entity_id", "key", "value", "created_at", "updated_at")

	for k, v := range entries {
		qb = qb.Values(entityType, entityID, k, v, now, now)
	}

	qb = qb.Suffix("ON CONFLICT (entity_type, entity_id, key) DO UPDATE SET value = EXCLUDED.value, updated_at = EXCLUDED.updated_at")

	sql, args, err := qb.ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildInsert)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, consts.TableEntityMetadata, nil)
	}

	return nil
}

// SetManyTx upserts multiple key-value pairs within an existing transaction.
func (r *GenericMetadataRepo) SetManyTx(ctx context.Context, tx pgx.Tx, entityType string, entityID uuid.UUID, entries map[string]string) error {
	if len(entries) == 0 {
		return nil
	}

	now := time.Now()
	qb := r.builder.
		Insert(consts.TableEntityMetadata).
		Columns("entity_type", "entity_id", "key", "value", "created_at", "updated_at")

	for k, v := range entries {
		qb = qb.Values(entityType, entityID, k, v, now, now)
	}

	qb = qb.Suffix("ON CONFLICT (entity_type, entity_id, key) DO UPDATE SET value = EXCLUDED.value, updated_at = EXCLUDED.updated_at")

	sql, args, err := qb.ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildInsert)
	}

	if _, err = tx.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, consts.TableEntityMetadata, nil)
	}

	return nil
}

// GetAll retrieves all key-value pairs for an entity.
func (r *GenericMetadataRepo) GetAll(ctx context.Context, entityType string, entityID uuid.UUID) (map[string]string, error) {
	sql, args, err := r.builder.
		Select("key", "value").
		From(consts.TableEntityMetadata).
		Where(squirrel.Eq{"entity_type": entityType, "entity_id": entityID}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableEntityMetadata, nil)
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, apperrors.HandlePgError(err, consts.TableEntityMetadata, nil)
		}
		result[k] = v
	}

	return result, nil
}

// GetAllTx retrieves all key-value pairs within an existing transaction.
func (r *GenericMetadataRepo) GetAllTx(ctx context.Context, tx pgx.Tx, entityType string, entityID uuid.UUID) (map[string]string, error) {
	sql, args, err := r.builder.
		Select("key", "value").
		From(consts.TableEntityMetadata).
		Where(squirrel.Eq{"entity_type": entityType, "entity_id": entityID}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	rows, err := tx.Query(ctx, sql, args...)
	if err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableEntityMetadata, nil)
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, apperrors.HandlePgError(err, consts.TableEntityMetadata, nil)
		}
		result[k] = v
	}

	return result, nil
}

// DeleteAll removes all metadata for an entity.
func (r *GenericMetadataRepo) DeleteAll(ctx context.Context, entityType string, entityID uuid.UUID) error {
	sql, args, err := r.builder.
		Delete(consts.TableEntityMetadata).
		Where(squirrel.Eq{"entity_type": entityType, "entity_id": entityID}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildDelete)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, consts.TableEntityMetadata, nil)
	}

	return nil
}

// DeleteAllTx removes all metadata within an existing transaction.
func (r *GenericMetadataRepo) DeleteAllTx(ctx context.Context, tx pgx.Tx, entityType string, entityID uuid.UUID) error {
	sql, args, err := r.builder.
		Delete(consts.TableEntityMetadata).
		Where(squirrel.Eq{"entity_type": entityType, "entity_id": entityID}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildDelete)
	}

	if _, err = tx.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, consts.TableEntityMetadata, nil)
	}

	return nil
}
```

- [ ] **Step 4: Verify it compiles**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go build ./internal/shared/...`
Expected: Clean build, no errors.

- [ ] **Step 5: Commit**

```bash
git add internal/shared/infrastructure/metadata/entity_types.go internal/shared/infrastructure/metadata/postgres.go internal/shared/domain/consts/tables.go
git commit -m "feat: add GenericMetadataRepo and entity_metadata EAV infrastructure"
```

---

### Task 2: Create database migration — entity_metadata table, data migration, DROP columns

**Files:**
- Create: `migrations/postgres/20260401000000_jsonb_to_eav.sql`

- [ ] **Step 1: Create migration file**

Create `migrations/postgres/20260401000000_jsonb_to_eav.sql`:

```sql
-- +goose Up
-- +goose StatementBegin

-- 1. Create the universal entity_metadata EAV table.
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
CREATE INDEX IF NOT EXISTS idx_entity_metadata_lookup ON entity_metadata(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_entity_metadata_type ON entity_metadata(entity_type);

-- 2. Migrate data from JSONB columns into entity_metadata.

-- users.attributes
INSERT INTO entity_metadata (entity_type, entity_id, key, value, value_type)
SELECT 'user_attributes', id, kv.key, kv.value, 'string'
FROM users, jsonb_each_text(attributes) AS kv
WHERE attributes IS NOT NULL AND attributes != '{}'::jsonb
ON CONFLICT DO NOTHING;

-- policy.conditions (may contain arrays, so use jsonb_each not jsonb_each_text)
INSERT INTO entity_metadata (entity_type, entity_id, key, value, value_type)
SELECT 'policy_conditions', id, kv.key,
    CASE
        WHEN jsonb_typeof(kv.value) = 'array' THEN kv.value::text
        ELSE kv.value #>> '{}'
    END,
    CASE
        WHEN jsonb_typeof(kv.value) = 'array' THEN 'json_array'
        ELSE 'string'
    END
FROM policy, jsonb_each(conditions) AS kv
WHERE conditions IS NOT NULL AND conditions != '{}'::jsonb
ON CONFLICT DO NOTHING;

-- audit_log.metadata
INSERT INTO entity_metadata (entity_type, entity_id, key, value, value_type)
SELECT 'audit_log_metadata', id, kv.key, kv.value, 'string'
FROM audit_log, jsonb_each_text(metadata) AS kv
WHERE metadata IS NOT NULL AND metadata != '{}'::jsonb
ON CONFLICT DO NOTHING;

-- system_errors.metadata
INSERT INTO entity_metadata (entity_type, entity_id, key, value, value_type)
SELECT 'system_error_metadata', id, kv.key, kv.value, 'string'
FROM system_errors, jsonb_each_text(metadata) AS kv
WHERE metadata IS NOT NULL AND metadata != '{}'::jsonb
ON CONFLICT DO NOTHING;

-- integrations.config
INSERT INTO entity_metadata (entity_type, entity_id, key, value, value_type)
SELECT 'integration_config', id, kv.key, kv.value, 'string'
FROM integrations, jsonb_each_text(config) AS kv
WHERE config IS NOT NULL AND config != '{}'::jsonb
ON CONFLICT DO NOTHING;

-- jobs.payload
INSERT INTO entity_metadata (entity_type, entity_id, key, value, value_type)
SELECT 'job_payload', id, kv.key, kv.value, 'string'
FROM jobs, jsonb_each_text(payload) AS kv
WHERE payload IS NOT NULL AND payload != '{}'::jsonb
ON CONFLICT DO NOTHING;

-- translations.data
INSERT INTO entity_metadata (entity_type, entity_id, key, value, value_type)
SELECT 'translation_data', id, kv.key, kv.value, 'string'
FROM translations, jsonb_each_text(data) AS kv
WHERE data IS NOT NULL AND data != '{}'::jsonb
ON CONFLICT DO NOTHING;

-- data_exports.filters
INSERT INTO entity_metadata (entity_type, entity_id, key, value, value_type)
SELECT 'data_export_filters', id, kv.key, kv.value, 'string'
FROM data_exports, jsonb_each_text(filters) AS kv
WHERE filters IS NOT NULL AND filters != '{}'::jsonb
ON CONFLICT DO NOTHING;

-- session.data
INSERT INTO entity_metadata (entity_type, entity_id, key, value, value_type)
SELECT 'session_data', id, kv.key, kv.value, 'string'
FROM session, jsonb_each_text(data) AS kv
WHERE data IS NOT NULL AND data != '{}'::jsonb
ON CONFLICT DO NOTHING;

-- 3. Drop JSONB columns.
ALTER TABLE users DROP COLUMN IF EXISTS attributes;
ALTER TABLE session DROP COLUMN IF EXISTS data;
ALTER TABLE policy DROP COLUMN IF EXISTS conditions;
ALTER TABLE audit_log DROP COLUMN IF EXISTS metadata;
ALTER TABLE system_errors DROP COLUMN IF EXISTS metadata;
ALTER TABLE integrations DROP COLUMN IF EXISTS config;
ALTER TABLE jobs DROP COLUMN IF EXISTS payload;
ALTER TABLE translations DROP COLUMN IF EXISTS data;
ALTER TABLE data_exports DROP COLUMN IF EXISTS filters;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Re-add JSONB columns.
ALTER TABLE users ADD COLUMN IF NOT EXISTS attributes JSONB NOT NULL DEFAULT '{}';
ALTER TABLE session ADD COLUMN IF NOT EXISTS data JSONB;
ALTER TABLE policy ADD COLUMN IF NOT EXISTS conditions JSONB NOT NULL DEFAULT '{}';
ALTER TABLE audit_log ADD COLUMN IF NOT EXISTS metadata JSONB;
ALTER TABLE system_errors ADD COLUMN IF NOT EXISTS metadata JSONB;
ALTER TABLE integrations ADD COLUMN IF NOT EXISTS config JSONB DEFAULT '{}'::jsonb;
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS payload JSONB NOT NULL DEFAULT '{}';
ALTER TABLE translations ADD COLUMN IF NOT EXISTS data JSONB NOT NULL DEFAULT '{}';
ALTER TABLE data_exports ADD COLUMN IF NOT EXISTS filters JSONB NOT NULL DEFAULT '{}';

-- Migrate data back from entity_metadata to JSONB columns.
UPDATE users u SET attributes = COALESCE((
    SELECT jsonb_object_agg(key, value) FROM entity_metadata
    WHERE entity_type = 'user_attributes' AND entity_id = u.id
), '{}');

UPDATE policy p SET conditions = COALESCE((
    SELECT jsonb_object_agg(key,
        CASE WHEN value_type = 'json_array' THEN value::jsonb ELSE to_jsonb(value) END
    ) FROM entity_metadata
    WHERE entity_type = 'policy_conditions' AND entity_id = p.id
), '{}');

UPDATE audit_log a SET metadata = COALESCE((
    SELECT jsonb_object_agg(key, value) FROM entity_metadata
    WHERE entity_type = 'audit_log_metadata' AND entity_id = a.id
), '{}');

UPDATE system_errors s SET metadata = COALESCE((
    SELECT jsonb_object_agg(key, value) FROM entity_metadata
    WHERE entity_type = 'system_error_metadata' AND entity_id = s.id
), '{}');

UPDATE integrations i SET config = COALESCE((
    SELECT jsonb_object_agg(key, value) FROM entity_metadata
    WHERE entity_type = 'integration_config' AND entity_id = i.id
), '{}');

DROP TABLE IF EXISTS entity_metadata;

-- +goose StatementEnd
```

- [ ] **Step 2: Commit**

```bash
git add migrations/postgres/20260401000000_jsonb_to_eav.sql
git commit -m "feat: add migration to create entity_metadata and move JSONB data"
```

---

### Task 3: Update User BC — domain + infrastructure

**Files:**
- Modify: `internal/user/domain/entity.go`
- Modify: `internal/user/domain/repository.go`
- Modify: `internal/user/application/dto.go`
- Modify: `internal/user/infrastructure/postgres/write_repo.go`
- Modify: `internal/user/infrastructure/postgres/read_repo.go`
- Modify: `internal/shared/domain/auth.go`

- [ ] **Step 1: Update User domain entity — map[string]any → map[string]string**

In `internal/user/domain/entity.go`:

Change line 29:
```go
	attributes map[string]any
```
to:
```go
	attributes map[string]string
```

Change line 47-49:
```go
func WithAttributes(attrs map[string]any) UserOption {
	return func(u *User) { u.attributes = attrs }
}
```
to:
```go
func WithAttributes(attrs map[string]string) UserOption {
	return func(u *User) { u.attributes = attrs }
}
```

Change line 61:
```go
		attributes:    make(map[string]any),
```
to:
```go
		attributes:    make(map[string]string),
```

Change the `ReconstructUser` parameter (line 83):
```go
	attributes map[string]any,
```
to:
```go
	attributes map[string]string,
```

Change line 88-90:
```go
	if attributes == nil {
		attributes = make(map[string]any)
	}
```
to:
```go
	if attributes == nil {
		attributes = make(map[string]string)
	}
```

Change line 229:
```go
func (u *User) Attributes() map[string]any { return u.attributes }
```
to:
```go
func (u *User) Attributes() map[string]string { return u.attributes }
```

- [ ] **Step 2: Update UserView in domain repository**

In `internal/user/domain/repository.go`, find the `UserView` struct and change the `Attributes` field from `map[string]any` to `map[string]string`:

```go
Attributes map[string]string `json:"attributes,omitempty"`
```

- [ ] **Step 3: Update application DTO**

In `internal/user/application/dto.go`, change line 16:
```go
	Attributes map[string]any `json:"attributes,omitempty"`
```
to:
```go
	Attributes map[string]string `json:"attributes,omitempty"`
```

- [ ] **Step 4: Update AuthUser in shared domain**

In `internal/shared/domain/auth.go`, change line 39:
```go
	Attributes map[string]any `json:"attributes,omitempty"`
```
to:
```go
	Attributes map[string]string `json:"attributes,omitempty"`
```

- [ ] **Step 5: Update UserWriteRepo — remove JSONB, add metadata composition**

In `internal/user/infrastructure/postgres/write_repo.go`:

Add `"gct/internal/shared/infrastructure/metadata"` to imports. Remove `"encoding/json"` from imports.

Change `userColumns` (lines 27-32) — remove `"attributes"`:
```go
var userColumns = []string{
	"id", "role_id", "username", "email", "phone",
	"password_hash", "salt",
	"active", "is_approved",
	"created_at", "updated_at", "deleted_at", "last_seen",
}
```

Add `metadata` field to the struct (lines 51-54):
```go
type UserWriteRepo struct {
	pool     *pgxpool.Pool
	builder  squirrel.StatementBuilderType
	metadata *metadata.GenericMetadataRepo
}
```

Update the constructor (lines 57-62):
```go
func NewUserWriteRepo(pool *pgxpool.Pool) *UserWriteRepo {
	return &UserWriteRepo{
		pool:     pool,
		builder:  squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		metadata: metadata.NewGenericMetadataRepo(pool),
	}
}
```

Update `Save` method (lines 65-120). Replace the `attrsJSON` marshal block (lines 67-70) and the `attrsJSON` value in the insert (line 94). The full Save method becomes:

```go
func (r *UserWriteRepo) Save(ctx context.Context, user *domain.User) error {
	return pgxutil.WithTx(ctx, r.pool, func(tx pgx.Tx) error {
		var emailVal *string
		if user.Email() != nil {
			v := user.Email().Value()
			emailVal = &v
		}

		var deletedAtVal int64
		if user.DeletedAt() != nil {
			deletedAtVal = user.DeletedAt().Unix()
		}

		sql, args, err := r.builder.
			Insert(usersTable).
			Columns(userColumns...).
			Values(
				user.ID(),
				user.RoleID(),
				user.Username(),
				emailVal,
				user.Phone().Value(),
				user.Password().Hash(),
				"",
				user.IsActive(),
				user.IsApproved(),
				user.CreatedAt(),
				user.UpdatedAt(),
				deletedAtVal,
				user.LastSeen(),
			).
			ToSql()
		if err != nil {
			return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildInsert)
		}

		if _, err = tx.Exec(ctx, sql, args...); err != nil {
			return apperrors.HandlePgError(err, usersTable, nil)
		}

		if err := r.metadata.SetManyTx(ctx, tx, metadata.EntityTypeUserAttributes, user.ID(), user.Attributes()); err != nil {
			return err
		}

		for _, s := range user.Sessions() {
			if err := r.insertSession(ctx, tx, &s); err != nil {
				return err
			}
		}

		return nil
	})
}
```

Update `Update` method (lines 198-262). Remove attrsJSON marshal and the `Set("attributes", attrsJSON)` line. Add metadata SetManyTx call:

```go
func (r *UserWriteRepo) Update(ctx context.Context, user *domain.User) error {
	return pgxutil.WithTx(ctx, r.pool, func(tx pgx.Tx) error {
		var emailVal *string
		if user.Email() != nil {
			v := user.Email().Value()
			emailVal = &v
		}

		var deletedAtVal int64
		if user.DeletedAt() != nil {
			deletedAtVal = user.DeletedAt().Unix()
		}

		sql, args, err := r.builder.
			Update(usersTable).
			Set("role_id", user.RoleID()).
			Set("username", user.Username()).
			Set("email", emailVal).
			Set("phone", user.Phone().Value()).
			Set("password_hash", user.Password().Hash()).
			Set("active", user.IsActive()).
			Set("is_approved", user.IsApproved()).
			Set("updated_at", user.UpdatedAt()).
			Set("deleted_at", deletedAtVal).
			Set("last_seen", user.LastSeen()).
			Where(squirrel.Eq{"id": user.ID()}).
			ToSql()
		if err != nil {
			return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildUpdate)
		}

		if _, err = tx.Exec(ctx, sql, args...); err != nil {
			return apperrors.HandlePgError(err, usersTable, nil)
		}

		if err := r.metadata.SetManyTx(ctx, tx, metadata.EntityTypeUserAttributes, user.ID(), user.Attributes()); err != nil {
			return err
		}

		for _, s := range user.Sessions() {
			upsertSQL, upsertArgs, upsertErr := r.builder.
				Insert(sessionTable).
				Columns(sessionInsertColumns...).
				Values(
					s.ID(), s.UserID(), s.DeviceID(), s.DeviceName(), string(s.DeviceType()),
					s.IPAddress(), s.UserAgent(), s.RefreshTokenHash(),
					s.ExpiresAt(), s.LastActivity(), s.IsRevoked(),
					s.CreatedAt(), s.UpdatedAt(),
				).
				Suffix("ON CONFLICT (id) DO UPDATE SET refresh_token_hash = EXCLUDED.refresh_token_hash, last_activity = EXCLUDED.last_activity, revoked = EXCLUDED.revoked, updated_at = EXCLUDED.updated_at").
				ToSql()
			if upsertErr != nil {
				return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildInsert)
			}
			if _, err = tx.Exec(ctx, upsertSQL, upsertArgs...); err != nil {
				return apperrors.HandlePgError(err, sessionTable, nil)
			}
		}

		return nil
	})
}
```

Update `FindByID` (line 189): replace `user.Attributes()` with a metadata lookup. After scanning the user (line 167), add metadata loading before reconstruct:

```go
func (r *UserWriteRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	sql, args, err := r.builder.
		Select(userColumns...).
		From(usersTable).
		Where(squirrel.Eq{"id": id}).
		Where(squirrel.Eq{"deleted_at": 0}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	row := r.pool.QueryRow(ctx, sql, args...)
	user, err := scanUser(row)
	if err != nil {
		return nil, apperrors.HandlePgError(err, usersTable, map[string]any{"id": id})
	}

	attrs, err := r.metadata.GetAll(ctx, metadata.EntityTypeUserAttributes, user.ID())
	if err != nil {
		return nil, err
	}

	sessions, err := r.findSessionsByUserID(ctx, user.ID())
	if err != nil {
		return nil, err
	}

	return domain.ReconstructUser(
		user.ID(),
		user.CreatedAt(),
		user.UpdatedAt(),
		user.DeletedAt(),
		user.Phone(),
		user.Email(),
		user.Username(),
		user.Password(),
		user.RoleID(),
		attrs,
		user.IsActive(),
		user.IsApproved(),
		user.LastSeen(),
		sessions,
	), nil
}
```

Apply the same pattern to `FindByPhone` and `FindByEmail` — add `r.metadata.GetAll` call and pass `attrs` to `ReconstructUser`.

Update `scanUser` (lines 463-495): remove `attrsJSON` variable and scan. The scan no longer reads `attributes`:

```go
func scanUser(row pgx.Row) (*domain.User, error) {
	var (
		id         uuid.UUID
		roleID     *uuid.UUID
		username   *string
		email      *string
		phone      string
		pwHash     string
		salt       *string
		active     bool
		isApproved bool
		createdAt  time.Time
		updatedAt  time.Time
		deletedAt  int64
		lastSeen   *time.Time
	)

	err := row.Scan(
		&id, &roleID, &username, &email, &phone,
		&pwHash, &salt,
		&active, &isApproved,
		&createdAt, &updatedAt, &deletedAt, &lastSeen,
	)
	if err != nil {
		return nil, err
	}

	return reconstructUserFromRow(
		id, roleID, username, email, phone, pwHash,
		active, isApproved, createdAt, updatedAt, deletedAt, lastSeen,
	), nil
}
```

Update `scanUserFromRows` similarly — remove `attrsJSON`.

Update `reconstructUserFromRow` — remove `attrsJSON` parameter, pass `nil` for attributes:

```go
func reconstructUserFromRow(
	id uuid.UUID,
	roleID *uuid.UUID,
	username *string,
	emailStr *string,
	phone, pwHash string,
	active, isApproved bool,
	createdAt, updatedAt time.Time,
	deletedAtUnix int64,
	lastSeen *time.Time,
) *domain.User {
	phonVO, _ := domain.NewPhone(phone)
	password := domain.NewPasswordFromHash(pwHash)

	var emailVO *domain.Email
	if emailStr != nil {
		e, err := domain.NewEmail(*emailStr)
		if err == nil {
			emailVO = &e
		}
	}

	var deletedAt *time.Time
	if deletedAtUnix != 0 {
		t := time.Unix(deletedAtUnix, 0)
		deletedAt = &t
	}

	return domain.ReconstructUser(
		id,
		createdAt, updatedAt, deletedAt,
		phonVO,
		emailVO,
		username,
		password,
		roleID,
		nil, // attributes loaded separately via metadata repo
		active, isApproved,
		lastSeen,
		nil,
	)
}
```

- [ ] **Step 6: Update UserReadRepo — remove JSONB scan**

In `internal/user/infrastructure/postgres/read_repo.go`:

Replace `"encoding/json"` with `"gct/internal/shared/infrastructure/metadata"` in imports.

Change `readUserColumns` (lines 20-24) — remove `"attributes"`:
```go
var readUserColumns = []string{
	"id", "role_id", "username", "email", "phone",
	"active", "is_approved",
	"last_seen", "created_at", "updated_at",
}
```

Add `metadata` field to `UserReadRepo` struct and update constructor:
```go
type UserReadRepo struct {
	pool     *pgxpool.Pool
	builder  squirrel.StatementBuilderType
	metadata *metadata.GenericMetadataRepo
}

func NewUserReadRepo(pool *pgxpool.Pool) *UserReadRepo {
	return &UserReadRepo{
		pool:     pool,
		builder:  squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		metadata: metadata.NewGenericMetadataRepo(pool),
	}
}
```

Update `FindByID` — remove `attrsJSON` scan, add metadata lookup:
```go
func (r *UserReadRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.UserView, error) {
	sql, args, err := r.builder.
		Select(readUserColumns...).
		From(consts.TableUsers).
		Where(squirrel.Eq{"id": id}).
		Where(squirrel.Eq{"deleted_at": 0}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	row := r.pool.QueryRow(ctx, sql, args...)

	var (
		uid        uuid.UUID
		roleID     *uuid.UUID
		username   *string
		email      *string
		phone      string
		active     bool
		isApproved bool
		lastSeen   *time.Time
		createdAt  time.Time
		updatedAt  time.Time
	)

	err = row.Scan(
		&uid, &roleID, &username, &email, &phone,
		&active, &isApproved,
		&lastSeen, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableUsers, map[string]any{"id": id})
	}

	attrs, err := r.metadata.GetAll(ctx, metadata.EntityTypeUserAttributes, uid)
	if err != nil {
		return nil, err
	}

	return &domain.UserView{
		ID:         uid,
		Phone:      phone,
		Email:      email,
		Username:   username,
		RoleID:     roleID,
		Attributes: attrs,
		Active:     active,
		IsApproved: isApproved,
	}, nil
}
```

Update `List` — same pattern: remove `attrsJSON` from scan, add per-user metadata lookup.

Update `FindUserForAuth` — remove `"attributes"` from SELECT columns, remove `attrsJSON` scan, add metadata lookup:
```go
func (r *UserReadRepo) FindUserForAuth(ctx context.Context, id uuid.UUID) (*shared.AuthUser, error) {
	sql, args, err := r.builder.
		Select("id", "role_id", "active", "is_approved").
		From(consts.TableUsers).
		Where(squirrel.Eq{"id": id}).
		Where(squirrel.Eq{"deleted_at": 0}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	row := r.pool.QueryRow(ctx, sql, args...)

	var u shared.AuthUser
	err = row.Scan(&u.ID, &u.RoleID, &u.Active, &u.IsApproved)
	if err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableUsers, map[string]any{"id": id})
	}

	attrs, err := r.metadata.GetAll(ctx, metadata.EntityTypeUserAttributes, u.ID)
	if err != nil {
		return nil, err
	}
	u.Attributes = attrs

	return &u, nil
}
```

- [ ] **Step 7: Verify it compiles**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go build ./internal/user/...`
Expected: Clean build.

- [ ] **Step 8: Commit**

```bash
git add internal/user/ internal/shared/domain/auth.go
git commit -m "refactor(user): replace JSONB attributes with entity_metadata EAV"
```

---

### Task 4: Update Authz BC (Policy) — domain + infrastructure

**Files:**
- Modify: `internal/authz/domain/policy.go`
- Modify: `internal/authz/domain/repository.go`
- Modify: `internal/authz/infrastructure/postgres/write_repo.go`
- Modify: `internal/authz/infrastructure/postgres/read_repo.go`
- Modify: `internal/authz/application/command/create_policy.go`
- Modify: `internal/authz/application/command/update_policy.go`

- [ ] **Step 1: Update Policy domain entity**

In `internal/authz/domain/policy.go`:

Change line 29: `conditions map[string]any` → `conditions map[string]string`

Change line 40: `conditions: make(map[string]any)` → `conditions: make(map[string]string)`

Change ReconstructPolicy parameter (line 53): `conditions map[string]any` → `conditions map[string]string`

Change line 56: `conditions = make(map[string]any)` → `conditions = make(map[string]string)`

Change line 81: `func (p *Policy) Conditions() map[string]any` → `func (p *Policy) Conditions() map[string]string`

Change line 104: `func (p *Policy) SetConditions(conditions map[string]any)` → `func (p *Policy) SetConditions(conditions map[string]string)`

Change line 106: `conditions = make(map[string]any)` → `conditions = make(map[string]string)`

- [ ] **Step 2: Update PolicyView in repository**

In `internal/authz/domain/repository.go`, change `PolicyView.Conditions`:
```go
Conditions map[string]string `json:"conditions,omitempty"`
```

- [ ] **Step 3: Update PolicyWriteRepo**

In `internal/authz/infrastructure/postgres/write_repo.go`:

Add `"gct/internal/shared/infrastructure/metadata"` to imports.

Change `policyColumns` (line 33) — remove `"conditions"`:
```go
policyColumns = []string{"id", "permission_id", "effect", "priority", "active", "created_at", "updated_at"}
```

Add `metadata` field to `PolicyWriteRepo`:
```go
type PolicyWriteRepo struct {
	pool     *pgxpool.Pool
	builder  squirrel.StatementBuilderType
	metadata *metadata.GenericMetadataRepo
}

func NewPolicyWriteRepo(pool *pgxpool.Pool) *PolicyWriteRepo {
	return &PolicyWriteRepo{
		pool:     pool,
		builder:  squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		metadata: metadata.NewGenericMetadataRepo(pool),
	}
}
```

Update `Save` — remove `condJSON` marshal, remove `condJSON` from Values, add metadata call:
```go
func (r *PolicyWriteRepo) Save(ctx context.Context, policy *domain.Policy) error {
	sql, args, err := r.builder.
		Insert(policyTable).
		Columns(policyColumns...).
		Values(
			policy.ID(),
			policy.PermissionID(),
			string(policy.Effect()),
			policy.Priority(),
			policy.IsActive(),
			policy.CreatedAt(),
			policy.UpdatedAt(),
		).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildInsert)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, policyTable, nil)
	}

	return r.metadata.SetMany(ctx, metadata.EntityTypePolicyConditions, policy.ID(), policy.Conditions())
}
```

Update `Update` — remove condJSON, remove `Set("conditions", condJSON)`, add metadata call:
```go
func (r *PolicyWriteRepo) Update(ctx context.Context, policy *domain.Policy) error {
	sql, args, err := r.builder.
		Update(policyTable).
		Set("effect", string(policy.Effect())).
		Set("priority", policy.Priority()).
		Set("active", policy.IsActive()).
		Set("updated_at", policy.UpdatedAt()).
		Where(squirrel.Eq{"id": policy.ID()}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildUpdate)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, policyTable, nil)
	}

	return r.metadata.SetMany(ctx, metadata.EntityTypePolicyConditions, policy.ID(), policy.Conditions())
}
```

Update `Delete` — add metadata cleanup:
```go
func (r *PolicyWriteRepo) Delete(ctx context.Context, id uuid.UUID) error {
	sql, args, err := r.builder.
		Delete(policyTable).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildDelete)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, policyTable, nil)
	}

	return r.metadata.DeleteAll(ctx, metadata.EntityTypePolicyConditions, id)
}
```

Update `FindByID` — load conditions from metadata:
```go
func (r *PolicyWriteRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Policy, error) {
	sql, args, err := r.builder.
		Select(policyColumns...).
		From(policyTable).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	row := r.pool.QueryRow(ctx, sql, args...)
	policy, err := scanPolicy(row)
	if err != nil {
		return nil, apperrors.HandlePgError(err, policyTable, map[string]any{"id": id})
	}

	conds, err := r.metadata.GetAll(ctx, metadata.EntityTypePolicyConditions, policy.ID())
	if err != nil {
		return nil, err
	}
	policy.SetConditions(conds)

	return policy, nil
}
```

Update `scanPolicy` — remove `condJSON` from scan:
```go
func scanPolicy(row pgx.Row) (*domain.Policy, error) {
	var (
		id           uuid.UUID
		permissionID uuid.UUID
		effect       string
		priority     int
		active       bool
		ct, ut       interface{}
	)

	err := row.Scan(&id, &permissionID, &effect, &priority, &active, &ct, &ut)
	if err != nil {
		return nil, apperrors.HandlePgError(err, policyTable, nil)
	}

	return domain.ReconstructPolicy(id, toTime(ct), toTime(ut), nil, permissionID, domain.PolicyEffect(effect), priority, active, nil), nil
}
```

Update `scanPolicyFromRows` similarly.

- [ ] **Step 4: Update AuthzReadRepo**

In `internal/authz/infrastructure/postgres/read_repo.go`:

Add metadata composition to `AuthzReadRepo`. In `ListPolicies`, replace the JSONB conditions scan with a metadata lookup per policy. The SELECT should no longer include `"conditions"`:

```go
Select("id", "permission_id", "effect", "priority", "active").
```

After scanning each policy row, call `r.metadata.GetAll(ctx, metadata.EntityTypePolicyConditions, v.ID)` and assign to `v.Conditions`.

- [ ] **Step 5: Update command handlers**

In `internal/authz/application/command/create_policy.go` and `update_policy.go`, change any `Conditions map[string]any` fields to `map[string]string`.

- [ ] **Step 6: Verify it compiles**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go build ./internal/authz/...`
Expected: Clean build.

- [ ] **Step 7: Commit**

```bash
git add internal/authz/
git commit -m "refactor(authz): replace JSONB conditions with entity_metadata EAV"
```

---

### Task 5: Update Audit BC — domain + infrastructure

**Files:**
- Modify: `internal/audit/domain/entity.go`
- Modify: `internal/audit/domain/repository.go`
- Modify: `internal/audit/infrastructure/postgres/write_repo.go`
- Modify: `internal/audit/infrastructure/postgres/read_repo.go`

- [ ] **Step 1: Update AuditLog domain entity**

In `internal/audit/domain/entity.go`:

Change line 57: `metadata map[string]any` → `metadata map[string]string`

Update `NewAuditLog` parameter (line 75): `metadata map[string]any` → `metadata map[string]string`

Change line 77: `metadata = make(map[string]any)` → `metadata = make(map[string]string)`

Update `ReconstructAuditLog` parameter (line 121): `metadata map[string]any` → `metadata map[string]string`

Change line 123: `metadata = make(map[string]any)` → `metadata = make(map[string]string)`

Change line 161: `func (a *AuditLog) Metadata() map[string]any` → `func (a *AuditLog) Metadata() map[string]string`

- [ ] **Step 2: Update AuditLogView**

In `internal/audit/domain/repository.go`, change `AuditLogView.Metadata`:
```go
Metadata map[string]string `json:"metadata,omitempty"`
```

- [ ] **Step 3: Update AuditLogWriteRepo**

In `internal/audit/infrastructure/postgres/write_repo.go`:

Replace `"encoding/json"` with `"gct/internal/shared/infrastructure/metadata"` in imports.

Remove `"metadata"` from `auditLogColumns`:
```go
var auditLogColumns = []string{
	"id", "user_id", "session_id", "action", "resource_type", "resource_id",
	"platform", "ip_address", "user_agent", "permission", "policy_id",
	"decision", "success", "error_message", "created_at",
}
```

Add `metadata` field to `AuditLogWriteRepo`:
```go
type AuditLogWriteRepo struct {
	pool     *pgxpool.Pool
	builder  squirrel.StatementBuilderType
	metadata *metadata.GenericMetadataRepo
}

func NewAuditLogWriteRepo(pool *pgxpool.Pool) *AuditLogWriteRepo {
	return &AuditLogWriteRepo{
		pool:     pool,
		builder:  squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		metadata: metadata.NewGenericMetadataRepo(pool),
	}
}
```

Update `Save` — remove metadataJSON marshal, remove it from Values, add metadata call after insert:
```go
func (r *AuditLogWriteRepo) Save(ctx context.Context, auditLog *domain.AuditLog) error {
	sql, args, err := r.builder.
		Insert(consts.TableAuditLog).
		Columns(auditLogColumns...).
		Values(
			auditLog.ID(),
			auditLog.UserID(),
			auditLog.SessionID(),
			string(auditLog.Action()),
			auditLog.ResourceType(),
			auditLog.ResourceID(),
			auditLog.Platform(),
			auditLog.IPAddress(),
			auditLog.UserAgent(),
			auditLog.Permission(),
			auditLog.PolicyID(),
			auditLog.Decision(),
			auditLog.Success(),
			auditLog.ErrorMessage(),
			auditLog.CreatedAt(),
		).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildInsert)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, consts.TableAuditLog, nil)
	}

	return r.metadata.SetMany(ctx, metadata.EntityTypeAuditLogMetadata, auditLog.ID(), auditLog.Metadata())
}
```

- [ ] **Step 4: Update AuditReadRepo**

In `internal/audit/infrastructure/postgres/read_repo.go`:

Add metadata to `AuditReadRepo`. In `ListAuditLogs`, remove `"metadata"` from the SELECT columns, remove `metadataJSON` scan variable. After scanning each row, call `r.metadata.GetAll` and assign to `view.Metadata`.

- [ ] **Step 5: Verify and commit**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go build ./internal/audit/...`

```bash
git add internal/audit/
git commit -m "refactor(audit): replace JSONB metadata with entity_metadata EAV"
```

---

### Task 6: Update SystemError BC — domain + infrastructure

**Files:**
- Modify: `internal/systemerror/domain/entity.go`
- Modify: `internal/systemerror/domain/repository.go`
- Modify: `internal/systemerror/infrastructure/postgres/write_repo.go`
- Modify: `internal/systemerror/infrastructure/postgres/read_repo.go`

- [ ] **Step 1: Update SystemError domain entity**

In `internal/systemerror/domain/entity.go`, change all `map[string]any` for metadata to `map[string]string`. Update:
- The `metadata` field type
- `NewSystemError` parameter and `make(map[string]any)` → `make(map[string]string)`
- `ReconstructSystemError` parameter
- `SetMetadata` parameter
- `Metadata()` return type

- [ ] **Step 2: Update SystemErrorView**

In `internal/systemerror/domain/repository.go`, change `SystemErrorView.Metadata`:
```go
Metadata map[string]string `json:"metadata,omitempty"`
```

- [ ] **Step 3: Update SystemErrorWriteRepo**

In `internal/systemerror/infrastructure/postgres/write_repo.go`:

Replace `"encoding/json"` with `"gct/internal/shared/infrastructure/metadata"` in imports.

Remove `"metadata"` from `writeColumns`:
```go
var writeColumns = []string{
	"id", "code", "message", "stack_trace",
	"severity", "service_name", "request_id", "user_id",
	"ip_address", "path", "method",
	"is_resolved", "resolved_at", "resolved_by", "created_at",
}
```

Add `metadata` field to struct. Update constructor.

In `Save`: remove `metaJSON` marshal, remove from Values (was at position 5), add `r.metadata.SetMany` after insert.

In `Update`: remove `metaJSON` marshal, remove `Set("metadata", metaJSON)`, add `r.metadata.SetMany` after update.

In `FindByID`: after `scanSystemError`, add `r.metadata.GetAll` call and use `se.SetMetadata(meta)`.

In `List`: after scanning each row, load metadata and set it.

In `scanSystemError` / `scanSystemErrorFromRows` / `reconstructFromRow`: remove `metaJSON` from scan and parameters. Pass `nil` for metadata in `ReconstructSystemError`.

- [ ] **Step 4: Update SystemErrorReadRepo**

Apply same pattern — remove JSONB scan, add metadata lookup per row.

- [ ] **Step 5: Verify and commit**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go build ./internal/systemerror/...`

```bash
git add internal/systemerror/
git commit -m "refactor(systemerror): replace JSONB metadata with entity_metadata EAV"
```

---

### Task 7: Update Integration BC — domain + infrastructure

**Files:**
- Modify: `internal/integration/domain/entity.go`
- Modify: `internal/integration/domain/repository.go`
- Modify: `internal/integration/infrastructure/postgres/write_repo.go`
- Modify: `internal/integration/infrastructure/postgres/read_repo.go`

- [ ] **Step 1: Update Integration domain entity**

In `internal/integration/domain/entity.go`, change `config map[string]any` → `map[string]string`. Update:
- Field type
- Constructor parameter
- `Config()` return type
- `UpdateDetails` parameter

- [ ] **Step 2: Update IntegrationView**

In `internal/integration/domain/repository.go`, change `IntegrationView.Config`:
```go
Config map[string]string `json:"config"`
```

- [ ] **Step 3: Update IntegrationWriteRepo**

In `internal/integration/infrastructure/postgres/write_repo.go`:

Replace `"encoding/json"` with `"gct/internal/shared/infrastructure/metadata"` in imports.

Remove `"config"` from `writeColumns`:
```go
var writeColumns = []string{
	"id", "name", "description", "base_url", "is_active", "created_at", "updated_at",
}
```

Add `metadata` field to struct. Update constructor.

In `Save`: remove configJSON marshal, remove from Values, add `r.metadata.SetMany(ctx, metadata.EntityTypeIntegrationConfig, i.ID(), i.Config())`.

In `Update`: remove configJSON marshal, remove `Set("config", configJSON)`, add metadata SetMany.

In `FindByID`: after scanIntegration, load metadata and reconstruct with it.

In `scanIntegration`: remove `configJSON` from scan, pass `nil` for config in `ReconstructIntegration`.

- [ ] **Step 4: Update IntegrationReadRepo**

Apply same pattern to `internal/integration/infrastructure/postgres/read_repo.go`.

- [ ] **Step 5: Verify and commit**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go build ./internal/integration/...`

```bash
git add internal/integration/
git commit -m "refactor(integration): replace JSONB config with entity_metadata EAV"
```

---

### Task 8: Update DataExport BC — remove unused JSONB

**Files:**
- Modify: `internal/dataexport/infrastructure/postgres/write_repo.go`
- Modify: `internal/dataexport/infrastructure/postgres/read_repo.go` (if it scans filters)

- [ ] **Step 1: Update DataExportWriteRepo**

In `internal/dataexport/infrastructure/postgres/write_repo.go`:

Remove `"filters"` from `writeColumns`:
```go
var writeColumns = []string{
	"id", "type", "status", "file_url",
	"created_by", "created_at", "completed_at",
}
```

In `Save` (line 48): remove the `"{}"` value from Values list. The Values should be:
```go
Values(
	de.ID(),
	de.DataType(),
	de.Status(),
	de.FileURL(),
	de.UserID(),
	de.CreatedAt(),
	nil,
)
```

In `scanDataExport`: remove `filtersJSON` from scan variables, remove `_ = filtersJSON` line. Update Scan call to match new columns.

- [ ] **Step 2: Update DataExportReadRepo if needed**

Check if `internal/dataexport/infrastructure/postgres/read_repo.go` references `filters` and remove it.

- [ ] **Step 3: Verify and commit**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go build ./internal/dataexport/...`

```bash
git add internal/dataexport/
git commit -m "refactor(dataexport): remove unused JSONB filters column"
```

---

### Task 9: Update seed data and fix remaining references

**Files:**
- Modify: `migrations/postgres/20260208000000_seed_authz.sql`
- Modify: `migrations/postgres/20260207000001_fix_missing_columns.sql` (if it adds JSONB columns)

- [ ] **Step 1: Update seed_authz.sql**

In `migrations/postgres/20260208000000_seed_authz.sql`, the policy inserts (lines 266-290) currently use `conditions` JSONB. Since the migration drops the `conditions` column, these inserts need to be split into two parts: insert policy without conditions, then insert conditions into entity_metadata.

Replace lines 264-291 with:

```sql
-- 6. POLICIES (ABAC Examples)
-- Policy: Managers/HR can only manage users if they belong to the same branch.
INSERT INTO policy (permission_id, effect, priority, active)
SELECT p.id, 'ALLOW', 10, true
FROM permission p WHERE p.name IN ('user.update', 'user.delete')
ON CONFLICT DO NOTHING;

-- Insert ABAC conditions for those policies
INSERT INTO entity_metadata (entity_type, entity_id, key, value, value_type)
SELECT 'policy_conditions', pol.id, 'user.relation_names_any', '$target.user.relation_names', 'string'
FROM policy pol
JOIN permission p ON pol.permission_id = p.id
WHERE p.name IN ('user.update', 'user.delete') AND pol.effect = 'ALLOW' AND pol.priority = 10
ON CONFLICT DO NOTHING;

-- Policy: Auditor can see everything regardless of branch
INSERT INTO policy (permission_id, effect, priority, active)
SELECT p.id, 'ALLOW', 100, true
FROM permission p WHERE p.name = 'user.view'
ON CONFLICT DO NOTHING;

INSERT INTO entity_metadata (entity_type, entity_id, key, value, value_type)
SELECT 'policy_conditions', pol.id, 'user.role_name', 'auditor', 'string'
FROM policy pol
JOIN permission p ON pol.permission_id = p.id
WHERE p.name = 'user.view' AND pol.effect = 'ALLOW' AND pol.priority = 100
ON CONFLICT DO NOTHING;

-- Policy: Restrict deletions to specific IP (example)
INSERT INTO policy (permission_id, effect, priority, active)
SELECT p.id, 'DENY', 100, false
FROM permission p WHERE p.name = 'root'
ON CONFLICT DO NOTHING;

INSERT INTO entity_metadata (entity_type, entity_id, key, value, value_type)
SELECT 'policy_conditions', pol.id, 'env.ip_not_in', '["127.0.0.1", "192.168.1.1"]', 'json_array'
FROM policy pol
JOIN permission p ON pol.permission_id = p.id
WHERE p.name = 'root' AND pol.effect = 'DENY' AND pol.priority = 100
ON CONFLICT DO NOTHING;
```

Also update the `-- +goose Down` section — remove the line that references `conditions->>'user.relation_names_any'`:
```sql
DELETE FROM entity_metadata WHERE entity_type = 'policy_conditions';
DELETE FROM policy WHERE id IN (SELECT entity_id FROM entity_metadata WHERE entity_type = 'policy_conditions');
```

- [ ] **Step 2: Remove or update fix_missing_columns migration**

In `migrations/postgres/20260207000001_fix_missing_columns.sql`, if it adds JSONB columns (`conditions`, `attributes`), either remove those ALTER TABLE statements or guard them with comments since the EAV migration handles this.

- [ ] **Step 3: Commit**

```bash
git add migrations/postgres/
git commit -m "refactor: update seed data and migrations for EAV"
```

---

### Task 10: Fix all remaining compile errors and tests

**Files:**
- Various test files and HTTP handlers that reference `map[string]any` for attributes/conditions/metadata/config

- [ ] **Step 1: Find all remaining compile errors**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go build ./...`

Identify all files that fail due to `map[string]any` vs `map[string]string` mismatches or missing `attributes`/`conditions`/`metadata`/`config` columns.

- [ ] **Step 2: Fix HTTP handlers and request DTOs**

Search for request structs that accept `map[string]any` for conditions, attributes, metadata, config and change to `map[string]string`. Files likely include:
- `internal/authz/interfaces/http/handler.go`
- `internal/authz/interfaces/http/request.go`
- `internal/user/interfaces/http/handler.go`
- `internal/user/interfaces/http/request.go`

- [ ] **Step 3: Fix test files**

Search for test files that use `map[string]any` for attributes/conditions/metadata:
- `internal/authz/domain/policy_test.go`
- `internal/authz/domain/event_test.go`
- `internal/authz/application/command/create_policy_test.go`
- `internal/authz/application/command/update_policy_test.go`

Change all `map[string]any{"key": "value"}` to `map[string]string{"key": "value"}`.

- [ ] **Step 4: Fix any callers that create AuditLog/SystemError with metadata**

Search for `NewAuditLog(` and `NewSystemError(` calls that pass `map[string]any` metadata and change to `map[string]string`.

- [ ] **Step 5: Fix any code that reads attributes with type assertion**

Search for code that does `attrs["key"].(string)` or similar type assertions — these are no longer needed since all values are already strings.

- [ ] **Step 6: Verify full build**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go build ./...`
Expected: Clean build, zero errors.

- [ ] **Step 7: Run tests**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go test ./...`
Fix any test failures.

- [ ] **Step 8: Commit**

```bash
git add -A
git commit -m "fix: resolve all compile errors and test failures after JSONB→EAV migration"
```

---

### Task 11: Clean up documentation and backup files

**Files:**
- Modify: `migrations/postgres/MIGRATIONS.md` — update schema docs to remove JSONB references
- Modify: `docs/swagger/` — update swagger annotations if they reference JSONB
- Modify: `internal/shared/infrastructure/errorx/README.md` — update JSONB reference

- [ ] **Step 1: Update MIGRATIONS.md**

Remove all `| column | JSONB |` entries and add `entity_metadata` table documentation.

- [ ] **Step 2: Update swagger docs**

Fix any swagger annotations that reference JSONB columns.

- [ ] **Step 3: Update errorx README**

Change the `| metadata | JSONB |` reference to note it's now in `entity_metadata` table.

- [ ] **Step 4: Commit**

```bash
git add migrations/postgres/MIGRATIONS.md docs/swagger/ internal/shared/infrastructure/errorx/README.md
git commit -m "docs: update documentation after JSONB→EAV migration"
```
