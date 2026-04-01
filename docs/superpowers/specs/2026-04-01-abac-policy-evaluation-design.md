# ABAC Policy Evaluation Design

**Date:** 2026-04-01
**Status:** Approved
**Approach:** Pure Go evaluator as domain service (no external dependencies)

## Problem

The authz BC stores ABAC policies (conditions, effects, priorities) in the database with full CRUD support, but **never evaluates them at runtime**. CheckAccess only performs RBAC (role -> permission -> scope). Policy conditions are dead data.

## Goal

Activate ABAC evaluation so that after RBAC passes, policies bound to matched permissions are evaluated against request context. The system must be extensible — adding new operators, namespaces, or condition keys requires zero structural changes.

## Architecture

### Components

All ABAC logic lives in `internal/authz/domain/` as a domain service:

```
internal/authz/domain/
  abac.go           — EvaluationContext, PolicyEvaluator, operator registry, condition parser
  abac_test.go      — Unit tests for all operators, parsing, evaluation logic
```

### EvaluationContext

A single struct carrying all attributes available for policy evaluation:

```go
type EvaluationContext struct {
    Attrs map[string]map[string]any
}

func (ec EvaluationContext) Resolve(key string) (any, bool)
```

Namespaces within `Attrs`:
- `user` — user ID, role ID, role name, and all user.Attributes from DB
- `env` — client IP, user agent, current time
- `resource` — endpoint-specific data (empty initially, populated per-endpoint later)
- `target` — relation/target entity data (empty initially, populated for relation-based checks later)

`Resolve("user.role_name")` returns `Attrs["user"]["role_name"]`.
`Resolve("$target.user.relation_names")` strips `$target.` prefix and returns `Attrs["target"]["user.relation_names"]`.

### Operator Registry

```go
type OperatorFunc func(fieldVal, conditionVal any) bool

var operators = map[string]OperatorFunc{
    "equals":     opEquals,
    "not_equals": opNotEquals,
    "in":         opIn,
    "not_in":     opNotIn,
    "contains":   opContains,
    "any":        opAny,
    "all":        opAll,
    "gt":         opGt,
    "gte":        opGte,
    "lt":         opLt,
    "lte":        opLte,
    "between":    opBetween,
}
```

Adding a new operator = registering one function. No other code changes needed.

### Operator Definitions

| Operator | Condition Example | Semantics |
|----------|-------------------|-----------|
| `equals` (default) | `"user.role_name": "auditor"` | `field == value` |
| `not_equals` | `"user.role_name_not_equals": "guest"` | `field != value` |
| `in` | `"user.role_name_in": ["admin","mgr"]` | `field in list` |
| `not_in` | `"env.ip_not_in": ["127.0.0.1"]` | `field not in list` |
| `contains` | `"user.tags_contains": "vip"` | `list-field contains value` |
| `any` | `"user.groups_any": ["a","b"]` | `intersection(field, value) not empty` |
| `all` | `"user.perms_all": ["read","write"]` | `value is subset of field` |
| `gt` | `"env.time_gt": "09:00"` | `field > value` |
| `gte` | `"env.time_gte": "09:00"` | `field >= value` |
| `lt` | `"env.time_lt": "18:00"` | `field < value` |
| `lte` | `"env.time_lte": "18:00"` | `field <= value` |
| `between` | `"env.time_between": ["09:00","18:00"]` | `value[0] <= field <= value[1]` |

Type coercion: numeric strings are compared as float64, time strings as string comparison (HH:MM format).

### Condition Key Parsing

Format: `namespace.field[_operator]`

Algorithm: split key by `_` from the right. If the last segment is a registered operator name, use it. Otherwise default to `equals`.

```
"user.role_name"           → namespace=user, field=role_name,       op=equals
"env.ip_not_in"            → namespace=env,  field=ip,              op=not_in
"user.relation_names_any"  → namespace=user, field=relation_names,  op=any
"env.time_between"         → namespace=env,  field=time,            op=between
```

For compound operators like `not_in`, `not_equals`: scan from right checking 2-word combos first, then 1-word.

### Dynamic References

When a condition value is a string starting with `$target.`, resolve it from the EvaluationContext instead of using the literal value.

```json
{"user.relation_names_any": "$target.user.relation_names"}
```

1. Parse key: namespace=user, field=relation_names, op=any
2. Resolve field: `ctx.Resolve("user.relation_names")` -> user's relation names
3. Resolve value: detect `$target.` prefix -> `ctx.Resolve("target.user.relation_names")` -> target's relation names
4. Apply operator: `opAny(userRelNames, targetRelNames)`

### PolicyEvaluator

```go
type PolicyEvaluator struct{}

func (e *PolicyEvaluator) Evaluate(policies []*Policy, ctx EvaluationContext) (effect PolicyEffect, matched bool)
```

Algorithm:
1. Filter: skip policies where `IsActive() == false`
2. Sort: by `Priority()` descending (higher number = higher priority)
3. For each policy:
   a. Evaluate all conditions (AND semantics — all must match)
   b. If all conditions match -> policy matches
4. Collect matched policies
5. Decision:
   - Any matched DENY -> return (DENY, true) — explicit deny always wins
   - Any matched ALLOW (no DENY) -> return (ALLOW, true)
   - No matches -> return ("", false) — RBAC result stands

## Integration Points

### CheckAccess Signature Change

```
Before: CheckAccess(ctx, roleID, path, method) (bool, error)
After:  CheckAccess(ctx, roleID, path, method, evalCtx EvaluationContext) (bool, error)
```

### CheckAccess Internal Flow

```
1. RBAC check: role -> role_permission -> permission_scope -> scope
   - No scope match -> return false (denied by RBAC)
   - Scope match -> collect matched permission IDs

2. ABAC check: fetch policies for matched permission IDs
   - No policies exist -> return true (RBAC sufficient)
   - Policies exist -> PolicyEvaluator.Evaluate(policies, evalCtx)
     - matched=false -> return true (no applicable policy, RBAC stands)
     - effect=ALLOW  -> return true
     - effect=DENY   -> return false
```

The RBAC query already joins role_permission, so matched permission IDs are available. A second query fetches policies by those permission IDs using the existing `FindByPermissionID` repository method (already implemented but unused).

### Middleware Changes

`AuthzMiddleware.Authz` builds `EvaluationContext` before calling CheckAccess:

```go
evalCtx := domain.EvaluationContext{
    Attrs: map[string]map[string]any{
        "user": mergeUserAttrs(user),
        "env":  buildEnvAttrs(ctx),
        "resource": {},
        "target":   {},
    },
}
```

Where:
- `mergeUserAttrs`: combines `AuthUser.Attributes` with `id`, `role_id` fields
- `buildEnvAttrs`: `ip` from `ctx.ClientIP()`, `user_agent` from header, `time` as `time.Now().Format("15:04")`

### Affected Signatures (cascading change)

1. `AuthzReadRepository.CheckAccess` — add `evalCtx` param
2. `CheckAccessQuery` — add `EvalCtx` field
3. `CheckAccessHandler.Handle` — pass through
4. `AuthzMiddleware.Authz` — build context, pass to handler

### Role Name in CheckAccess

The current CheckAccess query already fetches `r.name` for the super_admin bypass. This role name should be injected into `evalCtx.Attrs["user"]["role_name"]` inside the repository method before policy evaluation, so that `user.role_name` conditions work without the middleware needing to know the role name.

## Extensibility

| Future Need | What Changes | Scope |
|-------------|-------------|-------|
| New operator (e.g. `regex`) | Add one `OperatorFunc` to registry | 1 function |
| New namespace (e.g. `org`) | Populate `evalCtx.Attrs["org"]` in middleware | Middleware only |
| New condition key | Nothing — parsing is automatic | 0 changes |
| Resource-level ABAC | Populate `evalCtx.Attrs["resource"]` per-handler | Handler/middleware |
| Target/relation ABAC | Populate `evalCtx.Attrs["target"]` per-handler | Handler/middleware |
| OR logic between conditions | Add `"$or": [cond1, cond2]` support in evaluateConditions | Evaluator only |

## Testing Strategy

### Unit Tests (abac_test.go)

1. **Operator tests** — each operator with various types (string, number, list, nil)
2. **Key parsing tests** — all operator suffixes, default equals, compound operators
3. **Dynamic reference tests** — `$target.*` resolution
4. **EvaluationContext.Resolve tests** — valid keys, missing keys, nested namespaces
5. **PolicyEvaluator.Evaluate tests:**
   - Single ALLOW policy matches -> ALLOW
   - Single DENY policy matches -> DENY
   - DENY wins over ALLOW regardless of priority
   - Inactive policies skipped
   - No matching policies -> matched=false
   - Multiple conditions AND semantics (partial match = no match)
   - Priority ordering (higher priority evaluated first)
   - Empty conditions -> policy always matches
   - Dynamic reference in condition value

### Integration Tests (test/integration/authz/access/)

- Existing RBAC tests continue to pass (backward compatible when no policies exist)
- RBAC pass + DENY policy -> denied
- RBAC pass + ALLOW policy -> allowed
- RBAC pass + no policies -> allowed (existing behavior)
- RBAC fail -> denied (policies never consulted)

## Non-Goals

- OR logic between conditions (can be added later without structural changes)
- Nested condition groups (AND/OR trees) — flat condition map with AND semantics is sufficient
- Policy caching — correctness first, optimize later
- Admin UI for policy evaluation testing
