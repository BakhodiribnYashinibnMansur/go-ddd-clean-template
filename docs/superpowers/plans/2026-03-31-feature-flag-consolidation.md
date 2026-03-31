# Feature Flag Consolidation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the dual feature flag system (PostgreSQL CRUD + go-feature-flag YAML) with a single PostgreSQL-backed system that handles both management and runtime evaluation.

**Architecture:** The `internal/featureflag/` bounded context gets a full rewrite: new domain model with RuleGroup/Condition entities, PostgreSQL-backed evaluation engine, and in-memory cache with eventBus invalidation. The `internal/shared/infrastructure/featureflag/` package (go-feature-flag wrapper) is deleted entirely.

**Tech Stack:** Go, PostgreSQL, pgx/v5, squirrel, gin, sync.Map (cache), fnv hash (rollout)

**Spec:** `docs/superpowers/specs/2026-03-31-feature-flag-consolidation-design.md`

---

## File Structure

### New files
- `internal/featureflag/domain/condition.go` — Condition value object with operator matching
- `internal/featureflag/domain/rule_group.go` — RuleGroup entity with AND condition matching
- `internal/featureflag/domain/operator.go` — Operator constants and validation
- `internal/featureflag/domain/condition_test.go` — Unit tests for all operators
- `internal/featureflag/domain/rule_group_test.go` — Unit tests for AND matching
- `internal/featureflag/infrastructure/postgres/rule_group_repo.go` — RuleGroup write repo
- `internal/featureflag/infrastructure/postgres/rule_group_read_repo.go` — RuleGroup read repo
- `internal/featureflag/application/command/create_rule_group.go` — CreateRuleGroup command handler
- `internal/featureflag/application/command/update_rule_group.go` — UpdateRuleGroup command handler
- `internal/featureflag/application/command/delete_rule_group.go` — DeleteRuleGroup command handler
- `internal/featureflag/infrastructure/postgres/evaluator.go` — PostgreSQL-backed Evaluator implementation
- `internal/featureflag/infrastructure/cache/evaluator_cache.go` — In-memory cached Evaluator wrapper
- `migrations/postgres/20260331000000_feature_flag_consolidation.sql` — Schema migration

### Modified files
- `internal/featureflag/domain/entity.go` — Add key, flagType, defaultValue, rolloutPercentage, ruleGroups; add Evaluate method
- `internal/featureflag/domain/entity_test.go` — Tests for Evaluate with rules and rollout
- `internal/featureflag/domain/repository.go` — Add RuleGroup repos, Evaluator interface, update views
- `internal/featureflag/domain/event.go` — Add FlagCreated, FlagUpdated, FlagDeleted events
- `internal/featureflag/domain/error.go` — Add rule group errors
- `internal/featureflag/infrastructure/postgres/write_repo.go` — Persist new fields (key, flagType, defaultValue, rolloutPercentage)
- `internal/featureflag/infrastructure/postgres/read_repo.go` — Read new fields, join rule groups and conditions
- `internal/featureflag/application/dto.go` — Add RuleGroupView, ConditionView; update FeatureFlagView
- `internal/featureflag/application/command/create.go` — Add key, flagType, defaultValue fields to CreateCommand
- `internal/featureflag/application/command/update.go` — Add new updatable fields
- `internal/featureflag/interfaces/http/handler.go` — Add rule group endpoints
- `internal/featureflag/interfaces/http/request.go` — Add rule group request types
- `internal/featureflag/interfaces/http/routes.go` — Register rule group routes
- `internal/featureflag/bc.go` — Wire new handlers, evaluator, cache
- `internal/shared/domain/consts/tables.go` — Add TableFeatureFlagRuleGroups, TableFeatureFlagConditions
- `internal/app/ddd_bootstrap.go` — Pass evaluator to BC
- `internal/app/ddd_routes.go` — Register rule group routes
- `test/integration/featureflag/setup_test.go` — Clean new tables
- `test/integration/featureflag/integration_test.go` — Add rule group and evaluation tests

### Deleted files
- `internal/shared/infrastructure/featureflag/featureflag.go`
- `internal/shared/infrastructure/featureflag/featureflag_test.go`
- `internal/shared/infrastructure/featureflag/middleware.go`
- `internal/shared/infrastructure/featureflag/redis_retriever.go`
- `internal/shared/infrastructure/featureflag/user.go`
- `internal/shared/infrastructure/featureflag/README.md`
- `internal/shared/infrastructure/featureflag/FEATURE_FLAGS.md`
- `internal/shared/infrastructure/featureflag/FEATURE_FLAGS_SUMMARY.md`
- `internal/shared/infrastructure/featureflag/FEATURE_FLAGS_INTEGRATION.md`
- `internal/shared/infrastructure/featureflag/FEATURE_FLAGS_ARCHITECTURE.md`
- `config/flags.yaml`
- `config/featureflag.go`

---

### Task 1: Database Migration

**Files:**
- Create: `migrations/postgres/20260331000000_feature_flag_consolidation.sql`
- Modify: `internal/shared/domain/consts/tables.go`

- [ ] **Step 1: Create the migration file**

```sql
-- +goose Up

-- Add new columns to feature_flags
ALTER TABLE feature_flags ADD COLUMN IF NOT EXISTS rollout_percentage INT NOT NULL DEFAULT 0;
ALTER TABLE feature_flags ADD COLUMN IF NOT EXISTS default_value TEXT NOT NULL DEFAULT '';

-- Rename 'type' to 'flag_type' for Go keyword safety
ALTER TABLE feature_flags RENAME COLUMN type TO flag_type;

-- Create rule groups table
CREATE TABLE IF NOT EXISTS feature_flag_rule_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    flag_id UUID NOT NULL REFERENCES feature_flags(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    variation TEXT NOT NULL,
    priority INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_ff_rule_groups_flag_priority ON feature_flag_rule_groups(flag_id, priority);

-- Create conditions table
CREATE TABLE IF NOT EXISTS feature_flag_conditions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_group_id UUID NOT NULL REFERENCES feature_flag_rule_groups(id) ON DELETE CASCADE,
    attribute TEXT NOT NULL,
    operator TEXT NOT NULL CHECK (operator IN ('eq','not_eq','in','not_in','gt','gte','lt','lte','contains')),
    value TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_ff_conditions_rule_group ON feature_flag_conditions(rule_group_id);

-- +goose Down
DROP TABLE IF EXISTS feature_flag_conditions;
DROP TABLE IF EXISTS feature_flag_rule_groups;
ALTER TABLE feature_flags RENAME COLUMN flag_type TO type;
ALTER TABLE feature_flags DROP COLUMN IF EXISTS rollout_percentage;
ALTER TABLE feature_flags DROP COLUMN IF EXISTS default_value;
```

- [ ] **Step 2: Add table constants**

In `internal/shared/domain/consts/tables.go`, add after `TableFeatureFlags`:

```go
TableFeatureFlagRuleGroups  = "feature_flag_rule_groups"
TableFeatureFlagConditions  = "feature_flag_conditions"
```

- [ ] **Step 3: Commit**

```bash
git add migrations/postgres/20260331000000_feature_flag_consolidation.sql internal/shared/domain/consts/tables.go
git commit -m "feat(featureflag): add migration for rule groups and conditions tables"
```

---

### Task 2: Condition Value Object

**Files:**
- Create: `internal/featureflag/domain/operator.go`
- Create: `internal/featureflag/domain/condition.go`
- Create: `internal/featureflag/domain/condition_test.go`

- [ ] **Step 1: Write operator constants**

File: `internal/featureflag/domain/operator.go`

```go
package domain

// Operator constants for condition evaluation.
const (
	OpEq       = "eq"
	OpNotEq    = "not_eq"
	OpIn       = "in"
	OpNotIn    = "not_in"
	OpGt       = "gt"
	OpGte      = "gte"
	OpLt       = "lt"
	OpLte      = "lte"
	OpContains = "contains"
)

// ValidOperators is the set of all valid operator strings.
var ValidOperators = map[string]bool{
	OpEq: true, OpNotEq: true,
	OpIn: true, OpNotIn: true,
	OpGt: true, OpGte: true,
	OpLt: true, OpLte: true,
	OpContains: true,
}

// IsValidOperator checks if the given operator string is recognized.
func IsValidOperator(op string) bool {
	return ValidOperators[op]
}
```

- [ ] **Step 2: Write failing tests for Condition**

File: `internal/featureflag/domain/condition_test.go`

```go
package domain_test

import (
	"testing"

	"gct/internal/featureflag/domain"
)

func TestCondition_Match_Eq(t *testing.T) {
	c := domain.NewCondition("role", domain.OpEq, "admin")
	if !c.Match("admin") {
		t.Error("expected match for eq admin")
	}
	if c.Match("user") {
		t.Error("expected no match for eq user")
	}
}

func TestCondition_Match_NotEq(t *testing.T) {
	c := domain.NewCondition("role", domain.OpNotEq, "admin")
	if !c.Match("user") {
		t.Error("expected match for not_eq user")
	}
	if c.Match("admin") {
		t.Error("expected no match for not_eq admin")
	}
}

func TestCondition_Match_In(t *testing.T) {
	c := domain.NewCondition("country", domain.OpIn, "US,UK,UZ")
	if !c.Match("UZ") {
		t.Error("expected match for UZ in US,UK,UZ")
	}
	if c.Match("RU") {
		t.Error("expected no match for RU")
	}
}

func TestCondition_Match_NotIn(t *testing.T) {
	c := domain.NewCondition("country", domain.OpNotIn, "US,UK")
	if !c.Match("UZ") {
		t.Error("expected match for UZ not_in US,UK")
	}
	if c.Match("US") {
		t.Error("expected no match for US")
	}
}

func TestCondition_Match_Gt(t *testing.T) {
	c := domain.NewCondition("age", domain.OpGt, "18")
	if !c.Match("25") {
		t.Error("expected match for 25 > 18")
	}
	if c.Match("18") {
		t.Error("expected no match for 18 > 18")
	}
	if c.Match("10") {
		t.Error("expected no match for 10 > 18")
	}
}

func TestCondition_Match_Gte(t *testing.T) {
	c := domain.NewCondition("age", domain.OpGte, "18")
	if !c.Match("18") {
		t.Error("expected match for 18 >= 18")
	}
	if c.Match("17") {
		t.Error("expected no match for 17 >= 18")
	}
}

func TestCondition_Match_Lt(t *testing.T) {
	c := domain.NewCondition("age", domain.OpLt, "18")
	if !c.Match("10") {
		t.Error("expected match for 10 < 18")
	}
	if c.Match("18") {
		t.Error("expected no match for 18 < 18")
	}
}

func TestCondition_Match_Lte(t *testing.T) {
	c := domain.NewCondition("age", domain.OpLte, "18")
	if !c.Match("18") {
		t.Error("expected match for 18 <= 18")
	}
	if c.Match("19") {
		t.Error("expected no match for 19 <= 18")
	}
}

func TestCondition_Match_Contains(t *testing.T) {
	c := domain.NewCondition("email", domain.OpContains, "@example.com")
	if !c.Match("user@example.com") {
		t.Error("expected match for contains @example.com")
	}
	if c.Match("user@other.com") {
		t.Error("expected no match for user@other.com")
	}
}

func TestCondition_Match_InvalidOperator(t *testing.T) {
	c := domain.NewCondition("x", "invalid", "y")
	if c.Match("y") {
		t.Error("expected no match for invalid operator")
	}
}

func TestCondition_Match_NonNumeric_Gt(t *testing.T) {
	c := domain.NewCondition("age", domain.OpGt, "18")
	if c.Match("abc") {
		t.Error("expected no match for non-numeric value with gt")
	}
}
```

- [ ] **Step 3: Run tests to verify they fail**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go test ./internal/featureflag/domain/ -run TestCondition -v`
Expected: FAIL — `NewCondition` undefined

- [ ] **Step 4: Implement Condition**

File: `internal/featureflag/domain/condition.go`

```go
package domain

import (
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// Condition is a value object representing a single targeting condition.
// Multiple conditions within a RuleGroup are combined with AND logic.
type Condition struct {
	id          uuid.UUID
	ruleGroupID uuid.UUID
	attribute   string
	operator    string
	value       string
}

// NewCondition creates a new Condition value object.
func NewCondition(attribute, operator, value string) Condition {
	return Condition{
		id:        uuid.New(),
		attribute: attribute,
		operator:  operator,
		value:     value,
	}
}

// ReconstructCondition rebuilds a Condition from persisted data.
func ReconstructCondition(id, ruleGroupID uuid.UUID, attribute, operator, value string) Condition {
	return Condition{
		id:          id,
		ruleGroupID: ruleGroupID,
		attribute:   attribute,
		operator:    operator,
		value:       value,
	}
}

// Match evaluates this condition against a user attribute value.
func (c Condition) Match(userValue string) bool {
	switch c.operator {
	case OpEq:
		return userValue == c.value
	case OpNotEq:
		return userValue != c.value
	case OpIn:
		return containsInList(c.value, userValue)
	case OpNotIn:
		return !containsInList(c.value, userValue)
	case OpGt:
		return compareFloat(userValue, c.value, func(a, b float64) bool { return a > b })
	case OpGte:
		return compareFloat(userValue, c.value, func(a, b float64) bool { return a >= b })
	case OpLt:
		return compareFloat(userValue, c.value, func(a, b float64) bool { return a < b })
	case OpLte:
		return compareFloat(userValue, c.value, func(a, b float64) bool { return a <= b })
	case OpContains:
		return strings.Contains(userValue, c.value)
	default:
		return false
	}
}

// Getters
func (c Condition) ID() uuid.UUID          { return c.id }
func (c Condition) RuleGroupID() uuid.UUID  { return c.ruleGroupID }
func (c Condition) Attribute() string       { return c.attribute }
func (c Condition) Operator() string        { return c.operator }
func (c Condition) Value() string           { return c.value }

func containsInList(commaSeparated, target string) bool {
	for _, item := range strings.Split(commaSeparated, ",") {
		if strings.TrimSpace(item) == target {
			return true
		}
	}
	return false
}

func compareFloat(userVal, condVal string, cmp func(a, b float64) bool) bool {
	a, err := strconv.ParseFloat(userVal, 64)
	if err != nil {
		return false
	}
	b, err := strconv.ParseFloat(condVal, 64)
	if err != nil {
		return false
	}
	return cmp(a, b)
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go test ./internal/featureflag/domain/ -run TestCondition -v`
Expected: PASS — all 11 tests green

- [ ] **Step 6: Commit**

```bash
git add internal/featureflag/domain/operator.go internal/featureflag/domain/condition.go internal/featureflag/domain/condition_test.go
git commit -m "feat(featureflag): add Condition value object with operator matching"
```

---

### Task 3: RuleGroup Entity

**Files:**
- Create: `internal/featureflag/domain/rule_group.go`
- Create: `internal/featureflag/domain/rule_group_test.go`

- [ ] **Step 1: Write failing tests for RuleGroup**

File: `internal/featureflag/domain/rule_group_test.go`

```go
package domain_test

import (
	"testing"

	"gct/internal/featureflag/domain"

	"github.com/google/uuid"
)

func TestRuleGroup_MatchAll_AllConditionsTrue(t *testing.T) {
	rg := domain.NewRuleGroup(uuid.New(), "UZ admins", "true", 1)
	rg.AddCondition(domain.NewCondition("role", domain.OpEq, "admin"))
	rg.AddCondition(domain.NewCondition("country", domain.OpEq, "UZ"))

	attrs := map[string]string{"role": "admin", "country": "UZ"}
	if !rg.MatchAll(attrs) {
		t.Error("expected all conditions to match")
	}
}

func TestRuleGroup_MatchAll_OneConditionFalse(t *testing.T) {
	rg := domain.NewRuleGroup(uuid.New(), "UZ admins", "true", 1)
	rg.AddCondition(domain.NewCondition("role", domain.OpEq, "admin"))
	rg.AddCondition(domain.NewCondition("country", domain.OpEq, "UZ"))

	attrs := map[string]string{"role": "admin", "country": "US"}
	if rg.MatchAll(attrs) {
		t.Error("expected no match when one condition fails")
	}
}

func TestRuleGroup_MatchAll_MissingAttribute(t *testing.T) {
	rg := domain.NewRuleGroup(uuid.New(), "test", "true", 1)
	rg.AddCondition(domain.NewCondition("role", domain.OpEq, "admin"))

	attrs := map[string]string{"country": "UZ"}
	if rg.MatchAll(attrs) {
		t.Error("expected no match when attribute missing from user attrs")
	}
}

func TestRuleGroup_MatchAll_NoConditions(t *testing.T) {
	rg := domain.NewRuleGroup(uuid.New(), "empty", "true", 1)

	attrs := map[string]string{"role": "admin"}
	if rg.MatchAll(attrs) {
		t.Error("expected no match when rule group has no conditions")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go test ./internal/featureflag/domain/ -run TestRuleGroup -v`
Expected: FAIL — `NewRuleGroup` undefined

- [ ] **Step 3: Implement RuleGroup**

File: `internal/featureflag/domain/rule_group.go`

```go
package domain

import (
	"time"

	"github.com/google/uuid"
)

// RuleGroup is an entity holding a set of conditions combined with AND logic.
// If all conditions match, the group's variation is returned.
type RuleGroup struct {
	id         uuid.UUID
	flagID     uuid.UUID
	name       string
	variation  string
	priority   int
	conditions []Condition
	createdAt  time.Time
	updatedAt  time.Time
}

// NewRuleGroup creates a new RuleGroup entity.
func NewRuleGroup(flagID uuid.UUID, name, variation string, priority int) *RuleGroup {
	now := time.Now()
	return &RuleGroup{
		id:        uuid.New(),
		flagID:    flagID,
		name:      name,
		variation: variation,
		priority:  priority,
		createdAt: now,
		updatedAt: now,
	}
}

// ReconstructRuleGroup rebuilds a RuleGroup from persisted data.
func ReconstructRuleGroup(id, flagID uuid.UUID, name, variation string, priority int, createdAt, updatedAt time.Time, conditions []Condition) *RuleGroup {
	return &RuleGroup{
		id:         id,
		flagID:     flagID,
		name:       name,
		variation:  variation,
		priority:   priority,
		conditions: conditions,
		createdAt:  createdAt,
		updatedAt:  updatedAt,
	}
}

// AddCondition adds a condition to this rule group.
func (rg *RuleGroup) AddCondition(c Condition) {
	rg.conditions = append(rg.conditions, c)
}

// MatchAll returns true if ALL conditions match the given user attributes.
// Returns false if there are no conditions (empty rule group never matches).
func (rg *RuleGroup) MatchAll(userAttrs map[string]string) bool {
	if len(rg.conditions) == 0 {
		return false
	}
	for _, c := range rg.conditions {
		userVal, exists := userAttrs[c.Attribute()]
		if !exists {
			return false
		}
		if !c.Match(userVal) {
			return false
		}
	}
	return true
}

// UpdateDetails applies partial modifications to the rule group.
func (rg *RuleGroup) UpdateDetails(name, variation *string, priority *int) {
	if name != nil {
		rg.name = *name
	}
	if variation != nil {
		rg.variation = *variation
	}
	if priority != nil {
		rg.priority = *priority
	}
	rg.updatedAt = time.Now()
}

// Getters
func (rg *RuleGroup) ID() uuid.UUID          { return rg.id }
func (rg *RuleGroup) FlagID() uuid.UUID       { return rg.flagID }
func (rg *RuleGroup) Name() string            { return rg.name }
func (rg *RuleGroup) Variation() string        { return rg.variation }
func (rg *RuleGroup) Priority() int            { return rg.priority }
func (rg *RuleGroup) Conditions() []Condition  { return rg.conditions }
func (rg *RuleGroup) CreatedAt() time.Time     { return rg.createdAt }
func (rg *RuleGroup) UpdatedAt() time.Time     { return rg.updatedAt }
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go test ./internal/featureflag/domain/ -run TestRuleGroup -v`
Expected: PASS — all 4 tests green

- [ ] **Step 5: Commit**

```bash
git add internal/featureflag/domain/rule_group.go internal/featureflag/domain/rule_group_test.go
git commit -m "feat(featureflag): add RuleGroup entity with AND condition matching"
```

---

### Task 4: Rewrite FeatureFlag Entity with Evaluate

**Files:**
- Modify: `internal/featureflag/domain/entity.go`
- Modify: `internal/featureflag/domain/entity_test.go`

- [ ] **Step 1: Write failing tests for Evaluate**

Replace contents of `internal/featureflag/domain/entity_test.go`:

```go
package domain_test

import (
	"testing"

	"gct/internal/featureflag/domain"

	"github.com/google/uuid"
)

func TestNewFeatureFlag(t *testing.T) {
	ff := domain.NewFeatureFlag("dark-mode", "dark_mode", "Enable dark mode", "bool", "false", 0)

	if ff.Name() != "dark-mode" {
		t.Fatalf("expected name dark-mode, got %s", ff.Name())
	}
	if ff.Key() != "dark_mode" {
		t.Fatalf("expected key dark_mode, got %s", ff.Key())
	}
	if ff.FlagType() != "bool" {
		t.Fatalf("expected type bool, got %s", ff.FlagType())
	}
	if ff.DefaultValue() != "false" {
		t.Fatalf("expected default false, got %s", ff.DefaultValue())
	}
	if ff.IsActive() {
		t.Fatal("expected is_active false by default")
	}
}

func TestFeatureFlag_Evaluate_InactiveReturnsDefault(t *testing.T) {
	ff := domain.NewFeatureFlag("test", "test_key", "desc", "bool", "false", 50)
	// isActive defaults to false in NewFeatureFlag, call Activate to control
	result := ff.Evaluate(map[string]string{"user_id": "u1"})
	if result != "false" {
		t.Fatalf("expected false for inactive flag, got %s", result)
	}
}

func TestFeatureFlag_Evaluate_RuleGroupMatch(t *testing.T) {
	ff := domain.NewFeatureFlag("test", "test_key", "desc", "bool", "false", 0)
	ff.Activate()

	rg := domain.NewRuleGroup(ff.ID(), "admins", "true", 1)
	rg.AddCondition(domain.NewCondition("role", domain.OpEq, "admin"))
	ff.AddRuleGroup(rg)

	result := ff.Evaluate(map[string]string{"user_id": "u1", "role": "admin"})
	if result != "true" {
		t.Fatalf("expected true for admin, got %s", result)
	}
}

func TestFeatureFlag_Evaluate_RuleGroupNoMatch_FallsToDefault(t *testing.T) {
	ff := domain.NewFeatureFlag("test", "test_key", "desc", "bool", "false", 0)
	ff.Activate()

	rg := domain.NewRuleGroup(ff.ID(), "admins", "true", 1)
	rg.AddCondition(domain.NewCondition("role", domain.OpEq, "admin"))
	ff.AddRuleGroup(rg)

	result := ff.Evaluate(map[string]string{"user_id": "u1", "role": "user"})
	if result != "false" {
		t.Fatalf("expected false for non-admin, got %s", result)
	}
}

func TestFeatureFlag_Evaluate_PriorityOrder(t *testing.T) {
	ff := domain.NewFeatureFlag("test", "test_key", "desc", "string", "default", 0)
	ff.Activate()

	rg1 := domain.NewRuleGroup(ff.ID(), "low priority", "variant-b", 10)
	rg1.AddCondition(domain.NewCondition("role", domain.OpEq, "admin"))
	ff.AddRuleGroup(rg1)

	rg2 := domain.NewRuleGroup(ff.ID(), "high priority", "variant-a", 1)
	rg2.AddCondition(domain.NewCondition("role", domain.OpEq, "admin"))
	ff.AddRuleGroup(rg2)

	result := ff.Evaluate(map[string]string{"user_id": "u1", "role": "admin"})
	if result != "variant-a" {
		t.Fatalf("expected variant-a (priority 1), got %s", result)
	}
}

func TestFeatureFlag_Evaluate_RolloutPercentage(t *testing.T) {
	ff := domain.NewFeatureFlag("test", "rollout_test", "desc", "bool", "false", 100)
	ff.Activate()

	// 100% rollout — everyone gets "true" (the defaultValue for "on" based on flag type)
	result := ff.Evaluate(map[string]string{"user_id": "any-user"})
	if result != "true" {
		t.Fatalf("expected true for 100%% rollout, got %s", result)
	}
}

func TestFeatureFlag_Evaluate_RolloutZero(t *testing.T) {
	ff := domain.NewFeatureFlag("test", "rollout_zero", "desc", "bool", "false", 0)
	ff.Activate()

	result := ff.Evaluate(map[string]string{"user_id": "any-user"})
	if result != "false" {
		t.Fatalf("expected false for 0%% rollout, got %s", result)
	}
}

func TestFeatureFlag_Toggle(t *testing.T) {
	ff := domain.NewFeatureFlag("test", "test_key", "desc", "bool", "false", 0)

	ff.Activate()
	if !ff.IsActive() {
		t.Fatal("expected active after Activate")
	}

	ff.Deactivate()
	if ff.IsActive() {
		t.Fatal("expected inactive after Deactivate")
	}
}

func TestFeatureFlag_Evaluate_MultipleRuleGroups_AND(t *testing.T) {
	ff := domain.NewFeatureFlag("test", "test_key", "desc", "bool", "false", 0)
	ff.Activate()

	rg := domain.NewRuleGroup(ff.ID(), "UZ admins", "true", 1)
	rg.AddCondition(domain.NewCondition("role", domain.OpEq, "admin"))
	rg.AddCondition(domain.NewCondition("country", domain.OpEq, "UZ"))
	ff.AddRuleGroup(rg)

	// Both conditions match
	if ff.Evaluate(map[string]string{"user_id": "u1", "role": "admin", "country": "UZ"}) != "true" {
		t.Error("expected true when both conditions match")
	}

	// Only one matches
	if ff.Evaluate(map[string]string{"user_id": "u1", "role": "admin", "country": "US"}) != "false" {
		t.Error("expected false when country doesn't match")
	}
}

func TestFeatureFlag_ReconstructWithRuleGroups(t *testing.T) {
	flagID := uuid.New()
	rg := domain.ReconstructRuleGroup(
		uuid.New(), flagID, "test", "true", 1,
		timeNow(), timeNow(),
		[]domain.Condition{
			domain.ReconstructCondition(uuid.New(), uuid.New(), "role", domain.OpEq, "admin"),
		},
	)

	ff := domain.ReconstructFeatureFlag(
		flagID, timeNow(), timeNow(), nil,
		"test", "test_key", "desc", "bool", "false", 50, true,
		[]*domain.RuleGroup{rg},
	)

	if ff.Key() != "test_key" {
		t.Fatalf("expected key test_key, got %s", ff.Key())
	}
	if len(ff.RuleGroups()) != 1 {
		t.Fatalf("expected 1 rule group, got %d", len(ff.RuleGroups()))
	}
}
```

Add a helper at the bottom of the test file:

```go
func timeNow() time.Time {
	return time.Now()
}
```

And add `"time"` to the imports.

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go test ./internal/featureflag/domain/ -run TestNewFeatureFlag -v`
Expected: FAIL — wrong signature

- [ ] **Step 3: Rewrite entity.go**

Replace contents of `internal/featureflag/domain/entity.go`:

```go
package domain

import (
	"hash/fnv"
	"sort"
	"time"

	shared "gct/internal/shared/domain"

	"github.com/google/uuid"
)

// FeatureFlag is the aggregate root for feature flag management.
type FeatureFlag struct {
	shared.AggregateRoot
	name              string
	key               string
	description       string
	flagType          string // "bool", "string", "int", "float"
	defaultValue      string
	rolloutPercentage int // 0-100
	isActive          bool
	ruleGroups        []*RuleGroup
}

// NewFeatureFlag creates a new FeatureFlag aggregate. isActive defaults to false (must be explicitly activated).
func NewFeatureFlag(name, key, description, flagType, defaultValue string, rolloutPercentage int) *FeatureFlag {
	return &FeatureFlag{
		AggregateRoot:     shared.NewAggregateRoot(),
		name:              name,
		key:               key,
		description:       description,
		flagType:          flagType,
		defaultValue:      defaultValue,
		rolloutPercentage: rolloutPercentage,
		isActive:          false,
	}
}

// ReconstructFeatureFlag rebuilds a FeatureFlag aggregate from persisted data.
func ReconstructFeatureFlag(
	id uuid.UUID,
	createdAt, updatedAt time.Time,
	deletedAt *time.Time,
	name, key, description, flagType, defaultValue string,
	rolloutPercentage int,
	isActive bool,
	ruleGroups []*RuleGroup,
) *FeatureFlag {
	return &FeatureFlag{
		AggregateRoot:     shared.NewAggregateRootWithID(id, createdAt, updatedAt, deletedAt),
		name:              name,
		key:               key,
		description:       description,
		flagType:          flagType,
		defaultValue:      defaultValue,
		rolloutPercentage: rolloutPercentage,
		isActive:          isActive,
		ruleGroups:        ruleGroups,
	}
}

// Activate enables the feature flag.
func (ff *FeatureFlag) Activate() {
	ff.isActive = true
	ff.Touch()
	ff.AddEvent(NewFlagToggled(ff.ID(), true))
}

// Deactivate disables the feature flag (kill-switch).
func (ff *FeatureFlag) Deactivate() {
	ff.isActive = false
	ff.Touch()
	ff.AddEvent(NewFlagToggled(ff.ID(), false))
}

// AddRuleGroup adds a targeting rule group.
func (ff *FeatureFlag) AddRuleGroup(rg *RuleGroup) {
	ff.ruleGroups = append(ff.ruleGroups, rg)
}

// UpdateDetails applies partial modifications to the feature flag.
func (ff *FeatureFlag) UpdateDetails(name, key, description *string, flagType *string, defaultValue *string, rolloutPercentage *int, isActive *bool) {
	if name != nil {
		ff.name = *name
	}
	if key != nil {
		ff.key = *key
	}
	if description != nil {
		ff.description = *description
	}
	if flagType != nil {
		ff.flagType = *flagType
	}
	if defaultValue != nil {
		ff.defaultValue = *defaultValue
	}
	if rolloutPercentage != nil {
		ff.rolloutPercentage = *rolloutPercentage
	}
	if isActive != nil {
		ff.isActive = *isActive
	}
	ff.Touch()
}

// Evaluate determines the flag value for the given user attributes.
// 1. If inactive → defaultValue
// 2. Check rule groups by priority (ascending) — first full AND match wins
// 3. No match → rollout percentage check using hash(user_id + key)
// 4. Fallback → defaultValue
func (ff *FeatureFlag) Evaluate(userAttrs map[string]string) string {
	if !ff.isActive {
		return ff.defaultValue
	}

	// Sort by priority ascending
	sorted := make([]*RuleGroup, len(ff.ruleGroups))
	copy(sorted, ff.ruleGroups)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Priority() < sorted[j].Priority()
	})

	for _, rg := range sorted {
		if rg.MatchAll(userAttrs) {
			return rg.Variation()
		}
	}

	// Rollout percentage check
	if ff.rolloutPercentage > 0 {
		userID := userAttrs["user_id"]
		if userID != "" && ff.isInRollout(userID) {
			return ff.rolloutOnValue()
		}
	}

	return ff.defaultValue
}

// isInRollout uses FNV hash to deterministically assign a user to a rollout bucket.
func (ff *FeatureFlag) isInRollout(userID string) bool {
	h := fnv.New32a()
	h.Write([]byte(userID + ":" + ff.key))
	bucket := int(h.Sum32() % 100)
	return bucket < ff.rolloutPercentage
}

// rolloutOnValue returns the "on" value for the flag type.
func (ff *FeatureFlag) rolloutOnValue() string {
	switch ff.flagType {
	case "bool":
		return "true"
	default:
		return ff.defaultValue
	}
}

// Getters
func (ff *FeatureFlag) Name() string              { return ff.name }
func (ff *FeatureFlag) Key() string               { return ff.key }
func (ff *FeatureFlag) Description() string       { return ff.description }
func (ff *FeatureFlag) FlagType() string           { return ff.flagType }
func (ff *FeatureFlag) DefaultValue() string       { return ff.defaultValue }
func (ff *FeatureFlag) RolloutPercentage() int     { return ff.rolloutPercentage }
func (ff *FeatureFlag) IsActive() bool             { return ff.isActive }
func (ff *FeatureFlag) RuleGroups() []*RuleGroup   { return ff.ruleGroups }
```

- [ ] **Step 4: Run all domain tests**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go test ./internal/featureflag/domain/ -v`
Expected: PASS — all tests green

- [ ] **Step 5: Commit**

```bash
git add internal/featureflag/domain/entity.go internal/featureflag/domain/entity_test.go
git commit -m "feat(featureflag): rewrite FeatureFlag entity with Evaluate, rule groups, and rollout"
```

---

### Task 5: Update Domain Events and Errors

**Files:**
- Modify: `internal/featureflag/domain/event.go`
- Modify: `internal/featureflag/domain/error.go`

- [ ] **Step 1: Add new events**

Replace contents of `internal/featureflag/domain/event.go`:

```go
package domain

import (
	"time"

	"github.com/google/uuid"
)

// FlagToggled is emitted when a feature flag's active state changes.
type FlagToggled struct {
	aggregateID uuid.UUID
	occurredAt  time.Time
	Enabled     bool
}

func NewFlagToggled(id uuid.UUID, enabled bool) FlagToggled {
	return FlagToggled{aggregateID: id, occurredAt: time.Now(), Enabled: enabled}
}

func (e FlagToggled) EventName() string      { return "featureflag.toggled" }
func (e FlagToggled) OccurredAt() time.Time  { return e.occurredAt }
func (e FlagToggled) AggregateID() uuid.UUID { return e.aggregateID }

// FlagCreated is emitted when a new feature flag is created.
type FlagCreated struct {
	aggregateID uuid.UUID
	occurredAt  time.Time
}

func NewFlagCreated(id uuid.UUID) FlagCreated {
	return FlagCreated{aggregateID: id, occurredAt: time.Now()}
}

func (e FlagCreated) EventName() string      { return "featureflag.created" }
func (e FlagCreated) OccurredAt() time.Time  { return e.occurredAt }
func (e FlagCreated) AggregateID() uuid.UUID { return e.aggregateID }

// FlagUpdated is emitted when a feature flag is modified.
type FlagUpdated struct {
	aggregateID uuid.UUID
	occurredAt  time.Time
}

func NewFlagUpdated(id uuid.UUID) FlagUpdated {
	return FlagUpdated{aggregateID: id, occurredAt: time.Now()}
}

func (e FlagUpdated) EventName() string      { return "featureflag.updated" }
func (e FlagUpdated) OccurredAt() time.Time  { return e.occurredAt }
func (e FlagUpdated) AggregateID() uuid.UUID { return e.aggregateID }

// FlagDeleted is emitted when a feature flag is removed.
type FlagDeleted struct {
	aggregateID uuid.UUID
	occurredAt  time.Time
}

func NewFlagDeleted(id uuid.UUID) FlagDeleted {
	return FlagDeleted{aggregateID: id, occurredAt: time.Now()}
}

func (e FlagDeleted) EventName() string      { return "featureflag.deleted" }
func (e FlagDeleted) OccurredAt() time.Time  { return e.occurredAt }
func (e FlagDeleted) AggregateID() uuid.UUID { return e.aggregateID }
```

- [ ] **Step 2: Add rule group errors**

Replace contents of `internal/featureflag/domain/error.go`:

```go
package domain

import shared "gct/internal/shared/domain"

var (
	ErrFeatureFlagNotFound = shared.NewDomainError("FEATURE_FLAG_NOT_FOUND", "feature flag not found")
	ErrRuleGroupNotFound   = shared.NewDomainError("RULE_GROUP_NOT_FOUND", "rule group not found")
	ErrInvalidOperator     = shared.NewDomainError("INVALID_OPERATOR", "invalid condition operator")
	ErrDuplicateKey        = shared.NewDomainError("DUPLICATE_FLAG_KEY", "feature flag key already exists")
)
```

- [ ] **Step 3: Verify compile**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go build ./internal/featureflag/domain/`
Expected: Success

- [ ] **Step 4: Commit**

```bash
git add internal/featureflag/domain/event.go internal/featureflag/domain/error.go
git commit -m "feat(featureflag): add domain events and errors for rule groups"
```

---

### Task 6: Update Repository Interfaces and DTOs

**Files:**
- Modify: `internal/featureflag/domain/repository.go`
- Modify: `internal/featureflag/application/dto.go`

- [ ] **Step 1: Rewrite repository.go**

Replace contents of `internal/featureflag/domain/repository.go`:

```go
package domain

import (
	"context"

	"github.com/google/uuid"
)

// FeatureFlagFilter carries optional criteria for querying feature flags.
type FeatureFlagFilter struct {
	Search  *string
	Enabled *bool
	Limit   int64
	Offset  int64
}

// FeatureFlagView is a read-model DTO for feature flags.
type FeatureFlagView struct {
	ID                uuid.UUID        `json:"id"`
	Name              string           `json:"name"`
	Key               string           `json:"key"`
	Description       string           `json:"description"`
	FlagType          string           `json:"flag_type"`
	DefaultValue      string           `json:"default_value"`
	RolloutPercentage int              `json:"rollout_percentage"`
	IsActive          bool             `json:"is_active"`
	RuleGroups        []RuleGroupView  `json:"rule_groups"`
	CreatedAt         string           `json:"created_at"`
	UpdatedAt         string           `json:"updated_at"`
}

// RuleGroupView is a read-model DTO for rule groups.
type RuleGroupView struct {
	ID         uuid.UUID       `json:"id"`
	Name       string          `json:"name"`
	Variation  string          `json:"variation"`
	Priority   int             `json:"priority"`
	Conditions []ConditionView `json:"conditions"`
	CreatedAt  string          `json:"created_at"`
	UpdatedAt  string          `json:"updated_at"`
}

// ConditionView is a read-model DTO for conditions.
type ConditionView struct {
	ID        uuid.UUID `json:"id"`
	Attribute string    `json:"attribute"`
	Operator  string    `json:"operator"`
	Value     string    `json:"value"`
}

// FeatureFlagRepository is the write-side repository for the FeatureFlag aggregate.
type FeatureFlagRepository interface {
	Save(ctx context.Context, entity *FeatureFlag) error
	FindByID(ctx context.Context, id uuid.UUID) (*FeatureFlag, error)
	FindByKey(ctx context.Context, key string) (*FeatureFlag, error)
	Update(ctx context.Context, entity *FeatureFlag) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindAll(ctx context.Context) ([]*FeatureFlag, error)
}

// RuleGroupRepository is the write-side repository for rule groups.
type RuleGroupRepository interface {
	Save(ctx context.Context, rg *RuleGroup) error
	FindByID(ctx context.Context, id uuid.UUID) (*RuleGroup, error)
	Update(ctx context.Context, rg *RuleGroup) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByFlagID(ctx context.Context, flagID uuid.UUID) ([]*RuleGroup, error)
	SaveCondition(ctx context.Context, ruleGroupID uuid.UUID, c Condition) error
	DeleteConditionsByRuleGroupID(ctx context.Context, ruleGroupID uuid.UUID) error
}

// FeatureFlagReadRepository is the read-side (CQRS query) repository.
type FeatureFlagReadRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*FeatureFlagView, error)
	List(ctx context.Context, filter FeatureFlagFilter) ([]*FeatureFlagView, int64, error)
}

// Evaluator provides runtime feature flag evaluation.
type Evaluator interface {
	IsEnabled(ctx context.Context, flagKey string, userAttrs map[string]string) bool
	GetString(ctx context.Context, flagKey string, userAttrs map[string]string) string
	GetInt(ctx context.Context, flagKey string, userAttrs map[string]string) int
	GetFloat(ctx context.Context, flagKey string, userAttrs map[string]string) float64
}
```

- [ ] **Step 2: Rewrite dto.go**

Replace contents of `internal/featureflag/application/dto.go`:

```go
package application

import (
	"time"

	"github.com/google/uuid"
)

// FeatureFlagView is a read-model DTO returned by query handlers.
type FeatureFlagView struct {
	ID                uuid.UUID        `json:"id"`
	Name              string           `json:"name"`
	Key               string           `json:"key"`
	Description       string           `json:"description"`
	FlagType          string           `json:"flag_type"`
	DefaultValue      string           `json:"default_value"`
	RolloutPercentage int              `json:"rollout_percentage"`
	IsActive          bool             `json:"is_active"`
	RuleGroups        []RuleGroupView  `json:"rule_groups"`
	CreatedAt         time.Time        `json:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at"`
}

// RuleGroupView is a read-model DTO for rule groups.
type RuleGroupView struct {
	ID         uuid.UUID       `json:"id"`
	Name       string          `json:"name"`
	Variation  string          `json:"variation"`
	Priority   int             `json:"priority"`
	Conditions []ConditionView `json:"conditions"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

// ConditionView is a read-model DTO for conditions.
type ConditionView struct {
	ID        uuid.UUID `json:"id"`
	Attribute string    `json:"attribute"`
	Operator  string    `json:"operator"`
	Value     string    `json:"value"`
}
```

- [ ] **Step 3: Verify compile**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go build ./internal/featureflag/...`
Expected: May fail — downstream repos/handlers reference old types. This is expected; we fix them in the next tasks.

- [ ] **Step 4: Commit**

```bash
git add internal/featureflag/domain/repository.go internal/featureflag/application/dto.go
git commit -m "feat(featureflag): update repository interfaces and DTOs for rule groups"
```

---

### Task 7: Rewrite PostgreSQL Write Repo

**Files:**
- Modify: `internal/featureflag/infrastructure/postgres/write_repo.go`

- [ ] **Step 1: Rewrite write_repo.go**

Replace contents of `internal/featureflag/infrastructure/postgres/write_repo.go`:

```go
package postgres

import (
	"context"
	"time"

	"gct/internal/featureflag/domain"
	"gct/internal/shared/domain/consts"
	apperrors "gct/internal/shared/infrastructure/errors"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const tableName = consts.TableFeatureFlags

// FeatureFlagWriteRepo implements domain.FeatureFlagRepository using PostgreSQL.
type FeatureFlagWriteRepo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

// NewFeatureFlagWriteRepo creates a new FeatureFlagWriteRepo.
func NewFeatureFlagWriteRepo(pool *pgxpool.Pool) *FeatureFlagWriteRepo {
	return &FeatureFlagWriteRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// Save inserts a new FeatureFlag aggregate into the database.
func (r *FeatureFlagWriteRepo) Save(ctx context.Context, ff *domain.FeatureFlag) error {
	sql, args, err := r.builder.
		Insert(tableName).
		Columns("id", "key", "name", "flag_type", "value", "default_value", "description", "rollout_percentage", "is_active", "created_at", "updated_at").
		Values(
			ff.ID(),
			ff.Key(),
			ff.Name(),
			ff.FlagType(),
			ff.DefaultValue(),
			ff.DefaultValue(),
			ff.Description(),
			ff.RolloutPercentage(),
			ff.IsActive(),
			ff.CreatedAt(),
			ff.UpdatedAt(),
		).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildInsert)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}

// FindByID retrieves a FeatureFlag aggregate by ID (with rule groups and conditions).
func (r *FeatureFlagWriteRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.FeatureFlag, error) {
	sql, args, err := r.builder.
		Select("id", "key", "name", "flag_type", "default_value", "description", "rollout_percentage", "is_active", "created_at", "updated_at").
		From(tableName).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	row := r.pool.QueryRow(ctx, sql, args...)
	ff, err := scanFeatureFlag(row)
	if err != nil {
		return nil, err
	}

	ruleGroups, err := r.loadRuleGroups(ctx, ff.ID())
	if err != nil {
		return nil, err
	}
	for _, rg := range ruleGroups {
		ff.AddRuleGroup(rg)
	}

	return ff, nil
}

// FindByKey retrieves a FeatureFlag by its unique key (with rule groups and conditions).
func (r *FeatureFlagWriteRepo) FindByKey(ctx context.Context, key string) (*domain.FeatureFlag, error) {
	sql, args, err := r.builder.
		Select("id", "key", "name", "flag_type", "default_value", "description", "rollout_percentage", "is_active", "created_at", "updated_at").
		From(tableName).
		Where(squirrel.Eq{"key": key}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	row := r.pool.QueryRow(ctx, sql, args...)
	ff, err := scanFeatureFlag(row)
	if err != nil {
		return nil, err
	}

	ruleGroups, err := r.loadRuleGroups(ctx, ff.ID())
	if err != nil {
		return nil, err
	}
	for _, rg := range ruleGroups {
		ff.AddRuleGroup(rg)
	}

	return ff, nil
}

// FindAll retrieves all active feature flags with their rule groups and conditions.
func (r *FeatureFlagWriteRepo) FindAll(ctx context.Context) ([]*domain.FeatureFlag, error) {
	sql, args, err := r.builder.
		Select("id", "key", "name", "flag_type", "default_value", "description", "rollout_percentage", "is_active", "created_at", "updated_at").
		From(tableName).
		Where(squirrel.Eq{"deleted_at": nil}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, nil)
	}
	defer rows.Close()

	var flags []*domain.FeatureFlag
	for rows.Next() {
		ff, err := scanFeatureFlagFromRows(rows)
		if err != nil {
			return nil, err
		}
		flags = append(flags, ff)
	}

	// Load rule groups for each flag
	for _, ff := range flags {
		ruleGroups, err := r.loadRuleGroups(ctx, ff.ID())
		if err != nil {
			return nil, err
		}
		for _, rg := range ruleGroups {
			ff.AddRuleGroup(rg)
		}
	}

	return flags, nil
}

// Update updates a FeatureFlag aggregate in the database.
func (r *FeatureFlagWriteRepo) Update(ctx context.Context, ff *domain.FeatureFlag) error {
	sql, args, err := r.builder.
		Update(tableName).
		Set("name", ff.Name()).
		Set("key", ff.Key()).
		Set("flag_type", ff.FlagType()).
		Set("default_value", ff.DefaultValue()).
		Set("value", ff.DefaultValue()).
		Set("description", ff.Description()).
		Set("rollout_percentage", ff.RolloutPercentage()).
		Set("is_active", ff.IsActive()).
		Set("updated_at", ff.UpdatedAt()).
		Where(squirrel.Eq{"id": ff.ID()}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildUpdate)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}

// Delete removes a FeatureFlag by ID (cascades to rule groups and conditions).
func (r *FeatureFlagWriteRepo) Delete(ctx context.Context, id uuid.UUID) error {
	sql, args, err := r.builder.
		Delete(tableName).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildDelete)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}

// loadRuleGroups loads all rule groups and their conditions for a flag.
func (r *FeatureFlagWriteRepo) loadRuleGroups(ctx context.Context, flagID uuid.UUID) ([]*domain.RuleGroup, error) {
	rgSQL, rgArgs, err := r.builder.
		Select("id", "flag_id", "name", "variation", "priority", "created_at", "updated_at").
		From(consts.TableFeatureFlagRuleGroups).
		Where(squirrel.Eq{"flag_id": flagID}).
		OrderBy("priority ASC").
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	rgRows, err := r.pool.Query(ctx, rgSQL, rgArgs...)
	if err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableFeatureFlagRuleGroups, nil)
	}
	defer rgRows.Close()

	var ruleGroups []*domain.RuleGroup
	for rgRows.Next() {
		var (
			rgID      uuid.UUID
			rgFlagID  uuid.UUID
			name      string
			variation string
			priority  int
			createdAt time.Time
			updatedAt time.Time
		)
		if err := rgRows.Scan(&rgID, &rgFlagID, &name, &variation, &priority, &createdAt, &updatedAt); err != nil {
			return nil, apperrors.HandlePgError(err, consts.TableFeatureFlagRuleGroups, nil)
		}

		conditions, err := r.loadConditions(ctx, rgID)
		if err != nil {
			return nil, err
		}

		rg := domain.ReconstructRuleGroup(rgID, rgFlagID, name, variation, priority, createdAt, updatedAt, conditions)
		ruleGroups = append(ruleGroups, rg)
	}

	return ruleGroups, nil
}

// loadConditions loads all conditions for a rule group.
func (r *FeatureFlagWriteRepo) loadConditions(ctx context.Context, ruleGroupID uuid.UUID) ([]domain.Condition, error) {
	cSQL, cArgs, err := r.builder.
		Select("id", "rule_group_id", "attribute", "operator", "value").
		From(consts.TableFeatureFlagConditions).
		Where(squirrel.Eq{"rule_group_id": ruleGroupID}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	cRows, err := r.pool.Query(ctx, cSQL, cArgs...)
	if err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableFeatureFlagConditions, nil)
	}
	defer cRows.Close()

	var conditions []domain.Condition
	for cRows.Next() {
		var (
			cID       uuid.UUID
			cRGID     uuid.UUID
			attribute string
			operator  string
			value     string
		)
		if err := cRows.Scan(&cID, &cRGID, &attribute, &operator, &value); err != nil {
			return nil, apperrors.HandlePgError(err, consts.TableFeatureFlagConditions, nil)
		}
		conditions = append(conditions, domain.ReconstructCondition(cID, cRGID, attribute, operator, value))
	}

	return conditions, nil
}

func scanFeatureFlag(row pgx.Row) (*domain.FeatureFlag, error) {
	var (
		id                uuid.UUID
		key               string
		name              string
		flagType          string
		defaultValue      string
		description       string
		rolloutPercentage int
		isActive          bool
		createdAt         time.Time
		updatedAt         time.Time
	)

	err := row.Scan(&id, &key, &name, &flagType, &defaultValue, &description, &rolloutPercentage, &isActive, &createdAt, &updatedAt)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, map[string]any{"id": id})
	}

	return domain.ReconstructFeatureFlag(id, createdAt, updatedAt, nil, name, key, description, flagType, defaultValue, rolloutPercentage, isActive, nil), nil
}

func scanFeatureFlagFromRows(rows pgx.Rows) (*domain.FeatureFlag, error) {
	var (
		id                uuid.UUID
		key               string
		name              string
		flagType          string
		defaultValue      string
		description       string
		rolloutPercentage int
		isActive          bool
		createdAt         time.Time
		updatedAt         time.Time
	)

	err := rows.Scan(&id, &key, &name, &flagType, &defaultValue, &description, &rolloutPercentage, &isActive, &createdAt, &updatedAt)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, nil)
	}

	return domain.ReconstructFeatureFlag(id, createdAt, updatedAt, nil, name, key, description, flagType, defaultValue, rolloutPercentage, isActive, nil), nil
}
```

- [ ] **Step 2: Verify compile**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go build ./internal/featureflag/infrastructure/postgres/`
Expected: Success (or minor issues if read_repo still references old types)

- [ ] **Step 3: Commit**

```bash
git add internal/featureflag/infrastructure/postgres/write_repo.go
git commit -m "feat(featureflag): rewrite write repo for new schema with rule groups"
```

---

### Task 8: RuleGroup Write Repo

**Files:**
- Create: `internal/featureflag/infrastructure/postgres/rule_group_repo.go`

- [ ] **Step 1: Implement rule group repo**

File: `internal/featureflag/infrastructure/postgres/rule_group_repo.go`

```go
package postgres

import (
	"context"
	"time"

	"gct/internal/featureflag/domain"
	"gct/internal/shared/domain/consts"
	apperrors "gct/internal/shared/infrastructure/errors"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RuleGroupWriteRepo implements domain.RuleGroupRepository using PostgreSQL.
type RuleGroupWriteRepo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

// NewRuleGroupWriteRepo creates a new RuleGroupWriteRepo.
func NewRuleGroupWriteRepo(pool *pgxpool.Pool) *RuleGroupWriteRepo {
	return &RuleGroupWriteRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// Save inserts a new RuleGroup with its conditions.
func (r *RuleGroupWriteRepo) Save(ctx context.Context, rg *domain.RuleGroup) error {
	sql, args, err := r.builder.
		Insert(consts.TableFeatureFlagRuleGroups).
		Columns("id", "flag_id", "name", "variation", "priority", "created_at", "updated_at").
		Values(rg.ID(), rg.FlagID(), rg.Name(), rg.Variation(), rg.Priority(), rg.CreatedAt(), rg.UpdatedAt()).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildInsert)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, consts.TableFeatureFlagRuleGroups, nil)
	}

	// Save conditions
	for _, c := range rg.Conditions() {
		if err := r.SaveCondition(ctx, rg.ID(), c); err != nil {
			return err
		}
	}

	return nil
}

// FindByID retrieves a RuleGroup by ID with its conditions.
func (r *RuleGroupWriteRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.RuleGroup, error) {
	sql, args, err := r.builder.
		Select("id", "flag_id", "name", "variation", "priority", "created_at", "updated_at").
		From(consts.TableFeatureFlagRuleGroups).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	var (
		rgID      uuid.UUID
		flagID    uuid.UUID
		name      string
		variation string
		priority  int
		createdAt time.Time
		updatedAt time.Time
	)

	row := r.pool.QueryRow(ctx, sql, args...)
	if err := row.Scan(&rgID, &flagID, &name, &variation, &priority, &createdAt, &updatedAt); err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableFeatureFlagRuleGroups, map[string]any{"id": id})
	}

	// Load conditions
	conditions, err := r.loadConditions(ctx, rgID)
	if err != nil {
		return nil, err
	}

	return domain.ReconstructRuleGroup(rgID, flagID, name, variation, priority, createdAt, updatedAt, conditions), nil
}

// Update updates a RuleGroup. Replaces all conditions (delete + re-insert).
func (r *RuleGroupWriteRepo) Update(ctx context.Context, rg *domain.RuleGroup) error {
	sql, args, err := r.builder.
		Update(consts.TableFeatureFlagRuleGroups).
		Set("name", rg.Name()).
		Set("variation", rg.Variation()).
		Set("priority", rg.Priority()).
		Set("updated_at", rg.UpdatedAt()).
		Where(squirrel.Eq{"id": rg.ID()}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildUpdate)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, consts.TableFeatureFlagRuleGroups, nil)
	}

	// Replace conditions: delete old, insert new
	if err := r.DeleteConditionsByRuleGroupID(ctx, rg.ID()); err != nil {
		return err
	}
	for _, c := range rg.Conditions() {
		if err := r.SaveCondition(ctx, rg.ID(), c); err != nil {
			return err
		}
	}

	return nil
}

// Delete removes a RuleGroup by ID (conditions cascade via FK).
func (r *RuleGroupWriteRepo) Delete(ctx context.Context, id uuid.UUID) error {
	sql, args, err := r.builder.
		Delete(consts.TableFeatureFlagRuleGroups).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildDelete)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, consts.TableFeatureFlagRuleGroups, nil)
	}

	return nil
}

// FindByFlagID retrieves all rule groups for a flag, ordered by priority.
func (r *RuleGroupWriteRepo) FindByFlagID(ctx context.Context, flagID uuid.UUID) ([]*domain.RuleGroup, error) {
	sql, args, err := r.builder.
		Select("id", "flag_id", "name", "variation", "priority", "created_at", "updated_at").
		From(consts.TableFeatureFlagRuleGroups).
		Where(squirrel.Eq{"flag_id": flagID}).
		OrderBy("priority ASC").
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableFeatureFlagRuleGroups, nil)
	}
	defer rows.Close()

	var ruleGroups []*domain.RuleGroup
	for rows.Next() {
		var (
			rgID      uuid.UUID
			rgFlagID  uuid.UUID
			name      string
			variation string
			priority  int
			createdAt time.Time
			updatedAt time.Time
		)
		if err := rows.Scan(&rgID, &rgFlagID, &name, &variation, &priority, &createdAt, &updatedAt); err != nil {
			return nil, apperrors.HandlePgError(err, consts.TableFeatureFlagRuleGroups, nil)
		}

		conditions, err := r.loadConditions(ctx, rgID)
		if err != nil {
			return nil, err
		}

		ruleGroups = append(ruleGroups, domain.ReconstructRuleGroup(rgID, rgFlagID, name, variation, priority, createdAt, updatedAt, conditions))
	}

	return ruleGroups, nil
}

// SaveCondition inserts a single condition.
func (r *RuleGroupWriteRepo) SaveCondition(ctx context.Context, ruleGroupID uuid.UUID, c domain.Condition) error {
	sql, args, err := r.builder.
		Insert(consts.TableFeatureFlagConditions).
		Columns("id", "rule_group_id", "attribute", "operator", "value", "created_at").
		Values(c.ID(), ruleGroupID, c.Attribute(), c.Operator(), c.Value(), time.Now()).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildInsert)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, consts.TableFeatureFlagConditions, nil)
	}

	return nil
}

// DeleteConditionsByRuleGroupID removes all conditions for a rule group.
func (r *RuleGroupWriteRepo) DeleteConditionsByRuleGroupID(ctx context.Context, ruleGroupID uuid.UUID) error {
	sql, args, err := r.builder.
		Delete(consts.TableFeatureFlagConditions).
		Where(squirrel.Eq{"rule_group_id": ruleGroupID}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildDelete)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, consts.TableFeatureFlagConditions, nil)
	}

	return nil
}

func (r *RuleGroupWriteRepo) loadConditions(ctx context.Context, ruleGroupID uuid.UUID) ([]domain.Condition, error) {
	sql, args, err := r.builder.
		Select("id", "rule_group_id", "attribute", "operator", "value").
		From(consts.TableFeatureFlagConditions).
		Where(squirrel.Eq{"rule_group_id": ruleGroupID}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableFeatureFlagConditions, nil)
	}
	defer rows.Close()

	var conditions []domain.Condition
	for rows.Next() {
		var (
			cID       uuid.UUID
			cRGID     uuid.UUID
			attribute string
			operator  string
			value     string
		)
		if err := rows.Scan(&cID, &cRGID, &attribute, &operator, &value); err != nil {
			return nil, apperrors.HandlePgError(err, consts.TableFeatureFlagConditions, nil)
		}
		conditions = append(conditions, domain.ReconstructCondition(cID, cRGID, attribute, operator, value))
	}

	return conditions, nil
}
```

- [ ] **Step 2: Verify compile**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go build ./internal/featureflag/infrastructure/postgres/`
Expected: Success

- [ ] **Step 3: Commit**

```bash
git add internal/featureflag/infrastructure/postgres/rule_group_repo.go
git commit -m "feat(featureflag): add RuleGroup write repo with conditions"
```

---

### Task 9: Rewrite Read Repo

**Files:**
- Modify: `internal/featureflag/infrastructure/postgres/read_repo.go`

- [ ] **Step 1: Rewrite read_repo.go**

Replace contents of `internal/featureflag/infrastructure/postgres/read_repo.go`:

```go
package postgres

import (
	"context"
	"time"

	"gct/internal/featureflag/domain"
	"gct/internal/shared/domain/consts"
	apperrors "gct/internal/shared/infrastructure/errors"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// FeatureFlagReadRepo implements domain.FeatureFlagReadRepository for the CQRS read side.
type FeatureFlagReadRepo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

// NewFeatureFlagReadRepo creates a new FeatureFlagReadRepo.
func NewFeatureFlagReadRepo(pool *pgxpool.Pool) *FeatureFlagReadRepo {
	return &FeatureFlagReadRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// FindByID returns a FeatureFlagView with rule groups and conditions.
func (r *FeatureFlagReadRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.FeatureFlagView, error) {
	sql, args, err := r.builder.
		Select("id", "key", "name", "flag_type", "default_value", "description", "rollout_percentage", "is_active", "created_at", "updated_at").
		From(tableName).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	var (
		fID               uuid.UUID
		key               string
		name              string
		flagType          string
		defaultValue      string
		description       string
		rolloutPercentage int
		isActive          bool
		createdAt         time.Time
		updatedAt         time.Time
	)

	row := r.pool.QueryRow(ctx, sql, args...)
	if err := row.Scan(&fID, &key, &name, &flagType, &defaultValue, &description, &rolloutPercentage, &isActive, &createdAt, &updatedAt); err != nil {
		return nil, apperrors.HandlePgError(err, tableName, map[string]any{"id": id})
	}

	ruleGroups, err := r.loadRuleGroupViews(ctx, fID)
	if err != nil {
		return nil, err
	}

	return &domain.FeatureFlagView{
		ID:                fID,
		Name:              name,
		Key:               key,
		Description:       description,
		FlagType:          flagType,
		DefaultValue:      defaultValue,
		RolloutPercentage: rolloutPercentage,
		IsActive:          isActive,
		RuleGroups:        ruleGroups,
		CreatedAt:         createdAt.Format(time.RFC3339),
		UpdatedAt:         updatedAt.Format(time.RFC3339),
	}, nil
}

// List returns a paginated list of FeatureFlagView with optional filters.
func (r *FeatureFlagReadRepo) List(ctx context.Context, filter domain.FeatureFlagFilter) ([]*domain.FeatureFlagView, int64, error) {
	conds := squirrel.And{}
	if filter.Search != nil {
		conds = append(conds, squirrel.ILike{"name": "%" + *filter.Search + "%"})
	}
	if filter.Enabled != nil {
		conds = append(conds, squirrel.Eq{"is_active": *filter.Enabled})
	}

	// Count total
	countQB := r.builder.Select("COUNT(*)").From(tableName)
	if len(conds) > 0 {
		countQB = countQB.Where(conds)
	}
	countSQL, countArgs, err := countQB.ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	var total int64
	if err = r.pool.QueryRow(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return nil, 0, apperrors.HandlePgError(err, tableName, nil)
	}

	// Fetch page
	limit := filter.Limit
	if limit <= 0 {
		limit = 10
	}
	qb := r.builder.
		Select("id", "key", "name", "flag_type", "default_value", "description", "rollout_percentage", "is_active", "created_at", "updated_at").
		From(tableName).
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(filter.Offset))

	if len(conds) > 0 {
		qb = qb.Where(conds)
	}

	sql, args, err := qb.ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, apperrors.HandlePgError(err, tableName, nil)
	}
	defer rows.Close()

	var views []*domain.FeatureFlagView
	for rows.Next() {
		var (
			fID               uuid.UUID
			key               string
			name              string
			flagType          string
			defaultValue      string
			description       string
			rolloutPercentage int
			isActive          bool
			createdAt         time.Time
			updatedAt         time.Time
		)
		if err := rows.Scan(&fID, &key, &name, &flagType, &defaultValue, &description, &rolloutPercentage, &isActive, &createdAt, &updatedAt); err != nil {
			return nil, 0, apperrors.HandlePgError(err, tableName, nil)
		}

		ruleGroups, err := r.loadRuleGroupViews(ctx, fID)
		if err != nil {
			return nil, 0, err
		}

		views = append(views, &domain.FeatureFlagView{
			ID:                fID,
			Name:              name,
			Key:               key,
			Description:       description,
			FlagType:          flagType,
			DefaultValue:      defaultValue,
			RolloutPercentage: rolloutPercentage,
			IsActive:          isActive,
			RuleGroups:        ruleGroups,
			CreatedAt:         createdAt.Format(time.RFC3339),
			UpdatedAt:         updatedAt.Format(time.RFC3339),
		})
	}

	return views, total, nil
}

func (r *FeatureFlagReadRepo) loadRuleGroupViews(ctx context.Context, flagID uuid.UUID) ([]domain.RuleGroupView, error) {
	sql, args, err := r.builder.
		Select("id", "name", "variation", "priority", "created_at", "updated_at").
		From(consts.TableFeatureFlagRuleGroups).
		Where(squirrel.Eq{"flag_id": flagID}).
		OrderBy("priority ASC").
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableFeatureFlagRuleGroups, nil)
	}
	defer rows.Close()

	var views []domain.RuleGroupView
	for rows.Next() {
		var (
			rgID      uuid.UUID
			name      string
			variation string
			priority  int
			createdAt time.Time
			updatedAt time.Time
		)
		if err := rows.Scan(&rgID, &name, &variation, &priority, &createdAt, &updatedAt); err != nil {
			return nil, apperrors.HandlePgError(err, consts.TableFeatureFlagRuleGroups, nil)
		}

		conditions, err := r.loadConditionViews(ctx, rgID)
		if err != nil {
			return nil, err
		}

		views = append(views, domain.RuleGroupView{
			ID:         rgID,
			Name:       name,
			Variation:  variation,
			Priority:   priority,
			Conditions: conditions,
			CreatedAt:  createdAt.Format(time.RFC3339),
			UpdatedAt:  updatedAt.Format(time.RFC3339),
		})
	}

	return views, nil
}

func (r *FeatureFlagReadRepo) loadConditionViews(ctx context.Context, ruleGroupID uuid.UUID) ([]domain.ConditionView, error) {
	sql, args, err := r.builder.
		Select("id", "attribute", "operator", "value").
		From(consts.TableFeatureFlagConditions).
		Where(squirrel.Eq{"rule_group_id": ruleGroupID}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableFeatureFlagConditions, nil)
	}
	defer rows.Close()

	var views []domain.ConditionView
	for rows.Next() {
		var (
			cID       uuid.UUID
			attribute string
			operator  string
			value     string
		)
		if err := rows.Scan(&cID, &attribute, &operator, &value); err != nil {
			return nil, apperrors.HandlePgError(err, consts.TableFeatureFlagConditions, nil)
		}
		views = append(views, domain.ConditionView{
			ID:        cID,
			Attribute: attribute,
			Operator:  operator,
			Value:     value,
		})
	}

	return views, nil
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/featureflag/infrastructure/postgres/read_repo.go
git commit -m "feat(featureflag): rewrite read repo with rule group and condition views"
```

---

### Task 10: Rewrite Command Handlers

**Files:**
- Modify: `internal/featureflag/application/command/create.go`
- Modify: `internal/featureflag/application/command/update.go`
- Modify: `internal/featureflag/application/command/delete.go`
- Create: `internal/featureflag/application/command/create_rule_group.go`
- Create: `internal/featureflag/application/command/update_rule_group.go`
- Create: `internal/featureflag/application/command/delete_rule_group.go`

- [ ] **Step 1: Rewrite create.go**

Replace contents of `internal/featureflag/application/command/create.go`:

```go
package command

import (
	"context"

	"gct/internal/featureflag/domain"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"
)

// CreateCommand represents an intent to register a new feature flag.
type CreateCommand struct {
	Name              string
	Key               string
	Description       string
	FlagType          string
	DefaultValue      string
	RolloutPercentage int
	IsActive          bool
}

// CreateHandler orchestrates feature flag creation.
type CreateHandler struct {
	repo     domain.FeatureFlagRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewCreateHandler wires dependencies for feature flag creation.
func NewCreateHandler(repo domain.FeatureFlagRepository, eventBus application.EventBus, logger logger.Log) *CreateHandler {
	return &CreateHandler{repo: repo, eventBus: eventBus, logger: logger}
}

// Handle persists a new feature flag and publishes domain events.
func (h *CreateHandler) Handle(ctx context.Context, cmd CreateCommand) error {
	ff := domain.NewFeatureFlag(cmd.Name, cmd.Key, cmd.Description, cmd.FlagType, cmd.DefaultValue, cmd.RolloutPercentage)
	if cmd.IsActive {
		ff.Activate()
	}

	if err := h.repo.Save(ctx, ff); err != nil {
		h.logger.Errorf("failed to save feature flag: %v", err)
		return err
	}

	ff.AddEvent(domain.NewFlagCreated(ff.ID()))
	if err := h.eventBus.Publish(ctx, ff.Events()...); err != nil {
		h.logger.Errorf("failed to publish events: %v", err)
	}

	return nil
}
```

- [ ] **Step 2: Rewrite update.go**

Replace contents of `internal/featureflag/application/command/update.go`:

```go
package command

import (
	"context"

	"gct/internal/featureflag/domain"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"

	"github.com/google/uuid"
)

// UpdateCommand represents a partial update to an existing feature flag.
type UpdateCommand struct {
	ID                uuid.UUID
	Name              *string
	Key               *string
	Description       *string
	FlagType          *string
	DefaultValue      *string
	RolloutPercentage *int
	IsActive          *bool
}

// UpdateHandler applies partial modifications to an existing feature flag.
type UpdateHandler struct {
	repo     domain.FeatureFlagRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewUpdateHandler wires dependencies for feature flag updates.
func NewUpdateHandler(repo domain.FeatureFlagRepository, eventBus application.EventBus, logger logger.Log) *UpdateHandler {
	return &UpdateHandler{repo: repo, eventBus: eventBus, logger: logger}
}

// Handle fetches the flag by ID, applies non-nil field updates, persists, and publishes events.
func (h *UpdateHandler) Handle(ctx context.Context, cmd UpdateCommand) error {
	ff, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	ff.UpdateDetails(cmd.Name, cmd.Key, cmd.Description, cmd.FlagType, cmd.DefaultValue, cmd.RolloutPercentage, cmd.IsActive)

	if err := h.repo.Update(ctx, ff); err != nil {
		h.logger.Errorf("failed to update feature flag: %v", err)
		return err
	}

	ff.AddEvent(domain.NewFlagUpdated(ff.ID()))
	if err := h.eventBus.Publish(ctx, ff.Events()...); err != nil {
		h.logger.Errorf("failed to publish events: %v", err)
	}

	return nil
}
```

- [ ] **Step 3: Rewrite delete.go**

Replace contents of `internal/featureflag/application/command/delete.go`:

```go
package command

import (
	"context"

	"gct/internal/featureflag/domain"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"

	"github.com/google/uuid"
)

// DeleteCommand represents an intent to permanently remove a feature flag.
type DeleteCommand struct {
	ID uuid.UUID
}

// DeleteHandler performs hard deletion of a feature flag.
type DeleteHandler struct {
	repo     domain.FeatureFlagRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewDeleteHandler wires dependencies for feature flag deletion.
func NewDeleteHandler(repo domain.FeatureFlagRepository, eventBus application.EventBus, logger logger.Log) *DeleteHandler {
	return &DeleteHandler{repo: repo, eventBus: eventBus, logger: logger}
}

// Handle deletes the feature flag and publishes a FlagDeleted event.
func (h *DeleteHandler) Handle(ctx context.Context, cmd DeleteCommand) error {
	if err := h.repo.Delete(ctx, cmd.ID); err != nil {
		h.logger.Errorf("failed to delete feature flag: %v", err)
		return err
	}

	if err := h.eventBus.Publish(ctx, domain.NewFlagDeleted(cmd.ID)); err != nil {
		h.logger.Errorf("failed to publish events: %v", err)
	}

	return nil
}
```

- [ ] **Step 4: Create create_rule_group.go**

File: `internal/featureflag/application/command/create_rule_group.go`

```go
package command

import (
	"context"

	"gct/internal/featureflag/domain"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"

	"github.com/google/uuid"
)

// ConditionInput holds data for creating a condition.
type ConditionInput struct {
	Attribute string
	Operator  string
	Value     string
}

// CreateRuleGroupCommand represents an intent to add a rule group to a flag.
type CreateRuleGroupCommand struct {
	FlagID     uuid.UUID
	Name       string
	Variation  string
	Priority   int
	Conditions []ConditionInput
}

// CreateRuleGroupHandler creates a new rule group with conditions.
type CreateRuleGroupHandler struct {
	flagRepo  domain.FeatureFlagRepository
	rgRepo    domain.RuleGroupRepository
	eventBus  application.EventBus
	logger    logger.Log
}

// NewCreateRuleGroupHandler wires dependencies.
func NewCreateRuleGroupHandler(flagRepo domain.FeatureFlagRepository, rgRepo domain.RuleGroupRepository, eventBus application.EventBus, logger logger.Log) *CreateRuleGroupHandler {
	return &CreateRuleGroupHandler{flagRepo: flagRepo, rgRepo: rgRepo, eventBus: eventBus, logger: logger}
}

// Handle creates the rule group and publishes a FlagUpdated event.
func (h *CreateRuleGroupHandler) Handle(ctx context.Context, cmd CreateRuleGroupCommand) error {
	// Verify flag exists
	if _, err := h.flagRepo.FindByID(ctx, cmd.FlagID); err != nil {
		return err
	}

	// Validate operators
	for _, c := range cmd.Conditions {
		if !domain.IsValidOperator(c.Operator) {
			return domain.ErrInvalidOperator
		}
	}

	rg := domain.NewRuleGroup(cmd.FlagID, cmd.Name, cmd.Variation, cmd.Priority)
	for _, c := range cmd.Conditions {
		rg.AddCondition(domain.NewCondition(c.Attribute, c.Operator, c.Value))
	}

	if err := h.rgRepo.Save(ctx, rg); err != nil {
		h.logger.Errorf("failed to save rule group: %v", err)
		return err
	}

	if err := h.eventBus.Publish(ctx, domain.NewFlagUpdated(cmd.FlagID)); err != nil {
		h.logger.Errorf("failed to publish events: %v", err)
	}

	return nil
}
```

- [ ] **Step 5: Create update_rule_group.go**

File: `internal/featureflag/application/command/update_rule_group.go`

```go
package command

import (
	"context"

	"gct/internal/featureflag/domain"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"

	"github.com/google/uuid"
)

// UpdateRuleGroupCommand represents a partial update to a rule group.
type UpdateRuleGroupCommand struct {
	ID         uuid.UUID
	Name       *string
	Variation  *string
	Priority   *int
	Conditions *[]ConditionInput // nil = don't change, non-nil = replace all
}

// UpdateRuleGroupHandler modifies an existing rule group.
type UpdateRuleGroupHandler struct {
	rgRepo   domain.RuleGroupRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewUpdateRuleGroupHandler wires dependencies.
func NewUpdateRuleGroupHandler(rgRepo domain.RuleGroupRepository, eventBus application.EventBus, logger logger.Log) *UpdateRuleGroupHandler {
	return &UpdateRuleGroupHandler{rgRepo: rgRepo, eventBus: eventBus, logger: logger}
}

// Handle fetches, updates, persists, and publishes.
func (h *UpdateRuleGroupHandler) Handle(ctx context.Context, cmd UpdateRuleGroupCommand) error {
	rg, err := h.rgRepo.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	rg.UpdateDetails(cmd.Name, cmd.Variation, cmd.Priority)

	// Replace conditions if provided
	if cmd.Conditions != nil {
		for _, c := range *cmd.Conditions {
			if !domain.IsValidOperator(c.Operator) {
				return domain.ErrInvalidOperator
			}
		}

		// Rebuild conditions on the rule group
		rebuilt := domain.ReconstructRuleGroup(rg.ID(), rg.FlagID(), rg.Name(), rg.Variation(), rg.Priority(), rg.CreatedAt(), rg.UpdatedAt(), nil)
		for _, c := range *cmd.Conditions {
			rebuilt.AddCondition(domain.NewCondition(c.Attribute, c.Operator, c.Value))
		}
		rg = rebuilt
	}

	if err := h.rgRepo.Update(ctx, rg); err != nil {
		h.logger.Errorf("failed to update rule group: %v", err)
		return err
	}

	if err := h.eventBus.Publish(ctx, domain.NewFlagUpdated(rg.FlagID())); err != nil {
		h.logger.Errorf("failed to publish events: %v", err)
	}

	return nil
}
```

- [ ] **Step 6: Create delete_rule_group.go**

File: `internal/featureflag/application/command/delete_rule_group.go`

```go
package command

import (
	"context"

	"gct/internal/featureflag/domain"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"

	"github.com/google/uuid"
)

// DeleteRuleGroupCommand removes a rule group by ID.
type DeleteRuleGroupCommand struct {
	ID uuid.UUID
}

// DeleteRuleGroupHandler handles rule group deletion.
type DeleteRuleGroupHandler struct {
	rgRepo   domain.RuleGroupRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewDeleteRuleGroupHandler wires dependencies.
func NewDeleteRuleGroupHandler(rgRepo domain.RuleGroupRepository, eventBus application.EventBus, logger logger.Log) *DeleteRuleGroupHandler {
	return &DeleteRuleGroupHandler{rgRepo: rgRepo, eventBus: eventBus, logger: logger}
}

// Handle deletes the rule group and publishes a FlagUpdated event.
func (h *DeleteRuleGroupHandler) Handle(ctx context.Context, cmd DeleteRuleGroupCommand) error {
	rg, err := h.rgRepo.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	if err := h.rgRepo.Delete(ctx, cmd.ID); err != nil {
		h.logger.Errorf("failed to delete rule group: %v", err)
		return err
	}

	if err := h.eventBus.Publish(ctx, domain.NewFlagUpdated(rg.FlagID())); err != nil {
		h.logger.Errorf("failed to publish events: %v", err)
	}

	return nil
}
```

- [ ] **Step 7: Verify compile**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go build ./internal/featureflag/application/...`
Expected: Success

- [ ] **Step 8: Commit**

```bash
git add internal/featureflag/application/command/
git commit -m "feat(featureflag): rewrite command handlers with rule group CRUD"
```

---

### Task 11: Rewrite Query Handlers

**Files:**
- Modify: `internal/featureflag/application/query/get.go`
- Modify: `internal/featureflag/application/query/list.go`

- [ ] **Step 1: Rewrite get.go**

Replace contents of `internal/featureflag/application/query/get.go`:

```go
package query

import (
	"context"

	appdto "gct/internal/featureflag/application"
	"gct/internal/featureflag/domain"

	"github.com/google/uuid"
)

// GetQuery holds the input for fetching a single feature flag.
type GetQuery struct {
	ID uuid.UUID
}

// GetHandler handles the GetQuery.
type GetHandler struct {
	readRepo domain.FeatureFlagReadRepository
}

// NewGetHandler creates a new GetHandler.
func NewGetHandler(readRepo domain.FeatureFlagReadRepository) *GetHandler {
	return &GetHandler{readRepo: readRepo}
}

// Handle executes the GetQuery and returns a FeatureFlagView.
func (h *GetHandler) Handle(ctx context.Context, q GetQuery) (*appdto.FeatureFlagView, error) {
	view, err := h.readRepo.FindByID(ctx, q.ID)
	if err != nil {
		return nil, err
	}

	return mapToAppView(view), nil
}

func mapToAppView(v *domain.FeatureFlagView) *appdto.FeatureFlagView {
	ruleGroups := make([]appdto.RuleGroupView, len(v.RuleGroups))
	for i, rg := range v.RuleGroups {
		conditions := make([]appdto.ConditionView, len(rg.Conditions))
		for j, c := range rg.Conditions {
			conditions[j] = appdto.ConditionView{
				ID:        c.ID,
				Attribute: c.Attribute,
				Operator:  c.Operator,
				Value:     c.Value,
			}
		}
		ruleGroups[i] = appdto.RuleGroupView{
			ID:         rg.ID,
			Name:       rg.Name,
			Variation:  rg.Variation,
			Priority:   rg.Priority,
			Conditions: conditions,
		}
	}

	return &appdto.FeatureFlagView{
		ID:                v.ID,
		Name:              v.Name,
		Key:               v.Key,
		Description:       v.Description,
		FlagType:          v.FlagType,
		DefaultValue:      v.DefaultValue,
		RolloutPercentage: v.RolloutPercentage,
		IsActive:          v.IsActive,
		RuleGroups:        ruleGroups,
	}
}
```

- [ ] **Step 2: Rewrite list.go**

Replace contents of `internal/featureflag/application/query/list.go`:

```go
package query

import (
	"context"

	appdto "gct/internal/featureflag/application"
	"gct/internal/featureflag/domain"
)

// ListQuery holds the input for listing feature flags with filtering.
type ListQuery struct {
	Filter domain.FeatureFlagFilter
}

// ListResult holds the output of the list feature flags query.
type ListResult struct {
	Flags []*appdto.FeatureFlagView
	Total int64
}

// ListHandler handles the ListQuery.
type ListHandler struct {
	readRepo domain.FeatureFlagReadRepository
}

// NewListHandler creates a new ListHandler.
func NewListHandler(readRepo domain.FeatureFlagReadRepository) *ListHandler {
	return &ListHandler{readRepo: readRepo}
}

// Handle executes the ListQuery and returns a list of FeatureFlagView with total count.
func (h *ListHandler) Handle(ctx context.Context, q ListQuery) (*ListResult, error) {
	views, total, err := h.readRepo.List(ctx, q.Filter)
	if err != nil {
		return nil, err
	}

	result := make([]*appdto.FeatureFlagView, len(views))
	for i, v := range views {
		result[i] = mapToAppView(v)
	}

	return &ListResult{Flags: result, Total: total}, nil
}
```

- [ ] **Step 3: Commit**

```bash
git add internal/featureflag/application/query/
git commit -m "feat(featureflag): rewrite query handlers for new FeatureFlagView"
```

---

### Task 12: PostgreSQL Evaluator

**Files:**
- Create: `internal/featureflag/infrastructure/postgres/evaluator.go`

- [ ] **Step 1: Implement evaluator**

File: `internal/featureflag/infrastructure/postgres/evaluator.go`

```go
package postgres

import (
	"context"
	"strconv"

	"gct/internal/featureflag/domain"
)

// PgEvaluator implements domain.Evaluator by reading flags from PostgreSQL.
type PgEvaluator struct {
	repo domain.FeatureFlagRepository
}

// NewPgEvaluator creates a new PostgreSQL-backed evaluator.
func NewPgEvaluator(repo domain.FeatureFlagRepository) *PgEvaluator {
	return &PgEvaluator{repo: repo}
}

// IsEnabled checks if a boolean flag is enabled for the given user attributes.
func (e *PgEvaluator) IsEnabled(ctx context.Context, flagKey string, userAttrs map[string]string) bool {
	ff, err := e.repo.FindByKey(ctx, flagKey)
	if err != nil {
		return false
	}
	return ff.Evaluate(userAttrs) == "true"
}

// GetString returns the string value of a flag for the given user attributes.
func (e *PgEvaluator) GetString(ctx context.Context, flagKey string, userAttrs map[string]string) string {
	ff, err := e.repo.FindByKey(ctx, flagKey)
	if err != nil {
		return ""
	}
	return ff.Evaluate(userAttrs)
}

// GetInt returns the int value of a flag for the given user attributes.
func (e *PgEvaluator) GetInt(ctx context.Context, flagKey string, userAttrs map[string]string) int {
	ff, err := e.repo.FindByKey(ctx, flagKey)
	if err != nil {
		return 0
	}
	val, err := strconv.Atoi(ff.Evaluate(userAttrs))
	if err != nil {
		return 0
	}
	return val
}

// GetFloat returns the float value of a flag for the given user attributes.
func (e *PgEvaluator) GetFloat(ctx context.Context, flagKey string, userAttrs map[string]string) float64 {
	ff, err := e.repo.FindByKey(ctx, flagKey)
	if err != nil {
		return 0
	}
	val, err := strconv.ParseFloat(ff.Evaluate(userAttrs), 64)
	if err != nil {
		return 0
	}
	return val
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/featureflag/infrastructure/postgres/evaluator.go
git commit -m "feat(featureflag): add PostgreSQL-backed Evaluator"
```

---

### Task 13: In-Memory Cached Evaluator

**Files:**
- Create: `internal/featureflag/infrastructure/cache/evaluator_cache.go`

- [ ] **Step 1: Implement cached evaluator**

File: `internal/featureflag/infrastructure/cache/evaluator_cache.go`

```go
package cache

import (
	"context"
	"strconv"
	"sync"

	"gct/internal/featureflag/domain"
	"gct/internal/shared/infrastructure/logger"
)

// CachedEvaluator wraps a FeatureFlagRepository with an in-memory cache.
// Cache is loaded at startup and invalidated via eventBus events.
type CachedEvaluator struct {
	repo  domain.FeatureFlagRepository
	cache sync.Map // key string -> *domain.FeatureFlag
	log   logger.Log
}

// NewCachedEvaluator creates a cached evaluator and loads all flags into memory.
func NewCachedEvaluator(ctx context.Context, repo domain.FeatureFlagRepository, log logger.Log) (*CachedEvaluator, error) {
	ce := &CachedEvaluator{repo: repo, log: log}
	if err := ce.LoadAll(ctx); err != nil {
		return nil, err
	}
	return ce, nil
}

// LoadAll loads all feature flags from the database into the cache.
func (ce *CachedEvaluator) LoadAll(ctx context.Context) error {
	flags, err := ce.repo.FindAll(ctx)
	if err != nil {
		return err
	}

	// Clear existing cache
	ce.cache.Range(func(key, _ any) bool {
		ce.cache.Delete(key)
		return true
	})

	for _, ff := range flags {
		ce.cache.Store(ff.Key(), ff)
	}

	ce.log.Infow("feature flag cache loaded", "count", len(flags))
	return nil
}

// Invalidate reloads the entire cache. Called when a flag is created/updated/deleted.
func (ce *CachedEvaluator) Invalidate(ctx context.Context) {
	if err := ce.LoadAll(ctx); err != nil {
		ce.log.Errorw("failed to reload feature flag cache", "error", err)
	}
}

// IsEnabled checks if a boolean flag is enabled.
func (ce *CachedEvaluator) IsEnabled(ctx context.Context, flagKey string, userAttrs map[string]string) bool {
	ff := ce.getFlag(ctx, flagKey)
	if ff == nil {
		return false
	}
	return ff.Evaluate(userAttrs) == "true"
}

// GetString returns the string value of a flag.
func (ce *CachedEvaluator) GetString(ctx context.Context, flagKey string, userAttrs map[string]string) string {
	ff := ce.getFlag(ctx, flagKey)
	if ff == nil {
		return ""
	}
	return ff.Evaluate(userAttrs)
}

// GetInt returns the int value of a flag.
func (ce *CachedEvaluator) GetInt(ctx context.Context, flagKey string, userAttrs map[string]string) int {
	ff := ce.getFlag(ctx, flagKey)
	if ff == nil {
		return 0
	}
	val, err := strconv.Atoi(ff.Evaluate(userAttrs))
	if err != nil {
		return 0
	}
	return val
}

// GetFloat returns the float value of a flag.
func (ce *CachedEvaluator) GetFloat(ctx context.Context, flagKey string, userAttrs map[string]string) float64 {
	ff := ce.getFlag(ctx, flagKey)
	if ff == nil {
		return 0
	}
	val, err := strconv.ParseFloat(ff.Evaluate(userAttrs), 64)
	if err != nil {
		return 0
	}
	return val
}

// getFlag retrieves a flag from cache, falling back to DB on miss.
func (ce *CachedEvaluator) getFlag(ctx context.Context, key string) *domain.FeatureFlag {
	if val, ok := ce.cache.Load(key); ok {
		return val.(*domain.FeatureFlag)
	}

	// Cache miss — load from DB
	ff, err := ce.repo.FindByKey(ctx, key)
	if err != nil {
		ce.log.Debugw("feature flag not found", "key", key)
		return nil
	}

	ce.cache.Store(key, ff)
	return ff
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/featureflag/infrastructure/cache/evaluator_cache.go
git commit -m "feat(featureflag): add in-memory cached Evaluator with eventBus invalidation"
```

---

### Task 14: Rewrite Bounded Context Wiring

**Files:**
- Modify: `internal/featureflag/bc.go`

- [ ] **Step 1: Rewrite bc.go**

Replace contents of `internal/featureflag/bc.go`:

```go
package featureflag

import (
	"context"

	"gct/internal/featureflag/application/command"
	"gct/internal/featureflag/application/query"
	"gct/internal/featureflag/domain"
	ffcache "gct/internal/featureflag/infrastructure/cache"
	"gct/internal/featureflag/infrastructure/postgres"
	"gct/internal/shared/application"
	shareddomain "gct/internal/shared/domain"
	"gct/internal/shared/infrastructure/logger"

	"github.com/jackc/pgx/v5/pgxpool"
)

// BoundedContext wires together all command and query handlers for the FeatureFlag BC.
type BoundedContext struct {
	// Commands
	CreateFlag      *command.CreateHandler
	UpdateFlag      *command.UpdateHandler
	DeleteFlag      *command.DeleteHandler
	CreateRuleGroup *command.CreateRuleGroupHandler
	UpdateRuleGroup *command.UpdateRuleGroupHandler
	DeleteRuleGroup *command.DeleteRuleGroupHandler

	// Queries
	GetFlag   *query.GetHandler
	ListFlags *query.ListHandler

	// Runtime evaluator (cached)
	Evaluator domain.Evaluator
}

// NewBoundedContext creates a fully wired FeatureFlag bounded context.
func NewBoundedContext(ctx context.Context, pool *pgxpool.Pool, eventBus application.EventBus, l logger.Log) (*BoundedContext, error) {
	writeRepo := postgres.NewFeatureFlagWriteRepo(pool)
	readRepo := postgres.NewFeatureFlagReadRepo(pool)
	rgRepo := postgres.NewRuleGroupWriteRepo(pool)

	// Initialize cached evaluator
	cachedEval, err := ffcache.NewCachedEvaluator(ctx, writeRepo, l)
	if err != nil {
		return nil, err
	}

	// Subscribe to flag events for cache invalidation
	invalidate := func(_ context.Context, _ shareddomain.DomainEvent) error {
		cachedEval.Invalidate(context.Background())
		return nil
	}
	_ = eventBus.Subscribe("featureflag.created", invalidate)
	_ = eventBus.Subscribe("featureflag.updated", invalidate)
	_ = eventBus.Subscribe("featureflag.deleted", invalidate)
	_ = eventBus.Subscribe("featureflag.toggled", invalidate)

	return &BoundedContext{
		CreateFlag:      command.NewCreateHandler(writeRepo, eventBus, l),
		UpdateFlag:      command.NewUpdateHandler(writeRepo, eventBus, l),
		DeleteFlag:      command.NewDeleteHandler(writeRepo, eventBus, l),
		CreateRuleGroup: command.NewCreateRuleGroupHandler(writeRepo, rgRepo, eventBus, l),
		UpdateRuleGroup: command.NewUpdateRuleGroupHandler(rgRepo, eventBus, l),
		DeleteRuleGroup: command.NewDeleteRuleGroupHandler(rgRepo, eventBus, l),
		GetFlag:         query.NewGetHandler(readRepo),
		ListFlags:       query.NewListHandler(readRepo),
		Evaluator:       cachedEval,
	}, nil
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/featureflag/bc.go
git commit -m "feat(featureflag): rewrite BC wiring with rule group handlers and cached evaluator"
```

---

### Task 15: Update App Bootstrap

**Files:**
- Modify: `internal/app/ddd_bootstrap.go`

- [ ] **Step 1: Update NewDDDBoundedContexts**

The `featureflag.NewBoundedContext` now requires `ctx` and returns `(*BoundedContext, error)`. Update the bootstrap:

In `internal/app/ddd_bootstrap.go`, change the `NewDDDBoundedContexts` signature to accept `ctx context.Context` and return `(*DDDBoundedContexts, error)`:

```go
func NewDDDBoundedContexts(ctx context.Context, pool *pgxpool.Pool, eventBus application.EventBus, l logger.Log, jwtCfg command.JWTConfig) (*DDDBoundedContexts, error) {
```

Change the `FeatureFlag` line from:
```go
FeatureFlag:  featureflag.NewBoundedContext(pool, eventBus, l),
```
to create it separately:
```go
	ffBC, err := featureflag.NewBoundedContext(ctx, pool, eventBus, l)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize feature flag BC: %w", err)
	}
```

Then set `FeatureFlag: ffBC` in the struct literal. Add `"context"` and `"fmt"` to imports.

Update the return to `return &DDDBoundedContexts{...}, nil`.

- [ ] **Step 2: Update caller in app.go**

Find where `NewDDDBoundedContexts` is called in `internal/app/app.go` and pass `ctx` as the first argument. Handle the returned error.

- [ ] **Step 3: Verify compile**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go build ./internal/app/`
Expected: May fail if app.go call site needs updating — fix accordingly

- [ ] **Step 4: Commit**

```bash
git add internal/app/ddd_bootstrap.go internal/app/app.go
git commit -m "feat(featureflag): update app bootstrap for new BC signature"
```

---

### Task 16: Update HTTP Handler and Routes

**Files:**
- Modify: `internal/featureflag/interfaces/http/handler.go`
- Modify: `internal/featureflag/interfaces/http/request.go`
- Modify: `internal/featureflag/interfaces/http/routes.go`

- [ ] **Step 1: Update request.go**

Replace contents of `internal/featureflag/interfaces/http/request.go`:

```go
package http

// CreateRequest represents the request body for creating a feature flag.
type CreateRequest struct {
	Name              string `json:"name" binding:"required"`
	Key               string `json:"key" binding:"required"`
	Description       string `json:"description"`
	FlagType          string `json:"flag_type" binding:"required,oneof=bool string int float"`
	DefaultValue      string `json:"default_value"`
	RolloutPercentage int    `json:"rollout_percentage"`
	IsActive          bool   `json:"is_active"`
}

// UpdateRequest represents the request body for updating a feature flag.
type UpdateRequest struct {
	Name              *string `json:"name,omitempty"`
	Key               *string `json:"key,omitempty"`
	Description       *string `json:"description,omitempty"`
	FlagType          *string `json:"flag_type,omitempty"`
	DefaultValue      *string `json:"default_value,omitempty"`
	RolloutPercentage *int    `json:"rollout_percentage,omitempty"`
	IsActive          *bool   `json:"is_active,omitempty"`
}

// ConditionRequest represents a condition in a rule group.
type ConditionRequest struct {
	Attribute string `json:"attribute" binding:"required"`
	Operator  string `json:"operator" binding:"required"`
	Value     string `json:"value" binding:"required"`
}

// CreateRuleGroupRequest represents the request body for creating a rule group.
type CreateRuleGroupRequest struct {
	Name       string             `json:"name" binding:"required"`
	Variation  string             `json:"variation" binding:"required"`
	Priority   int                `json:"priority"`
	Conditions []ConditionRequest `json:"conditions" binding:"required,min=1"`
}

// UpdateRuleGroupRequest represents the request body for updating a rule group.
type UpdateRuleGroupRequest struct {
	Name       *string             `json:"name,omitempty"`
	Variation  *string             `json:"variation,omitempty"`
	Priority   *int                `json:"priority,omitempty"`
	Conditions *[]ConditionRequest `json:"conditions,omitempty"`
}
```

- [ ] **Step 2: Update handler.go**

Replace contents of `internal/featureflag/interfaces/http/handler.go`:

```go
package http

import (
	"net/http"
	"strconv"

	"gct/internal/featureflag"
	"gct/internal/featureflag/application/command"
	"gct/internal/featureflag/application/query"
	"gct/internal/featureflag/domain"
	"gct/internal/shared/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler provides HTTP endpoints for the FeatureFlag bounded context.
type Handler struct {
	bc *featureflag.BoundedContext
	l  logger.Log
}

// NewHandler creates a new FeatureFlag HTTP handler.
func NewHandler(bc *featureflag.BoundedContext, l logger.Log) *Handler {
	return &Handler{bc: bc, l: l}
}

// Create creates a new feature flag.
func (h *Handler) Create(ctx *gin.Context) {
	var req CreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cmd := command.CreateCommand{
		Name:              req.Name,
		Key:               req.Key,
		Description:       req.Description,
		FlagType:          req.FlagType,
		DefaultValue:      req.DefaultValue,
		RolloutPercentage: req.RolloutPercentage,
		IsActive:          req.IsActive,
	}
	if err := h.bc.CreateFlag.Handle(ctx.Request.Context(), cmd); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"success": true})
}

// List returns a paginated list of feature flags.
func (h *Handler) List(ctx *gin.Context) {
	limit, _ := strconv.ParseInt(ctx.DefaultQuery("limit", "10"), 10, 64)
	offset, _ := strconv.ParseInt(ctx.DefaultQuery("offset", "0"), 10, 64)

	q := query.ListQuery{
		Filter: domain.FeatureFlagFilter{Limit: limit, Offset: offset},
	}
	result, err := h.bc.ListFlags.Handle(ctx.Request.Context(), q)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result.Flags, "total": result.Total})
}

// Get returns a single feature flag by ID.
func (h *Handler) Get(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	result, err := h.bc.GetFlag.Handle(ctx.Request.Context(), query.GetQuery{ID: id})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result})
}

// Update updates a feature flag.
func (h *Handler) Update(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req UpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cmd := command.UpdateCommand{
		ID:                id,
		Name:              req.Name,
		Key:               req.Key,
		Description:       req.Description,
		FlagType:          req.FlagType,
		DefaultValue:      req.DefaultValue,
		RolloutPercentage: req.RolloutPercentage,
		IsActive:          req.IsActive,
	}
	if err := h.bc.UpdateFlag.Handle(ctx.Request.Context(), cmd); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

// Delete deletes a feature flag.
func (h *Handler) Delete(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.bc.DeleteFlag.Handle(ctx.Request.Context(), command.DeleteCommand{ID: id}); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

// CreateRuleGroup adds a rule group to a feature flag.
func (h *Handler) CreateRuleGroup(ctx *gin.Context) {
	flagID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid flag id"})
		return
	}
	var req CreateRuleGroupRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conditions := make([]command.ConditionInput, len(req.Conditions))
	for i, c := range req.Conditions {
		conditions[i] = command.ConditionInput{
			Attribute: c.Attribute,
			Operator:  c.Operator,
			Value:     c.Value,
		}
	}

	cmd := command.CreateRuleGroupCommand{
		FlagID:     flagID,
		Name:       req.Name,
		Variation:  req.Variation,
		Priority:   req.Priority,
		Conditions: conditions,
	}
	if err := h.bc.CreateRuleGroup.Handle(ctx.Request.Context(), cmd); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"success": true})
}

// UpdateRuleGroup updates a rule group.
func (h *Handler) UpdateRuleGroup(ctx *gin.Context) {
	groupID, err := uuid.Parse(ctx.Param("groupId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}
	var req UpdateRuleGroupRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := command.UpdateRuleGroupCommand{
		ID:        groupID,
		Name:      req.Name,
		Variation: req.Variation,
		Priority:  req.Priority,
	}

	if req.Conditions != nil {
		conditions := make([]command.ConditionInput, len(*req.Conditions))
		for i, c := range *req.Conditions {
			conditions[i] = command.ConditionInput{
				Attribute: c.Attribute,
				Operator:  c.Operator,
				Value:     c.Value,
			}
		}
		cmd.Conditions = &conditions
	}

	if err := h.bc.UpdateRuleGroup.Handle(ctx.Request.Context(), cmd); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

// DeleteRuleGroup deletes a rule group.
func (h *Handler) DeleteRuleGroup(ctx *gin.Context) {
	groupID, err := uuid.Parse(ctx.Param("groupId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}
	if err := h.bc.DeleteRuleGroup.Handle(ctx.Request.Context(), command.DeleteRuleGroupCommand{ID: groupID}); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"success": true})
}
```

- [ ] **Step 3: Update routes.go**

Replace contents of `internal/featureflag/interfaces/http/routes.go`:

```go
package http

import "github.com/gin-gonic/gin"

// RegisterRoutes registers all FeatureFlag HTTP routes on the given router group.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/feature-flags")
	g.POST("", h.Create)
	g.GET("", h.List)
	g.GET("/:id", h.Get)
	g.PATCH("/:id", h.Update)
	g.DELETE("/:id", h.Delete)

	// Rule group routes
	g.POST("/:id/rule-groups", h.CreateRuleGroup)
	g.PATCH("/:id/rule-groups/:groupId", h.UpdateRuleGroup)
	g.DELETE("/:id/rule-groups/:groupId", h.DeleteRuleGroup)
}
```

- [ ] **Step 4: Update ddd_routes.go**

In `internal/app/ddd_routes.go`, find the feature flags route registration block and replace it with:

```go
	ffHandler := featureflaghttp.NewHandler(bcs.FeatureFlag, l)
	ffHandler.RegisterRoutes(protected)
```

Remove the inline route definitions for feature flags.

- [ ] **Step 5: Verify compile**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go build ./...`
Expected: May have compile errors — fix any remaining references

- [ ] **Step 6: Commit**

```bash
git add internal/featureflag/interfaces/http/ internal/app/ddd_routes.go
git commit -m "feat(featureflag): update HTTP handler with rule group endpoints"
```

---

### Task 17: Delete go-feature-flag Infrastructure

**Files:**
- Delete: entire `internal/shared/infrastructure/featureflag/` directory
- Delete: `config/flags.yaml`
- Delete: `config/featureflag.go`

- [ ] **Step 1: Remove the shared featureflag directory**

```bash
rm -rf "internal/shared/infrastructure/featureflag/"
```

- [ ] **Step 2: Remove config files**

```bash
rm -f config/flags.yaml config/featureflag.go
```

- [ ] **Step 3: Remove go-feature-flag from go.mod**

```bash
cd "/Users/mrb/Desktop/Golang Template/Backend" && go mod tidy
```

- [ ] **Step 4: Remove any remaining references**

Search for any remaining `go-feature-flag` or `config.FeatureFlag` references and remove them. Check `config/config.go` for a `FeatureFlag` field.

- [ ] **Step 5: Verify compile**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go build ./...`
Expected: Success

- [ ] **Step 6: Commit**

```bash
git add -A
git commit -m "refactor(featureflag): remove go-feature-flag library, YAML config, and shared infrastructure"
```

---

### Task 18: Delete Old Test Files and Stale Command Tests

**Files:**
- Delete: `internal/featureflag/application/command/error_paths_test.go`
- Delete: `internal/featureflag/application/command/create_test.go`
- Delete: `internal/featureflag/application/command/update_test.go`
- Delete: `internal/featureflag/application/query/get_test.go`
- Delete: `internal/featureflag/application/query/list_test.go`
- Delete: `internal/featureflag/interfaces/http/handler_test.go`

These tests reference the old entity signatures and must be rewritten. Remove them now; integration tests in Task 19 will cover the full flow.

- [ ] **Step 1: Remove stale test files**

```bash
rm -f internal/featureflag/application/command/error_paths_test.go
rm -f internal/featureflag/application/command/create_test.go
rm -f internal/featureflag/application/command/update_test.go
rm -f internal/featureflag/application/query/get_test.go
rm -f internal/featureflag/application/query/list_test.go
rm -f internal/featureflag/interfaces/http/handler_test.go
```

- [ ] **Step 2: Verify compile and domain tests still pass**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go test ./internal/featureflag/domain/ -v`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add -A
git commit -m "chore(featureflag): remove stale test files for old entity signatures"
```

---

### Task 19: Integration Tests

**Files:**
- Modify: `test/integration/featureflag/setup_test.go`
- Modify: `test/integration/featureflag/integration_test.go`

- [ ] **Step 1: Update setup_test.go**

Replace contents of `test/integration/featureflag/setup_test.go`:

```go
package featureflag

import (
	"testing"

	"gct/test/integration/common/setup"
)

func TestMain(m *testing.M) {
	setup.SetupTestEnvironment(m)
}

func cleanDB(t *testing.T) {
	t.Helper()
	setup.CleanDB(t)
	ctx := t.Context()
	_, err := setup.TestPG.Pool.Exec(ctx, `DELETE FROM feature_flag_conditions`)
	if err != nil {
		t.Fatalf("cleanDB feature_flag_conditions error: %s", err)
	}
	_, err = setup.TestPG.Pool.Exec(ctx, `DELETE FROM feature_flag_rule_groups`)
	if err != nil {
		t.Fatalf("cleanDB feature_flag_rule_groups error: %s", err)
	}
	_, err = setup.TestPG.Pool.Exec(ctx, `DELETE FROM feature_flags`)
	if err != nil {
		t.Fatalf("cleanDB feature_flags error: %s", err)
	}
}
```

- [ ] **Step 2: Rewrite integration_test.go**

Replace contents of `test/integration/featureflag/integration_test.go`:

```go
package featureflag

import (
	"context"
	"testing"

	"gct/internal/featureflag"
	"gct/internal/featureflag/application/command"
	"gct/internal/featureflag/application/query"
	"gct/internal/featureflag/domain"
	"gct/internal/shared/infrastructure/eventbus"
	"gct/internal/shared/infrastructure/logger"
	"gct/test/integration/common/setup"
)

func newTestBC(t *testing.T) *featureflag.BoundedContext {
	t.Helper()
	eb := eventbus.NewInMemoryEventBus()
	l := logger.New("error")
	bc, err := featureflag.NewBoundedContext(context.Background(), setup.TestPG.Pool, eb, l)
	if err != nil {
		t.Fatalf("NewBoundedContext: %v", err)
	}
	return bc
}

func TestIntegration_CreateAndGetFeatureFlag(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	err := bc.CreateFlag.Handle(ctx, command.CreateCommand{
		Name:              "Dark Mode",
		Key:               "dark_mode",
		Description:       "Enable dark mode UI",
		FlagType:          "bool",
		DefaultValue:      "false",
		RolloutPercentage: 0,
		IsActive:          true,
	})
	if err != nil {
		t.Fatalf("CreateFlag: %v", err)
	}

	result, err := bc.ListFlags.Handle(ctx, query.ListQuery{
		Filter: domain.FeatureFlagFilter{Limit: 10},
	})
	if err != nil {
		t.Fatalf("ListFlags: %v", err)
	}
	if result.Total != 1 {
		t.Fatalf("expected 1 flag, got %d", result.Total)
	}

	f := result.Flags[0]
	if f.Key != "dark_mode" {
		t.Errorf("expected key dark_mode, got %s", f.Key)
	}
	if !f.IsActive {
		t.Error("expected flag to be active")
	}
	if f.FlagType != "bool" {
		t.Errorf("expected flag_type bool, got %s", f.FlagType)
	}
}

func TestIntegration_UpdateFeatureFlag(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	err := bc.CreateFlag.Handle(ctx, command.CreateCommand{
		Name:     "Checkout",
		Key:      "new_checkout",
		FlagType: "bool",
		DefaultValue: "false",
	})
	if err != nil {
		t.Fatalf("CreateFlag: %v", err)
	}

	list, _ := bc.ListFlags.Handle(ctx, query.ListQuery{Filter: domain.FeatureFlagFilter{Limit: 10}})
	fID := list.Flags[0].ID

	newName := "Updated Checkout"
	newActive := true
	newRollout := 75
	err = bc.UpdateFlag.Handle(ctx, command.UpdateCommand{
		ID:                fID,
		Name:              &newName,
		IsActive:          &newActive,
		RolloutPercentage: &newRollout,
	})
	if err != nil {
		t.Fatalf("UpdateFlag: %v", err)
	}

	view, _ := bc.GetFlag.Handle(ctx, query.GetQuery{ID: fID})
	if view.Name != "Updated Checkout" {
		t.Errorf("name not updated, got %s", view.Name)
	}
	if !view.IsActive {
		t.Error("flag should be active after update")
	}
	if view.RolloutPercentage != 75 {
		t.Errorf("expected rollout 75, got %d", view.RolloutPercentage)
	}
}

func TestIntegration_DeleteFeatureFlag(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	err := bc.CreateFlag.Handle(ctx, command.CreateCommand{
		Name: "To Delete", Key: "to_delete", FlagType: "bool", DefaultValue: "false",
	})
	if err != nil {
		t.Fatalf("CreateFlag: %v", err)
	}

	list, _ := bc.ListFlags.Handle(ctx, query.ListQuery{Filter: domain.FeatureFlagFilter{Limit: 10}})
	fID := list.Flags[0].ID

	err = bc.DeleteFlag.Handle(ctx, command.DeleteCommand{ID: fID})
	if err != nil {
		t.Fatalf("DeleteFlag: %v", err)
	}

	list2, _ := bc.ListFlags.Handle(ctx, query.ListQuery{Filter: domain.FeatureFlagFilter{Limit: 10}})
	if list2.Total != 0 {
		t.Errorf("expected 0 flags after delete, got %d", list2.Total)
	}
}

func TestIntegration_RuleGroupCRUD(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	// Create flag
	err := bc.CreateFlag.Handle(ctx, command.CreateCommand{
		Name: "Admin Feature", Key: "admin_feature", FlagType: "bool",
		DefaultValue: "false", IsActive: true,
	})
	if err != nil {
		t.Fatalf("CreateFlag: %v", err)
	}

	list, _ := bc.ListFlags.Handle(ctx, query.ListQuery{Filter: domain.FeatureFlagFilter{Limit: 10}})
	flagID := list.Flags[0].ID

	// Create rule group
	err = bc.CreateRuleGroup.Handle(ctx, command.CreateRuleGroupCommand{
		FlagID:    flagID,
		Name:      "Admin users",
		Variation: "true",
		Priority:  1,
		Conditions: []command.ConditionInput{
			{Attribute: "role", Operator: "eq", Value: "admin"},
			{Attribute: "country", Operator: "in", Value: "US,UK,UZ"},
		},
	})
	if err != nil {
		t.Fatalf("CreateRuleGroup: %v", err)
	}

	// Verify via Get
	view, err := bc.GetFlag.Handle(ctx, query.GetQuery{ID: flagID})
	if err != nil {
		t.Fatalf("GetFlag: %v", err)
	}
	if len(view.RuleGroups) != 1 {
		t.Fatalf("expected 1 rule group, got %d", len(view.RuleGroups))
	}
	rg := view.RuleGroups[0]
	if rg.Name != "Admin users" {
		t.Errorf("expected name 'Admin users', got %s", rg.Name)
	}
	if len(rg.Conditions) != 2 {
		t.Fatalf("expected 2 conditions, got %d", len(rg.Conditions))
	}

	// Delete rule group
	err = bc.DeleteRuleGroup.Handle(ctx, command.DeleteRuleGroupCommand{ID: rg.ID})
	if err != nil {
		t.Fatalf("DeleteRuleGroup: %v", err)
	}

	view2, _ := bc.GetFlag.Handle(ctx, query.GetQuery{ID: flagID})
	if len(view2.RuleGroups) != 0 {
		t.Errorf("expected 0 rule groups after delete, got %d", len(view2.RuleGroups))
	}
}

func TestIntegration_Evaluator(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	// Create flag with rule group
	err := bc.CreateFlag.Handle(ctx, command.CreateCommand{
		Name: "Premium Feature", Key: "premium_feature", FlagType: "bool",
		DefaultValue: "false", IsActive: true,
	})
	if err != nil {
		t.Fatalf("CreateFlag: %v", err)
	}

	list, _ := bc.ListFlags.Handle(ctx, query.ListQuery{Filter: domain.FeatureFlagFilter{Limit: 10}})
	flagID := list.Flags[0].ID

	err = bc.CreateRuleGroup.Handle(ctx, command.CreateRuleGroupCommand{
		FlagID:    flagID,
		Name:      "Premium users",
		Variation: "true",
		Priority:  1,
		Conditions: []command.ConditionInput{
			{Attribute: "plan", Operator: "eq", Value: "premium"},
		},
	})
	if err != nil {
		t.Fatalf("CreateRuleGroup: %v", err)
	}

	// Test evaluator — premium user
	if !bc.Evaluator.IsEnabled(ctx, "premium_feature", map[string]string{"user_id": "u1", "plan": "premium"}) {
		t.Error("expected premium_feature enabled for premium user")
	}

	// Test evaluator — free user
	if bc.Evaluator.IsEnabled(ctx, "premium_feature", map[string]string{"user_id": "u2", "plan": "free"}) {
		t.Error("expected premium_feature disabled for free user")
	}

	// Test evaluator — nonexistent flag
	if bc.Evaluator.IsEnabled(ctx, "nonexistent", map[string]string{"user_id": "u1"}) {
		t.Error("expected false for nonexistent flag")
	}
}
```

- [ ] **Step 3: Run integration tests**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go test ./test/integration/featureflag/ -v -count=1`
Expected: PASS — all 5 tests green

- [ ] **Step 4: Commit**

```bash
git add test/integration/featureflag/
git commit -m "test(featureflag): rewrite integration tests for consolidated feature flag system"
```

---

### Task 20: Full Build and Test Verification

- [ ] **Step 1: Run full build**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go build ./...`
Expected: Success — zero errors

- [ ] **Step 2: Run all domain tests**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go test ./internal/featureflag/... -v`
Expected: PASS

- [ ] **Step 3: Run go vet**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go vet ./...`
Expected: No issues

- [ ] **Step 4: Fix any issues found and commit**

If any issues found, fix them and commit:

```bash
git add -A
git commit -m "fix(featureflag): resolve build/vet issues from consolidation"
```
