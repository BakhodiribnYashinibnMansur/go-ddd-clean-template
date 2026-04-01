# ABAC Policy Evaluation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Activate ABAC policy evaluation so that after RBAC passes, policies bound to matched permissions are evaluated against request context (user attributes, IP, time, etc.)

**Architecture:** Pure Go evaluator as domain service. Operator registry pattern for extensibility. CheckAccess enhanced with ABAC step after RBAC. Middleware builds EvaluationContext from request.

**Tech Stack:** Go, Gin, pgx, squirrel (no new dependencies)

**Spec:** `docs/superpowers/specs/2026-04-01-abac-policy-evaluation-design.md`

---

## File Structure

| Action | File | Responsibility |
|--------|------|---------------|
| Create | `internal/authz/domain/abac.go` | EvaluationContext, key parser, operator registry, PolicyEvaluator |
| Create | `internal/authz/domain/abac_test.go` | Unit tests for all ABAC logic |
| Modify | `internal/authz/domain/repository.go:102-110` | Add `evalCtx` to `CheckAccess` + add `FindPoliciesByPermissionIDs` |
| Modify | `internal/authz/application/query/check_access.go` | Add `EvalCtx` to query struct, pass through |
| Modify | `internal/authz/infrastructure/postgres/read_repo.go:290-349` | Collect permission IDs, fetch policies, evaluate |
| Modify | `internal/authz/interfaces/http/middleware/authz.go:48-111` | Build EvaluationContext from request |
| Modify | `internal/authz/application/query/get_role_test.go:18-75` | Update mock to match new interface |
| Modify | `internal/authz/application/query/check_access_test.go` | Update tests for new signature |
| Modify | `internal/authz/interfaces/http/handler_test.go` | Update mock to match new interface |
| Modify | `internal/authz/interfaces/http/middleware/authz_test.go` | Update tests for EvalCtx |
| Modify | `internal/authz/infrastructure/postgres/read_repo_test.go` | Existing matchScope tests stay, no change needed |
| Create | `test/integration/authz/access/abac_test.go` | Integration tests for ABAC flow |

---

### Task 1: EvaluationContext and Key Parser

**Files:**
- Create: `internal/authz/domain/abac.go`
- Create: `internal/authz/domain/abac_test.go`

- [ ] **Step 1: Write failing tests for EvaluationContext.Resolve and parseConditionKey**

```go
// internal/authz/domain/abac_test.go
package domain

import (
	"testing"
)

func TestEvaluationContext_Resolve(t *testing.T) {
	ctx := EvaluationContext{
		Attrs: map[string]map[string]any{
			"user": {"role_name": "admin", "id": "123"},
			"env":  {"ip": "10.0.0.1"},
		},
	}

	tests := []struct {
		key      string
		wantVal  any
		wantOK   bool
	}{
		{"user.role_name", "admin", true},
		{"user.id", "123", true},
		{"env.ip", "10.0.0.1", true},
		{"user.missing", nil, false},
		{"unknown.field", nil, false},
		{"invalid", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			val, ok := ctx.Resolve(tt.key)
			if ok != tt.wantOK {
				t.Errorf("Resolve(%q) ok = %v, want %v", tt.key, ok, tt.wantOK)
			}
			if val != tt.wantVal {
				t.Errorf("Resolve(%q) = %v, want %v", tt.key, val, tt.wantVal)
			}
		})
	}
}

func TestEvaluationContext_Resolve_TargetRef(t *testing.T) {
	ctx := EvaluationContext{
		Attrs: map[string]map[string]any{
			"target": {"user.relation_names": []string{"dept_a", "dept_b"}},
		},
	}

	val, ok := ctx.Resolve("target.user.relation_names")
	if !ok {
		t.Fatal("expected ok=true for target ref")
	}
	names, _ := val.([]string)
	if len(names) != 2 {
		t.Errorf("expected 2 names, got %d", len(names))
	}
}

func TestParseConditionKey(t *testing.T) {
	tests := []struct {
		key    string
		wantNS string
		wantF  string
		wantOp string
	}{
		{"user.role_name", "user", "role_name", "equals"},
		{"env.ip_not_in", "env", "ip", "not_in"},
		{"user.relation_names_any", "user", "relation_names", "any"},
		{"env.time_between", "env", "time", "between"},
		{"env.time_gt", "env", "time", "gt"},
		{"user.role_name_not_equals", "user", "role_name", "not_equals"},
		{"user.some_field_contains", "user", "some_field", "contains"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			ns, field, op := parseConditionKey(tt.key)
			if ns != tt.wantNS {
				t.Errorf("namespace = %q, want %q", ns, tt.wantNS)
			}
			if field != tt.wantF {
				t.Errorf("field = %q, want %q", field, tt.wantF)
			}
			if op != tt.wantOp {
				t.Errorf("operator = %q, want %q", op, tt.wantOp)
			}
		})
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go test ./internal/authz/domain/... -run "TestEvaluationContext|TestParseConditionKey" -v`
Expected: FAIL — types and functions not defined

- [ ] **Step 3: Implement EvaluationContext and parseConditionKey**

```go
// internal/authz/domain/abac.go
package domain

import (
	"strings"
)

// EvaluationContext carries all attributes available for ABAC policy evaluation.
// Attrs keys are namespaces: "user", "env", "resource", "target".
type EvaluationContext struct {
	Attrs map[string]map[string]any
}

// Resolve looks up a dotted key like "user.role_name" from the context.
// For target references the key format is "target.user.relation_names" which
// maps to Attrs["target"]["user.relation_names"].
func (ec EvaluationContext) Resolve(key string) (any, bool) {
	dotIdx := strings.Index(key, ".")
	if dotIdx < 0 {
		return nil, false
	}

	ns := key[:dotIdx]
	field := key[dotIdx+1:]

	nsMap, ok := ec.Attrs[ns]
	if !ok {
		return nil, false
	}

	val, ok := nsMap[field]
	return val, ok
}

// OperatorFunc evaluates a condition: fieldVal is the resolved attribute,
// conditionVal is the value from the policy condition map.
type OperatorFunc func(fieldVal, conditionVal any) bool

// operators is the global registry of condition operators.
// To add a new operator, register a function here.
var operators = map[string]OperatorFunc{}

// compound operators that are two words joined by underscore.
var compoundOperators = []string{"not_equals", "not_in"}

// parseConditionKey splits "user.role_name_not_in" into (namespace, field, operator).
// It scans from the right for known operator suffixes. Default operator is "equals".
func parseConditionKey(key string) (namespace, field, operator string) {
	dotIdx := strings.Index(key, ".")
	if dotIdx < 0 {
		return key, "", "equals"
	}

	namespace = key[:dotIdx]
	rest := key[dotIdx+1:]

	// Try compound operators first (2-word: not_in, not_equals).
	for _, cop := range compoundOperators {
		suffix := "_" + cop
		if strings.HasSuffix(rest, suffix) {
			field = strings.TrimSuffix(rest, suffix)
			return namespace, field, cop
		}
	}

	// Try single-word operators.
	lastUnderscore := strings.LastIndex(rest, "_")
	if lastUnderscore >= 0 {
		candidate := rest[lastUnderscore+1:]
		if _, ok := operators[candidate]; ok {
			return namespace, rest[:lastUnderscore], candidate
		}
	}

	return namespace, rest, "equals"
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go test ./internal/authz/domain/... -run "TestEvaluationContext|TestParseConditionKey" -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/authz/domain/abac.go internal/authz/domain/abac_test.go
git commit -m "feat(authz): add EvaluationContext and condition key parser for ABAC"
```

---

### Task 2: Operator Registry — All 12 Operators

**Files:**
- Modify: `internal/authz/domain/abac.go`
- Modify: `internal/authz/domain/abac_test.go`

- [ ] **Step 1: Write failing tests for all operators**

Append to `internal/authz/domain/abac_test.go`:

```go
func TestOpEquals(t *testing.T) {
	tests := []struct {
		name  string
		field any
		cond  any
		want  bool
	}{
		{"string match", "admin", "admin", true},
		{"string mismatch", "admin", "user", false},
		{"int match", 42, 42, true},
		{"float match", 3.14, 3.14, true},
		{"nil field", nil, "admin", false},
		{"nil both", nil, nil, true},
		{"float64 vs int", float64(42), 42, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := opEquals(tt.field, tt.cond); got != tt.want {
				t.Errorf("opEquals(%v, %v) = %v, want %v", tt.field, tt.cond, got, tt.want)
			}
		})
	}
}

func TestOpNotEquals(t *testing.T) {
	if !opNotEquals("admin", "user") {
		t.Error("expected true")
	}
	if opNotEquals("admin", "admin") {
		t.Error("expected false")
	}
}

func TestOpIn(t *testing.T) {
	list := []any{"admin", "manager", "editor"}
	if !opIn("admin", list) {
		t.Error("expected admin in list")
	}
	if opIn("guest", list) {
		t.Error("expected guest not in list")
	}
	if opIn("admin", "not-a-list") {
		t.Error("expected false for non-list condition")
	}
}

func TestOpNotIn(t *testing.T) {
	list := []any{"127.0.0.1", "192.168.1.1"}
	if !opNotIn("10.0.0.1", list) {
		t.Error("expected 10.0.0.1 not in list")
	}
	if opNotIn("127.0.0.1", list) {
		t.Error("expected 127.0.0.1 in list")
	}
}

func TestOpContains(t *testing.T) {
	fieldList := []any{"vip", "premium", "beta"}
	if !opContains(fieldList, "vip") {
		t.Error("expected list contains vip")
	}
	if opContains(fieldList, "free") {
		t.Error("expected list does not contain free")
	}
	if opContains("not-a-list", "x") {
		t.Error("expected false for non-list field")
	}
}

func TestOpAny(t *testing.T) {
	field := []any{"dept_a", "dept_b", "dept_c"}
	cond := []any{"dept_b", "dept_x"}
	if !opAny(field, cond) {
		t.Error("expected intersection not empty")
	}
	if opAny(field, []any{"dept_x", "dept_y"}) {
		t.Error("expected no intersection")
	}
}

func TestOpAll(t *testing.T) {
	field := []any{"read", "write", "delete"}
	if !opAll(field, []any{"read", "write"}) {
		t.Error("expected all present")
	}
	if opAll(field, []any{"read", "execute"}) {
		t.Error("expected not all present")
	}
}

func TestOpGt(t *testing.T) {
	if !opGt("10:00", "09:00") {
		t.Error("expected 10:00 > 09:00")
	}
	if opGt("09:00", "10:00") {
		t.Error("expected 09:00 not > 10:00")
	}
	if !opGt(float64(10), float64(5)) {
		t.Error("expected 10 > 5")
	}
}

func TestOpGte(t *testing.T) {
	if !opGte("09:00", "09:00") {
		t.Error("expected 09:00 >= 09:00")
	}
	if !opGte("10:00", "09:00") {
		t.Error("expected 10:00 >= 09:00")
	}
}

func TestOpLt(t *testing.T) {
	if !opLt("09:00", "10:00") {
		t.Error("expected 09:00 < 10:00")
	}
	if opLt("10:00", "09:00") {
		t.Error("expected 10:00 not < 09:00")
	}
}

func TestOpLte(t *testing.T) {
	if !opLte("09:00", "09:00") {
		t.Error("expected 09:00 <= 09:00")
	}
	if !opLte("08:00", "09:00") {
		t.Error("expected 08:00 <= 09:00")
	}
}

func TestOpBetween(t *testing.T) {
	if !opBetween("10:00", []any{"09:00", "18:00"}) {
		t.Error("expected 10:00 between 09:00 and 18:00")
	}
	if opBetween("08:00", []any{"09:00", "18:00"}) {
		t.Error("expected 08:00 not between 09:00 and 18:00")
	}
	if opBetween("19:00", []any{"09:00", "18:00"}) {
		t.Error("expected 19:00 not between 09:00 and 18:00")
	}
	if !opBetween("09:00", []any{"09:00", "18:00"}) {
		t.Error("expected 09:00 at lower bound is between")
	}
	if !opBetween(float64(10), []any{float64(5), float64(15)}) {
		t.Error("expected 10 between 5 and 15")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go test ./internal/authz/domain/... -run "TestOp" -v`
Expected: FAIL — operator functions not defined

- [ ] **Step 3: Implement all 12 operators**

Append to `internal/authz/domain/abac.go`:

```go
import "fmt"

func init() {
	operators["equals"] = opEquals
	operators["not_equals"] = opNotEquals
	operators["in"] = opIn
	operators["not_in"] = opNotIn
	operators["contains"] = opContains
	operators["any"] = opAny
	operators["all"] = opAll
	operators["gt"] = opGt
	operators["gte"] = opGte
	operators["lt"] = opLt
	operators["lte"] = opLte
	operators["between"] = opBetween
}

func opEquals(fieldVal, conditionVal any) bool {
	return fmt.Sprintf("%v", fieldVal) == fmt.Sprintf("%v", conditionVal)
}

func opNotEquals(fieldVal, conditionVal any) bool {
	return !opEquals(fieldVal, conditionVal)
}

func opIn(fieldVal, conditionVal any) bool {
	list, ok := toSlice(conditionVal)
	if !ok {
		return false
	}
	for _, item := range list {
		if opEquals(fieldVal, item) {
			return true
		}
	}
	return false
}

func opNotIn(fieldVal, conditionVal any) bool {
	return !opIn(fieldVal, conditionVal)
}

func opContains(fieldVal, conditionVal any) bool {
	list, ok := toSlice(fieldVal)
	if !ok {
		return false
	}
	for _, item := range list {
		if opEquals(item, conditionVal) {
			return true
		}
	}
	return false
}

func opAny(fieldVal, conditionVal any) bool {
	fieldList, ok1 := toSlice(fieldVal)
	condList, ok2 := toSlice(conditionVal)
	if !ok1 || !ok2 {
		return false
	}
	for _, f := range fieldList {
		for _, c := range condList {
			if opEquals(f, c) {
				return true
			}
		}
	}
	return false
}

func opAll(fieldVal, conditionVal any) bool {
	fieldList, ok1 := toSlice(fieldVal)
	condList, ok2 := toSlice(conditionVal)
	if !ok1 || !ok2 {
		return false
	}
	for _, c := range condList {
		found := false
		for _, f := range fieldList {
			if opEquals(f, c) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func opGt(fieldVal, conditionVal any) bool {
	return compareValues(fieldVal, conditionVal) > 0
}

func opGte(fieldVal, conditionVal any) bool {
	return compareValues(fieldVal, conditionVal) >= 0
}

func opLt(fieldVal, conditionVal any) bool {
	return compareValues(fieldVal, conditionVal) < 0
}

func opLte(fieldVal, conditionVal any) bool {
	return compareValues(fieldVal, conditionVal) <= 0
}

func opBetween(fieldVal, conditionVal any) bool {
	bounds, ok := toSlice(conditionVal)
	if !ok || len(bounds) != 2 {
		return false
	}
	return compareValues(fieldVal, bounds[0]) >= 0 && compareValues(fieldVal, bounds[1]) <= 0
}

// compareValues compares two values as strings. Returns -1, 0, or 1.
func compareValues(a, b any) int {
	sa := fmt.Sprintf("%v", a)
	sb := fmt.Sprintf("%v", b)
	if sa < sb {
		return -1
	}
	if sa > sb {
		return 1
	}
	return 0
}

// toSlice converts any to []any. Handles []any and []string (common from JSON).
func toSlice(v any) ([]any, bool) {
	switch s := v.(type) {
	case []any:
		return s, true
	case []string:
		result := make([]any, len(s))
		for i, item := range s {
			result[i] = item
		}
		return result, true
	default:
		return nil, false
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go test ./internal/authz/domain/... -run "TestOp" -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/authz/domain/abac.go internal/authz/domain/abac_test.go
git commit -m "feat(authz): implement 12 ABAC operators with registry pattern"
```

---

### Task 3: PolicyEvaluator with Dynamic References

**Files:**
- Modify: `internal/authz/domain/abac.go`
- Modify: `internal/authz/domain/abac_test.go`

- [ ] **Step 1: Write failing tests for PolicyEvaluator**

Append to `internal/authz/domain/abac_test.go`:

```go
func TestPolicyEvaluator_SingleAllow(t *testing.T) {
	permID := uuid.New()
	policy := NewPolicy(permID, PolicyAllow)
	policy.SetConditions(map[string]any{"user.role_name": "auditor"})

	ctx := EvaluationContext{
		Attrs: map[string]map[string]any{
			"user": {"role_name": "auditor"},
		},
	}

	ev := &PolicyEvaluator{}
	effect, matched := ev.Evaluate([]*Policy{policy}, ctx)
	if !matched {
		t.Fatal("expected matched")
	}
	if effect != PolicyAllow {
		t.Errorf("expected ALLOW, got %s", effect)
	}
}

func TestPolicyEvaluator_SingleDeny(t *testing.T) {
	permID := uuid.New()
	policy := NewPolicy(permID, PolicyDeny)
	policy.SetConditions(map[string]any{"env.ip_in": []any{"10.0.0.1"}})

	ctx := EvaluationContext{
		Attrs: map[string]map[string]any{
			"env": {"ip": "10.0.0.1"},
		},
	}

	ev := &PolicyEvaluator{}
	effect, matched := ev.Evaluate([]*Policy{policy}, ctx)
	if !matched {
		t.Fatal("expected matched")
	}
	if effect != PolicyDeny {
		t.Errorf("expected DENY, got %s", effect)
	}
}

func TestPolicyEvaluator_DenyWinsOverAllow(t *testing.T) {
	permID := uuid.New()
	allow := NewPolicy(permID, PolicyAllow)
	allow.SetPriority(100)
	allow.SetConditions(map[string]any{"user.role_name": "admin"})

	deny := NewPolicy(permID, PolicyDeny)
	deny.SetPriority(1) // lower priority
	deny.SetConditions(map[string]any{"env.ip_in": []any{"10.0.0.1"}})

	ctx := EvaluationContext{
		Attrs: map[string]map[string]any{
			"user": {"role_name": "admin"},
			"env":  {"ip": "10.0.0.1"},
		},
	}

	ev := &PolicyEvaluator{}
	effect, matched := ev.Evaluate([]*Policy{allow, deny}, ctx)
	if !matched {
		t.Fatal("expected matched")
	}
	if effect != PolicyDeny {
		t.Error("expected DENY to win over ALLOW regardless of priority")
	}
}

func TestPolicyEvaluator_InactivePolicySkipped(t *testing.T) {
	permID := uuid.New()
	policy := NewPolicy(permID, PolicyDeny)
	policy.SetConditions(map[string]any{"user.role_name": "admin"})
	policy.Toggle() // deactivate

	ctx := EvaluationContext{
		Attrs: map[string]map[string]any{
			"user": {"role_name": "admin"},
		},
	}

	ev := &PolicyEvaluator{}
	_, matched := ev.Evaluate([]*Policy{policy}, ctx)
	if matched {
		t.Error("expected inactive policy to be skipped")
	}
}

func TestPolicyEvaluator_NoMatch(t *testing.T) {
	permID := uuid.New()
	policy := NewPolicy(permID, PolicyAllow)
	policy.SetConditions(map[string]any{"user.role_name": "auditor"})

	ctx := EvaluationContext{
		Attrs: map[string]map[string]any{
			"user": {"role_name": "admin"}, // does not match
		},
	}

	ev := &PolicyEvaluator{}
	_, matched := ev.Evaluate([]*Policy{policy}, ctx)
	if matched {
		t.Error("expected no match")
	}
}

func TestPolicyEvaluator_ANDSemantics(t *testing.T) {
	permID := uuid.New()
	policy := NewPolicy(permID, PolicyAllow)
	policy.SetConditions(map[string]any{
		"user.role_name": "admin",
		"env.ip_in":      []any{"10.0.0.1"},
	})

	// Only role matches, IP does not.
	ctx := EvaluationContext{
		Attrs: map[string]map[string]any{
			"user": {"role_name": "admin"},
			"env":  {"ip": "192.168.1.1"},
		},
	}

	ev := &PolicyEvaluator{}
	_, matched := ev.Evaluate([]*Policy{policy}, ctx)
	if matched {
		t.Error("expected no match when only partial conditions match (AND semantics)")
	}
}

func TestPolicyEvaluator_EmptyConditions_AlwaysMatches(t *testing.T) {
	permID := uuid.New()
	policy := NewPolicy(permID, PolicyAllow)
	// empty conditions = matches everything

	ctx := EvaluationContext{
		Attrs: map[string]map[string]any{},
	}

	ev := &PolicyEvaluator{}
	effect, matched := ev.Evaluate([]*Policy{policy}, ctx)
	if !matched {
		t.Fatal("expected empty conditions to match")
	}
	if effect != PolicyAllow {
		t.Errorf("expected ALLOW, got %s", effect)
	}
}

func TestPolicyEvaluator_DynamicReference(t *testing.T) {
	permID := uuid.New()
	policy := NewPolicy(permID, PolicyAllow)
	policy.SetConditions(map[string]any{
		"user.relation_names_any": "$target.user.relation_names",
	})

	ctx := EvaluationContext{
		Attrs: map[string]map[string]any{
			"user":   {"relation_names": []any{"dept_a", "dept_b"}},
			"target": {"user.relation_names": []any{"dept_b", "dept_c"}},
		},
	}

	ev := &PolicyEvaluator{}
	effect, matched := ev.Evaluate([]*Policy{policy}, ctx)
	if !matched {
		t.Fatal("expected dynamic reference to resolve and match")
	}
	if effect != PolicyAllow {
		t.Errorf("expected ALLOW, got %s", effect)
	}
}

func TestPolicyEvaluator_DynamicReference_NoMatch(t *testing.T) {
	permID := uuid.New()
	policy := NewPolicy(permID, PolicyAllow)
	policy.SetConditions(map[string]any{
		"user.relation_names_any": "$target.user.relation_names",
	})

	ctx := EvaluationContext{
		Attrs: map[string]map[string]any{
			"user":   {"relation_names": []any{"dept_a"}},
			"target": {"user.relation_names": []any{"dept_x", "dept_y"}},
		},
	}

	ev := &PolicyEvaluator{}
	_, matched := ev.Evaluate([]*Policy{policy}, ctx)
	if matched {
		t.Error("expected no match when intersection is empty")
	}
}

func TestPolicyEvaluator_NoPolicies(t *testing.T) {
	ev := &PolicyEvaluator{}
	_, matched := ev.Evaluate(nil, EvaluationContext{})
	if matched {
		t.Error("expected no match for empty policy list")
	}
}
```

Add import `"github.com/google/uuid"` to the test file imports.

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go test ./internal/authz/domain/... -run "TestPolicyEvaluator" -v`
Expected: FAIL — PolicyEvaluator not defined

- [ ] **Step 3: Implement PolicyEvaluator**

Append to `internal/authz/domain/abac.go`:

```go
import "sort"

// PolicyEvaluator evaluates ABAC policies against a request context.
type PolicyEvaluator struct{}

// Evaluate runs all active policies against the context.
// Returns the resulting effect and whether any policy matched.
// Explicit DENY always wins over ALLOW regardless of priority.
func (e *PolicyEvaluator) Evaluate(policies []*Policy, ctx EvaluationContext) (PolicyEffect, bool) {
	if len(policies) == 0 {
		return "", false
	}

	// Filter active and sort by priority descending.
	var active []*Policy
	for _, p := range policies {
		if p.IsActive() {
			active = append(active, p)
		}
	}

	if len(active) == 0 {
		return "", false
	}

	sort.Slice(active, func(i, j int) bool {
		return active[i].Priority() > active[j].Priority()
	})

	var hasAllow, hasDeny bool

	for _, p := range active {
		if evaluateConditions(p.Conditions(), ctx) {
			if p.Effect() == PolicyDeny {
				hasDeny = true
			} else {
				hasAllow = true
			}
		}
	}

	if hasDeny {
		return PolicyDeny, true
	}
	if hasAllow {
		return PolicyAllow, true
	}
	return "", false
}

// evaluateConditions checks all conditions in a policy (AND semantics).
// Returns true if ALL conditions match.
func evaluateConditions(conditions map[string]any, ctx EvaluationContext) bool {
	if len(conditions) == 0 {
		return true
	}

	for key, conditionVal := range conditions {
		ns, field, op := parseConditionKey(key)

		// Resolve field value from context.
		fieldVal, _ := ctx.Resolve(ns + "." + field)

		// Resolve dynamic references in condition value.
		conditionVal = resolveDynamicRef(conditionVal, ctx)

		opFunc, ok := operators[op]
		if !ok {
			return false // unknown operator = condition fails
		}

		if !opFunc(fieldVal, conditionVal) {
			return false // AND semantics: first failure = overall failure
		}
	}

	return true
}

// resolveDynamicRef checks if a condition value is a "$target.*" reference
// and resolves it from the EvaluationContext.
func resolveDynamicRef(val any, ctx EvaluationContext) any {
	s, ok := val.(string)
	if !ok || !strings.HasPrefix(s, "$target.") {
		return val
	}
	ref := strings.TrimPrefix(s, "$")
	resolved, ok := ctx.Resolve(ref)
	if !ok {
		return val
	}
	return resolved
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go test ./internal/authz/domain/... -run "TestPolicyEvaluator" -v`
Expected: PASS

- [ ] **Step 5: Run all domain tests**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go test ./internal/authz/domain/... -v`
Expected: ALL PASS

- [ ] **Step 6: Commit**

```bash
git add internal/authz/domain/abac.go internal/authz/domain/abac_test.go
git commit -m "feat(authz): implement PolicyEvaluator with dynamic references and AND semantics"
```

---

### Task 4: Update Interfaces and Signatures

**Files:**
- Modify: `internal/authz/domain/repository.go:102-110`
- Modify: `internal/authz/application/query/check_access.go`

- [ ] **Step 1: Update AuthzReadRepository interface**

In `internal/authz/domain/repository.go`, change line 109:

```go
// Before:
CheckAccess(ctx context.Context, roleID uuid.UUID, path, method string) (bool, error)

// After:
CheckAccess(ctx context.Context, roleID uuid.UUID, path, method string, evalCtx EvaluationContext) (bool, error)
```

Also add a new method to the interface:

```go
FindPoliciesByPermissionIDs(ctx context.Context, permissionIDs []uuid.UUID) ([]*Policy, error)
```

- [ ] **Step 2: Update CheckAccessQuery and Handler**

In `internal/authz/application/query/check_access.go`:

```go
// CheckAccessQuery holds the input for checking whether a role has access to a specific endpoint.
type CheckAccessQuery struct {
	RoleID  uuid.UUID
	Path    string
	Method  string
	EvalCtx domain.EvaluationContext
}

// Handle executes the CheckAccessQuery and returns true if the role has access.
func (h *CheckAccessHandler) Handle(ctx context.Context, q CheckAccessQuery) (bool, error) {
	allowed, err := h.readRepo.CheckAccess(ctx, q.RoleID, q.Path, q.Method, q.EvalCtx)
	if err != nil {
		h.logger.Errorf("check access failed for role %s on %s %s: %v", q.RoleID, q.Method, q.Path, err)
		return false, err
	}
	return allowed, nil
}
```

- [ ] **Step 3: Verify compilation fails at expected places**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go build ./... 2>&1 | head -20`
Expected: Compilation errors in read_repo.go, handler_test.go, middleware files — confirming cascading signature change is needed.

- [ ] **Step 4: Commit**

```bash
git add internal/authz/domain/repository.go internal/authz/application/query/check_access.go
git commit -m "refactor(authz): update CheckAccess signature to accept EvaluationContext"
```

---

### Task 5: Update Read Repository — ABAC in CheckAccess

**Files:**
- Modify: `internal/authz/infrastructure/postgres/read_repo.go:290-349`

- [ ] **Step 1: Add FindPoliciesByPermissionIDs to read repo**

Add to `internal/authz/infrastructure/postgres/read_repo.go`:

```go
// FindPoliciesByPermissionIDs returns all policies for the given permission IDs.
func (r *AuthzReadRepo) FindPoliciesByPermissionIDs(ctx context.Context, permissionIDs []uuid.UUID) ([]*domain.Policy, error) {
	if len(permissionIDs) == 0 {
		return nil, nil
	}

	sql, args, err := r.builder.
		Select("id", "permission_id", "effect", "priority", "active", "conditions", "created_at", "updated_at").
		From(consts.TablePolicy).
		Where(squirrel.Eq{"permission_id": permissionIDs}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, apperrors.HandlePgError(err, consts.TablePolicy, nil)
	}
	defer rows.Close()

	var policies []*domain.Policy
	for rows.Next() {
		var (
			id, permID       uuid.UUID
			effect           string
			priority         int
			active           bool
			condJSON         []byte
			createdAt, updatedAt time.Time
		)
		if err := rows.Scan(&id, &permID, &effect, &priority, &active, &condJSON, &createdAt, &updatedAt); err != nil {
			return nil, apperrors.HandlePgError(err, consts.TablePolicy, nil)
		}
		conds := make(map[string]any)
		if len(condJSON) > 0 {
			_ = json.Unmarshal(condJSON, &conds)
		}
		policies = append(policies, domain.ReconstructPolicy(id, createdAt, updatedAt, nil, permID, domain.PolicyEffect(effect), priority, active, conds))
	}

	return policies, nil
}
```

Add `"time"` to the import block if not already present.

- [ ] **Step 2: Rewrite CheckAccess to collect permission IDs and evaluate policies**

Replace the `CheckAccess` method (lines 290-349) in `internal/authz/infrastructure/postgres/read_repo.go`:

```go
func (r *AuthzReadRepo) CheckAccess(ctx context.Context, roleID uuid.UUID, path, method string, evalCtx domain.EvaluationContext) (bool, error) {
	// Query: fetch role name + permission IDs + scopes.
	sql, args, err := r.builder.
		Select("r.name", "rp.permission_id", "s.path", "s.method").
		From(consts.TableRole + " r").
		LeftJoin(rolePermissionTable + " rp ON r.id = rp.role_id").
		LeftJoin(permissionScopeTable + " ps ON rp.permission_id = ps.permission_id").
		LeftJoin(consts.TableScope + " s ON ps.path = s.path AND ps.method = s.method").
		Where(squirrel.Eq{"r.id": roleID}).
		ToSql()
	if err != nil {
		return false, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return false, apperrors.HandlePgError(err, consts.TableRole, map[string]any{"id": roleID})
	}
	defer rows.Close()

	var roleName string
	foundRole := false
	matchedPermIDs := make(map[uuid.UUID]struct{})

	for rows.Next() {
		var (
			rName       string
			permID      *uuid.UUID
			scopePath   *string
			scopeMethod *string
		)
		if err := rows.Scan(&rName, &permID, &scopePath, &scopeMethod); err != nil {
			return false, apperrors.HandlePgError(err, consts.TableRole, map[string]any{"id": roleID})
		}

		if !foundRole {
			roleName = rName
			foundRole = true

			if strings.ToLower(roleName) == "super_admin" {
				return true, nil
			}
		}

		if scopePath == nil || scopeMethod == nil || permID == nil {
			continue
		}

		if matchScope(*scopePath, *scopeMethod, path, method) {
			matchedPermIDs[*permID] = struct{}{}
		}
	}

	if !foundRole {
		return false, apperrors.HandlePgError(fmt.Errorf("role not found"), consts.TableRole, map[string]any{"id": roleID})
	}

	// RBAC denied — no matching scopes.
	if len(matchedPermIDs) == 0 {
		return false, nil
	}

	// Collect permission IDs for policy lookup.
	permIDs := make([]uuid.UUID, 0, len(matchedPermIDs))
	for id := range matchedPermIDs {
		permIDs = append(permIDs, id)
	}

	// Fetch policies for matched permissions.
	policies, err := r.FindPoliciesByPermissionIDs(ctx, permIDs)
	if err != nil {
		return false, err
	}

	// No policies — RBAC is sufficient.
	if len(policies) == 0 {
		return true, nil
	}

	// Inject role name into eval context for user.role_name conditions.
	if evalCtx.Attrs != nil {
		if evalCtx.Attrs["user"] == nil {
			evalCtx.Attrs["user"] = make(map[string]any)
		}
		evalCtx.Attrs["user"]["role_name"] = roleName
	}

	// Evaluate ABAC policies.
	evaluator := &domain.PolicyEvaluator{}
	effect, matched := evaluator.Evaluate(policies, evalCtx)
	if !matched {
		return true, nil // No applicable policy — RBAC stands.
	}

	return effect == domain.PolicyAllow, nil
}
```

- [ ] **Step 3: Verify it compiles (with remaining test failures expected)**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go build ./internal/authz/...`
Expected: May still fail in test files — that's Task 6.

- [ ] **Step 4: Commit**

```bash
git add internal/authz/infrastructure/postgres/read_repo.go
git commit -m "feat(authz): integrate ABAC evaluation into CheckAccess with policy fetching"
```

---

### Task 6: Update Middleware to Build EvaluationContext

**Files:**
- Modify: `internal/authz/interfaces/http/middleware/authz.go`

- [ ] **Step 1: Update middleware to build EvaluationContext and pass to CheckAccess**

Replace the middleware's `Authz` method in `internal/authz/interfaces/http/middleware/authz.go`:

```go
package middleware

import (
	"net/http"
	"strings"
	"time"

	access "gct/internal/authz/application/query"
	"gct/internal/authz/domain"
	shared "gct/internal/shared/domain"
	"gct/internal/shared/domain/consts"
	"gct/internal/shared/infrastructure/httpx"
	"gct/internal/shared/infrastructure/httpx/response"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/user/application/query"

	"github.com/gin-gonic/gin"
)

type AuthzMiddleware struct {
	checkAccess     *access.CheckAccessHandler
	findUserForAuth *query.FindUserForAuthHandler
	l               logger.Log
}

func NewAuthzMiddleware(
	checkAccess *access.CheckAccessHandler,
	findUserForAuth *query.FindUserForAuthHandler,
	l logger.Log,
) *AuthzMiddleware {
	return &AuthzMiddleware{
		checkAccess:     checkAccess,
		findUserForAuth: findUserForAuth,
		l:               l,
	}
}

func (m *AuthzMiddleware) Authz(ctx *gin.Context) {
	sessionVal, exists := ctx.Get(consts.CtxSession)
	if !exists {
		response.ControllerResponse(ctx, http.StatusUnauthorized, httpx.ErrUnAuth, nil, false)
		ctx.Abort()
		return
	}

	session, ok := sessionVal.(*shared.AuthSession)
	if !ok {
		m.l.Error("AuthzMiddleware - Authz - session type cast error")
		response.ControllerResponse(ctx, http.StatusInternalServerError, httpx.ErrInternalError, nil, false)
		ctx.Abort()
		return
	}

	user, err := m.findUserForAuth.Handle(ctx.Request.Context(), query.FindUserForAuthQuery{
		UserID: session.UserID,
	})
	if err != nil {
		m.l.Errorw("AuthzMiddleware - Authz - FindUserForAuth", "error", err)
		response.ControllerResponse(ctx, http.StatusUnauthorized, httpx.ErrUserNotFound, nil, false)
		ctx.Abort()
		return
	}

	if user.RoleID == nil {
		m.l.Warnw("AuthzMiddleware - Authz - User has no role", "user_id", user.ID)
		response.ControllerResponse(ctx, http.StatusForbidden, httpx.ErrAccessDenied, nil, false)
		ctx.Abort()
		return
	}

	path := ctx.FullPath()
	if path == "" {
		path = ctx.Request.URL.Path
	}
	method := ctx.Request.Method

	evalCtx := buildEvaluationContext(user, ctx)

	allowed, err := m.checkAccess.Handle(ctx.Request.Context(), access.CheckAccessQuery{
		RoleID:  *user.RoleID,
		Path:    path,
		Method:  strings.ToUpper(method),
		EvalCtx: evalCtx,
	})
	if err != nil {
		m.l.Errorw("AuthzMiddleware - Authz - CheckAccess", "error", err)
		response.ControllerResponse(ctx, http.StatusInternalServerError, httpx.ErrInternalError, nil, false)
		ctx.Abort()
		return
	}

	if !allowed {
		response.ControllerResponse(ctx, http.StatusForbidden, httpx.ErrAccessDenied, nil, false)
		ctx.Abort()
		return
	}

	ctx.Next()
}

func buildEvaluationContext(user *shared.AuthUser, ctx *gin.Context) domain.EvaluationContext {
	userAttrs := map[string]any{
		"id": user.ID.String(),
	}
	if user.RoleID != nil {
		userAttrs["role_id"] = user.RoleID.String()
	}
	for k, v := range user.Attributes {
		userAttrs[k] = v
	}

	envAttrs := map[string]any{
		"ip":         ctx.ClientIP(),
		"user_agent": ctx.GetHeader("User-Agent"),
		"time":       time.Now().Format("15:04"),
	}

	return domain.EvaluationContext{
		Attrs: map[string]map[string]any{
			"user":     userAttrs,
			"env":      envAttrs,
			"resource": {},
			"target":   {},
		},
	}
}
```

- [ ] **Step 2: Verify compilation**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go build ./internal/authz/...`
Expected: May still fail in test files only.

- [ ] **Step 3: Commit**

```bash
git add internal/authz/interfaces/http/middleware/authz.go
git commit -m "feat(authz): build EvaluationContext in middleware and pass to CheckAccess"
```

---

### Task 7: Fix All Existing Tests for New Signature

**Files:**
- Modify: `internal/authz/application/query/get_role_test.go`
- Modify: `internal/authz/application/query/check_access_test.go`
- Modify: `internal/authz/interfaces/http/handler_test.go`
- Modify: `internal/authz/interfaces/http/middleware/authz_test.go`

- [ ] **Step 1: Update mock in query tests**

In `internal/authz/application/query/get_role_test.go`, update the mock struct and `CheckAccess` method:

```go
// Change field:
checkAccessFn func(ctx context.Context, roleID uuid.UUID, path, method string, evalCtx domain.EvaluationContext) (bool, error)

// Change method:
func (m *mockAuthzReadRepository) CheckAccess(ctx context.Context, roleID uuid.UUID, path, method string, evalCtx domain.EvaluationContext) (bool, error) {
	if m.checkAccessFn != nil {
		return m.checkAccessFn(ctx, roleID, path, method, evalCtx)
	}
	return false, nil
}
```

Add `FindPoliciesByPermissionIDs` to the mock:

```go
findPoliciesFn func(ctx context.Context, permIDs []uuid.UUID) ([]*domain.Policy, error)

func (m *mockAuthzReadRepository) FindPoliciesByPermissionIDs(ctx context.Context, permIDs []uuid.UUID) ([]*domain.Policy, error) {
	if m.findPoliciesFn != nil {
		return m.findPoliciesFn(ctx, permIDs)
	}
	return nil, nil
}
```

Add import `"gct/internal/authz/domain"` to the import block.

- [ ] **Step 2: Update check_access_test.go mock calls**

In `internal/authz/application/query/check_access_test.go`, update all `checkAccessFn` signatures and `CheckAccessQuery` structs to include `evalCtx`:

```go
// Update each checkAccessFn to accept evalCtx:
checkAccessFn: func(_ context.Context, rid uuid.UUID, path, method string, _ domain.EvaluationContext) (bool, error) {

// Update each Handle call to include EvalCtx:
handler.Handle(context.Background(), CheckAccessQuery{
    RoleID:  roleID,
    Path:    "/api/v1/users",
    Method:  "GET",
    EvalCtx: domain.EvaluationContext{},
})
```

Add import `"gct/internal/authz/domain"`.

- [ ] **Step 3: Update handler_test.go mock**

In `internal/authz/interfaces/http/handler_test.go`, update `mockAuthzReadRepository.CheckAccess` signature and add `FindPoliciesByPermissionIDs`:

```go
func (m *mockAuthzReadRepository) CheckAccess(_ context.Context, _ uuid.UUID, _, _ string, _ domain.EvaluationContext) (bool, error) {
	return false, nil
}

func (m *mockAuthzReadRepository) FindPoliciesByPermissionIDs(_ context.Context, _ []uuid.UUID) ([]*domain.Policy, error) {
	return nil, nil
}
```

- [ ] **Step 4: Update middleware authz_test.go mock**

In `internal/authz/interfaces/http/middleware/authz_test.go`, update `mockAuthzReadRepository.CheckAccess` signature and add `FindPoliciesByPermissionIDs`:

```go
func (m *mockAuthzReadRepository) CheckAccess(ctx context.Context, roleID uuid.UUID, path, method string, _ domain.EvaluationContext) (bool, error) {
	if m.checkAccessFn != nil {
		return m.checkAccessFn(ctx, roleID, path, method)
	}
	return false, nil
}

func (m *mockAuthzReadRepository) FindPoliciesByPermissionIDs(_ context.Context, _ []uuid.UUID) ([]*domain.Policy, error) {
	return nil, nil
}
```

Update the `checkAccessFn` type in the struct to keep the 4-param internal signature (middleware tests don't need to inspect evalCtx).

- [ ] **Step 5: Run all tests**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go test ./internal/authz/... -v -count=1 2>&1 | tail -40`
Expected: ALL PASS

- [ ] **Step 6: Commit**

```bash
git add internal/authz/application/query/get_role_test.go internal/authz/application/query/check_access_test.go internal/authz/interfaces/http/handler_test.go internal/authz/interfaces/http/middleware/authz_test.go
git commit -m "test(authz): update all mocks and tests for new CheckAccess signature with EvalCtx"
```

---

### Task 8: Integration Tests for ABAC Flow

**Files:**
- Create: `test/integration/authz/access/abac_test.go`

- [ ] **Step 1: Write ABAC integration tests**

```go
// test/integration/authz/access/abac_test.go
package access

import (
	"context"
	"testing"

	"gct/internal/authz/application/command"
	"gct/internal/authz/application/query"
	"gct/internal/authz/domain"
	shared "gct/internal/shared/domain"
)

func TestIntegration_ABAC_RBACPassNoPolicies_Allowed(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	seedRoleWithScope(t, bc, "viewer", "pages.read", "/api/v1/pages", "GET")

	roles, _ := bc.ListRoles.Handle(ctx, query.ListRolesQuery{Pagination: shared.Pagination{Limit: 10}})
	roleID := roles.Roles[0].ID

	allowed, err := bc.CheckAccess.Handle(ctx, query.CheckAccessQuery{
		RoleID:  roleID,
		Path:    "/api/v1/pages",
		Method:  "GET",
		EvalCtx: domain.EvaluationContext{Attrs: map[string]map[string]any{"user": {}, "env": {}}},
	})
	if err != nil {
		t.Fatalf("CheckAccess: %v", err)
	}
	if !allowed {
		t.Error("expected allowed when no policies exist (RBAC sufficient)")
	}
}

func TestIntegration_ABAC_DenyPolicy_Denied(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	seedRoleWithScope(t, bc, "editor", "articles.edit", "/api/v1/articles", "PUT")

	roles, _ := bc.ListRoles.Handle(ctx, query.ListRolesQuery{Pagination: shared.Pagination{Limit: 10}})
	perms, _ := bc.ListPermissions.Handle(ctx, query.ListPermissionsQuery{Pagination: shared.Pagination{Limit: 10}})
	roleID := roles.Roles[0].ID
	permID := perms.Permissions[0].ID

	// Create DENY policy: deny if IP is in blocklist.
	err := bc.CreatePolicy.Handle(ctx, command.CreatePolicyCommand{
		PermissionID: permID,
		Effect:       "DENY",
		Priority:     10,
		Conditions:   map[string]any{"env.ip_in": []any{"192.168.1.100"}},
	})
	if err != nil {
		t.Fatalf("CreatePolicy: %v", err)
	}

	// Access from blocked IP should be denied.
	allowed, err := bc.CheckAccess.Handle(ctx, query.CheckAccessQuery{
		RoleID: roleID,
		Path:   "/api/v1/articles",
		Method: "PUT",
		EvalCtx: domain.EvaluationContext{
			Attrs: map[string]map[string]any{
				"user": {},
				"env":  {"ip": "192.168.1.100"},
			},
		},
	})
	if err != nil {
		t.Fatalf("CheckAccess: %v", err)
	}
	if allowed {
		t.Error("expected DENY policy to block access")
	}
}

func TestIntegration_ABAC_DenyPolicy_DifferentIP_Allowed(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	seedRoleWithScope(t, bc, "editor", "articles.edit", "/api/v1/articles", "PUT")

	roles, _ := bc.ListRoles.Handle(ctx, query.ListRolesQuery{Pagination: shared.Pagination{Limit: 10}})
	perms, _ := bc.ListPermissions.Handle(ctx, query.ListPermissionsQuery{Pagination: shared.Pagination{Limit: 10}})
	roleID := roles.Roles[0].ID
	permID := perms.Permissions[0].ID

	err := bc.CreatePolicy.Handle(ctx, command.CreatePolicyCommand{
		PermissionID: permID,
		Effect:       "DENY",
		Priority:     10,
		Conditions:   map[string]any{"env.ip_in": []any{"192.168.1.100"}},
	})
	if err != nil {
		t.Fatalf("CreatePolicy: %v", err)
	}

	// Access from a different IP should be allowed (policy condition doesn't match).
	allowed, err := bc.CheckAccess.Handle(ctx, query.CheckAccessQuery{
		RoleID: roleID,
		Path:   "/api/v1/articles",
		Method: "PUT",
		EvalCtx: domain.EvaluationContext{
			Attrs: map[string]map[string]any{
				"user": {},
				"env":  {"ip": "10.0.0.1"},
			},
		},
	})
	if err != nil {
		t.Fatalf("CheckAccess: %v", err)
	}
	if !allowed {
		t.Error("expected allowed — DENY policy condition does not match this IP")
	}
}

func TestIntegration_ABAC_AllowPolicy_Matches(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	seedRoleWithScope(t, bc, "auditor", "audit.view", "/api/v1/audit", "GET")

	roles, _ := bc.ListRoles.Handle(ctx, query.ListRolesQuery{Pagination: shared.Pagination{Limit: 10}})
	perms, _ := bc.ListPermissions.Handle(ctx, query.ListPermissionsQuery{Pagination: shared.Pagination{Limit: 10}})
	roleID := roles.Roles[0].ID
	permID := perms.Permissions[0].ID

	err := bc.CreatePolicy.Handle(ctx, command.CreatePolicyCommand{
		PermissionID: permID,
		Effect:       "ALLOW",
		Priority:     10,
		Conditions:   map[string]any{"user.role_name": "auditor"},
	})
	if err != nil {
		t.Fatalf("CreatePolicy: %v", err)
	}

	// role_name is injected by CheckAccess from the DB.
	allowed, err := bc.CheckAccess.Handle(ctx, query.CheckAccessQuery{
		RoleID: roleID,
		Path:   "/api/v1/audit",
		Method: "GET",
		EvalCtx: domain.EvaluationContext{
			Attrs: map[string]map[string]any{
				"user": {},
				"env":  {},
			},
		},
	})
	if err != nil {
		t.Fatalf("CheckAccess: %v", err)
	}
	if !allowed {
		t.Error("expected ALLOW policy to grant access")
	}
}

func TestIntegration_ABAC_RBACFail_PoliciesNotConsulted(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	seedRoleWithScope(t, bc, "viewer", "pages.read", "/api/v1/pages", "GET")

	roles, _ := bc.ListRoles.Handle(ctx, query.ListRolesQuery{Pagination: shared.Pagination{Limit: 10}})
	roleID := roles.Roles[0].ID

	// Try accessing a path not in scopes — RBAC should deny before policies are checked.
	allowed, err := bc.CheckAccess.Handle(ctx, query.CheckAccessQuery{
		RoleID: roleID,
		Path:   "/api/v1/admin",
		Method: "DELETE",
		EvalCtx: domain.EvaluationContext{
			Attrs: map[string]map[string]any{
				"user": {},
				"env":  {},
			},
		},
	})
	if err != nil {
		t.Fatalf("CheckAccess: %v", err)
	}
	if allowed {
		t.Error("expected RBAC denial — path not in scopes")
	}
}

func TestIntegration_ABAC_InactivePolicy_Ignored(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	seedRoleWithScope(t, bc, "editor", "docs.edit", "/api/v1/docs", "PUT")

	roles, _ := bc.ListRoles.Handle(ctx, query.ListRolesQuery{Pagination: shared.Pagination{Limit: 10}})
	perms, _ := bc.ListPermissions.Handle(ctx, query.ListPermissionsQuery{Pagination: shared.Pagination{Limit: 10}})
	roleID := roles.Roles[0].ID
	permID := perms.Permissions[0].ID

	// Create DENY policy then toggle it off.
	err := bc.CreatePolicy.Handle(ctx, command.CreatePolicyCommand{
		PermissionID: permID,
		Effect:       "DENY",
		Priority:     100,
		Conditions:   map[string]any{"env.ip_in": []any{"10.0.0.1"}},
	})
	if err != nil {
		t.Fatalf("CreatePolicy: %v", err)
	}

	policies, _ := bc.ListPolicies.Handle(ctx, query.ListPoliciesQuery{Pagination: shared.Pagination{Limit: 10}})
	policyID := policies.Policies[0].ID

	err = bc.TogglePolicy.Handle(ctx, command.TogglePolicyCommand{ID: policyID})
	if err != nil {
		t.Fatalf("TogglePolicy: %v", err)
	}

	// Inactive policy should not block access.
	allowed, err := bc.CheckAccess.Handle(ctx, query.CheckAccessQuery{
		RoleID: roleID,
		Path:   "/api/v1/docs",
		Method: "PUT",
		EvalCtx: domain.EvaluationContext{
			Attrs: map[string]map[string]any{
				"user": {},
				"env":  {"ip": "10.0.0.1"},
			},
		},
	})
	if err != nil {
		t.Fatalf("CheckAccess: %v", err)
	}
	if !allowed {
		t.Error("expected allowed — DENY policy is inactive")
	}
}
```

- [ ] **Step 2: Also update existing integration tests in `integration_test.go` to pass EvalCtx**

In `test/integration/authz/access/integration_test.go`, add `EvalCtx: domain.EvaluationContext{Attrs: map[string]map[string]any{}}` to every `CheckAccessQuery`. Add `"gct/internal/authz/domain"` to imports.

- [ ] **Step 3: Run all integration tests**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go test ./test/integration/authz/... -v -count=1 2>&1 | grep -E "PASS|FAIL|ok"`
Expected: ALL PASS

- [ ] **Step 4: Commit**

```bash
git add test/integration/authz/access/abac_test.go test/integration/authz/access/integration_test.go
git commit -m "test(authz): add ABAC integration tests — deny policy, allow policy, inactive policy, RBAC fallback"
```

---

### Task 9: Final Verification

- [ ] **Step 1: Run entire authz test suite with coverage**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go test ./internal/authz/... -coverprofile=/tmp/abac_cover.out -count=1`
Expected: ALL PASS, domain coverage 100%

- [ ] **Step 2: Run integration tests**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go test ./test/integration/authz/... -v -count=1`
Expected: ALL PASS

- [ ] **Step 3: Run full project build**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go build ./...`
Expected: No errors

- [ ] **Step 4: Final commit**

```bash
git add -A
git commit -m "feat(authz): complete ABAC policy evaluation — operators, evaluator, integration"
```
