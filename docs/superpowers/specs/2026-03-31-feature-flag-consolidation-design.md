# Feature Flag Consolidation — PostgreSQL-Only Evaluation

**Date:** 2026-03-31
**Status:** Approved

## Problem

The project has two disconnected feature flag systems:
1. **PostgreSQL CRUD** (`internal/featureflag/`) — admin panel for managing flags, but no runtime evaluation
2. **go-feature-flag + YAML/Redis** (`internal/shared/infrastructure/featureflag/`) — runtime evaluation, but disconnected from DB

Admin panel changes never reach runtime. This spec consolidates both into a single PostgreSQL-backed system with built-in evaluation.

## Decision

Remove `go-feature-flag` library, YAML config, and Redis retriever. Build an in-house evaluation engine that reads directly from PostgreSQL with in-memory caching.

## Database Schema

### `feature_flags` (existing — modified)

| Column | Type | Description |
|--------|------|-------------|
| id | UUID PK | DEFAULT gen_random_uuid() |
| key | TEXT UNIQUE NOT NULL | Runtime lookup key |
| name | TEXT NOT NULL | Display name |
| description | TEXT NOT NULL DEFAULT '' | |
| flag_type | TEXT NOT NULL | CHECK (type IN ('bool', 'string', 'int', 'float')) |
| default_value | TEXT NOT NULL DEFAULT '' | Value when no rule matches |
| rollout_percentage | INT NOT NULL DEFAULT 0 | 0-100, gradual rollout for default |
| is_active | BOOL NOT NULL DEFAULT true | Kill-switch: false = always off |
| created_at | TIMESTAMPTZ NOT NULL DEFAULT NOW() | |
| updated_at | TIMESTAMPTZ NOT NULL DEFAULT NOW() | |
| deleted_at | TIMESTAMPTZ | Soft delete |

### `feature_flag_rule_groups` (new)

| Column | Type | Description |
|--------|------|-------------|
| id | UUID PK | DEFAULT gen_random_uuid() |
| flag_id | UUID NOT NULL FK | References feature_flags(id) ON DELETE CASCADE |
| name | TEXT NOT NULL | Human-readable, e.g. "Enable for UZ admins" |
| variation | TEXT NOT NULL | Value returned when all conditions match |
| priority | INT NOT NULL | Lower = checked first |
| created_at | TIMESTAMPTZ NOT NULL DEFAULT NOW() | |
| updated_at | TIMESTAMPTZ NOT NULL DEFAULT NOW() | |

Indexes: `(flag_id, priority)` for ordered retrieval.

### `feature_flag_conditions` (new)

| Column | Type | Description |
|--------|------|-------------|
| id | UUID PK | DEFAULT gen_random_uuid() |
| rule_group_id | UUID NOT NULL FK | References feature_flag_rule_groups(id) ON DELETE CASCADE |
| attribute | TEXT NOT NULL | User attribute key: "role", "country", "email", "plan", etc. |
| operator | TEXT NOT NULL | CHECK (operator IN ('eq','not_eq','in','not_in','gt','gte','lt','lte','contains')) |
| value | TEXT NOT NULL | Target value. For 'in'/'not_in': comma-separated list |
| created_at | TIMESTAMPTZ NOT NULL DEFAULT NOW() | |

Index: `(rule_group_id)` for loading conditions with their group.

## Domain Model

```
FeatureFlag (aggregate root)
├── key, name, description
├── flagType: "bool" | "string" | "int" | "float"
├── defaultValue: string
├── rolloutPercentage: int (0-100)
├── isActive: bool
├── ruleGroups: []RuleGroup (sorted by priority)

RuleGroup (entity)
├── name: string
├── variation: string
├── priority: int
├── conditions: []Condition

Condition (value object)
├── attribute: string
├── operator: string
├── value: string
```

## Evaluation Logic

`FeatureFlag.Evaluate(userAttrs map[string]string) string`:

1. If `isActive == false` → return `defaultValue`
2. Iterate rule groups by priority (ascending):
   - For each group, check ALL conditions (AND logic)
   - If all conditions match → return `group.variation`
3. No rule matched → check `rolloutPercentage`:
   - `hash(userAttrs["user_id"] + flag.key) % 100 < rolloutPercentage` → return flag_type's "on" value
   - Otherwise → return `defaultValue`

### Operator Matching

`Condition.Match(userValue string) bool`:

| Operator | Logic |
|----------|-------|
| eq | `userValue == value` |
| not_eq | `userValue != value` |
| in | `userValue` is in comma-separated `value` list |
| not_in | `userValue` is NOT in comma-separated `value` list |
| gt | `toFloat(userValue) > toFloat(value)` |
| gte | `toFloat(userValue) >= toFloat(value)` |
| lt | `toFloat(userValue) < toFloat(value)` |
| lte | `toFloat(userValue) <= toFloat(value)` |
| contains | `strings.Contains(userValue, value)` |

## Runtime Evaluator Interface

Located in `internal/shared/infrastructure/featureflag/` (replaces current code):

```go
type Evaluator interface {
    IsEnabled(ctx context.Context, flagKey string, userAttrs map[string]string) bool
    GetString(ctx context.Context, flagKey string, userAttrs map[string]string) string
    GetInt(ctx context.Context, flagKey string, userAttrs map[string]string) int
    GetFloat(ctx context.Context, flagKey string, userAttrs map[string]string) float64
}
```

## Caching

- In-memory cache using `sync.Map`
- All flags loaded at app startup
- Cache invalidated via existing eventBus when flags are created/updated/deleted
- On cache miss: read from DB, populate cache
- No external dependencies (no Redis needed)

## Removed Components

| What | Path |
|------|------|
| go-feature-flag dependency | go.mod |
| YAML config file | config/flags.yaml |
| FeatureFlag config struct | config/featureflag.go |
| Shared FF infrastructure | internal/shared/infrastructure/featureflag/ (entire directory) |

## Modified Components

| What | Change |
|------|--------|
| internal/featureflag/ | Full rewrite: new entity with rule groups/conditions, new repos, evaluation logic, updated handlers |
| Migration | New migration for schema changes (add rule_groups, conditions tables; alter feature_flags) |

## Admin HTTP Endpoints

| Method | Path | Description |
|--------|------|-------------|
| POST | /feature-flags | Create flag (with rule groups and conditions) |
| GET | /feature-flags | List flags (pagination, filter) |
| GET | /feature-flags/:id | Get single flag (with rule groups and conditions) |
| PUT | /feature-flags/:id | Update flag |
| DELETE | /feature-flags/:id | Delete flag (cascades to rules) |
| POST | /feature-flags/:id/rule-groups | Add rule group |
| PUT | /feature-flags/:id/rule-groups/:groupId | Update rule group |
| DELETE | /feature-flags/:id/rule-groups/:groupId | Delete rule group (cascades to conditions) |

## Testing

- Domain: unit tests for Evaluate, Condition.Match (all operators), RuleGroup.MatchAll
- Repos: integration tests with real PostgreSQL
- Cache: unit test for invalidation via eventBus
- Handlers: unit tests with mocked BC
