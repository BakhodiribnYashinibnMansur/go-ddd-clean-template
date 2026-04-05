package domain

import (
	"testing"

	"github.com/google/uuid"
)

// --- EvaluationContext ---

func TestEvaluationContext_Resolve(t *testing.T) {
	t.Parallel()

	ctx := EvaluationContext{
		Attrs: map[string]map[string]any{
			"user": {"role_name": "admin", "department": "engineering"},
			"env":  {"ip": "10.0.0.1"},
		},
	}

	val, ok := ctx.Resolve("user.role_name")
	if !ok || val != "admin" {
		t.Errorf("expected admin, got %v (ok=%v)", val, ok)
	}

	val, ok = ctx.Resolve("env.ip")
	if !ok || val != "10.0.0.1" {
		t.Errorf("expected 10.0.0.1, got %v (ok=%v)", val, ok)
	}

	// Missing key in namespace
	_, ok = ctx.Resolve("user.nonexistent")
	if ok {
		t.Error("expected false for missing key")
	}

	// Missing namespace
	_, ok = ctx.Resolve("unknown.field")
	if ok {
		t.Error("expected false for missing namespace")
	}

	// No dot
	_, ok = ctx.Resolve("nodot")
	if ok {
		t.Error("expected false for key without dot")
	}
}

func TestEvaluationContext_Resolve_TargetRef(t *testing.T) {
	t.Parallel()

	ctx := EvaluationContext{
		Attrs: map[string]map[string]any{
			"target": {"user.relation_names": []any{"friend", "coworker"}},
		},
	}

	val, ok := ctx.Resolve("target.user.relation_names")
	if !ok {
		t.Fatal("expected ok for target compound key")
	}
	list, listOk := val.([]any)
	if !listOk || len(list) != 2 {
		t.Errorf("expected []any with 2 elements, got %v", val)
	}
}

// --- parseConditionKey ---

func TestParseConditionKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		key       string
		wantNS    string
		wantField string
		wantOp    string
	}{
		{"user.role_name", "user", "role_name", "equals"},
		{"env.ip_not_in", "env", "ip", "not_in"},
		{"env.ip_not_equals", "env", "ip", "not_equals"},
		{"user.relation_names_any", "user", "relation_names", "any"},
		{"user.relation_names_all", "user", "relation_names", "all"},
		{"user.age_gt", "user", "age", "gt"},
		{"user.age_gte", "user", "age", "gte"},
		{"user.age_lt", "user", "age", "lt"},
		{"user.age_lte", "user", "age", "lte"},
		{"user.score_between", "user", "score", "between"},
		{"env.ip_in", "env", "ip", "in"},
		{"user.role_name_contains", "user", "role_name", "contains"},
	}

	for _, tc := range tests {
		ns, field, op := parseConditionKey(tc.key)
		if ns != tc.wantNS || field != tc.wantField || op != tc.wantOp {
			t.Errorf("parseConditionKey(%q) = (%q, %q, %q); want (%q, %q, %q)",
				tc.key, ns, field, op, tc.wantNS, tc.wantField, tc.wantOp)
		}
	}
}

// --- Operator Tests ---

func TestOpEquals(t *testing.T) {
	t.Parallel()

	op := operators["equals"]
	if !op("admin", "admin") {
		t.Error("expected equal strings to match")
	}
	if op("admin", "user") {
		t.Error("expected different strings to not match")
	}
	if !op(42, 42) {
		t.Error("expected equal ints to match")
	}
	if !op(3.14, 3.14) {
		t.Error("expected equal floats to match")
	}
	if op(nil, "x") {
		t.Error("expected nil != x")
	}
	// float64 vs int
	if !op(float64(42), 42) {
		t.Error("expected float64(42) == 42")
	}
}

func TestOpNotEquals(t *testing.T) {
	t.Parallel()

	op := operators["not_equals"]
	if !op("admin", "user") {
		t.Error("expected different strings to match not_equals")
	}
	if op("admin", "admin") {
		t.Error("expected same strings to not match not_equals")
	}
}

func TestOpIn(t *testing.T) {
	t.Parallel()

	op := operators["in"]
	if !op("admin", []any{"admin", "user"}) {
		t.Error("expected admin in [admin, user]")
	}
	if op("guest", []any{"admin", "user"}) {
		t.Error("expected guest not in [admin, user]")
	}
	// non-list condition
	if op("admin", "admin") {
		t.Error("expected false for non-list conditionVal")
	}
}

func TestOpNotIn(t *testing.T) {
	t.Parallel()

	op := operators["not_in"]
	if !op("guest", []any{"admin", "user"}) {
		t.Error("expected guest not_in [admin, user]")
	}
	if op("admin", []any{"admin", "user"}) {
		t.Error("expected admin to fail not_in [admin, user]")
	}
}

func TestOpContains(t *testing.T) {
	t.Parallel()

	op := operators["contains"]
	if !op([]any{"a", "b", "c"}, "b") {
		t.Error("expected [a,b,c] contains b")
	}
	if op([]any{"a", "b", "c"}, "z") {
		t.Error("expected [a,b,c] not contains z")
	}
	// non-list field
	if op("hello", "h") {
		t.Error("expected false for non-list fieldVal")
	}
}

func TestOpAny(t *testing.T) {
	t.Parallel()

	op := operators["any"]
	if !op([]any{"a", "b"}, []any{"b", "c"}) {
		t.Error("expected intersection")
	}
	if op([]any{"a", "b"}, []any{"c", "d"}) {
		t.Error("expected no intersection")
	}
}

func TestOpAll(t *testing.T) {
	t.Parallel()

	op := operators["all"]
	if !op([]any{"a", "b", "c"}, []any{"a", "c"}) {
		t.Error("expected all present")
	}
	if op([]any{"a", "b"}, []any{"a", "c"}) {
		t.Error("expected not all present")
	}
}

func TestOpGt(t *testing.T) {
	t.Parallel()

	op := operators["gt"]
	if !op(10, 5) {
		t.Error("expected 10 > 5")
	}
	if op(5, 10) {
		t.Error("expected !(5 > 10)")
	}
	if op(5, 5) {
		t.Error("expected !(5 > 5)")
	}
}

func TestOpGte(t *testing.T) {
	t.Parallel()

	op := operators["gte"]
	if !op(10, 5) {
		t.Error("expected 10 >= 5")
	}
	if !op(5, 5) {
		t.Error("expected 5 >= 5")
	}
	if op(3, 5) {
		t.Error("expected !(3 >= 5)")
	}
}

func TestOpLt(t *testing.T) {
	t.Parallel()

	op := operators["lt"]
	if !op(3, 5) {
		t.Error("expected 3 < 5")
	}
	if op(5, 3) {
		t.Error("expected !(5 < 3)")
	}
}

func TestOpLte(t *testing.T) {
	t.Parallel()

	op := operators["lte"]
	if !op(3, 5) {
		t.Error("expected 3 <= 5")
	}
	if !op(5, 5) {
		t.Error("expected 5 <= 5")
	}
	if op(7, 5) {
		t.Error("expected !(7 <= 5)")
	}
}

func TestOpBetween(t *testing.T) {
	t.Parallel()

	op := operators["between"]
	if !op(5, []any{1, 10}) {
		t.Error("expected 5 between 1..10")
	}
	if op(0, []any{1, 10}) {
		t.Error("expected 0 not between 1..10")
	}
	if op(11, []any{1, 10}) {
		t.Error("expected 11 not between 1..10")
	}
	// at bounds
	if !op(1, []any{1, 10}) {
		t.Error("expected 1 between 1..10 (inclusive)")
	}
	if !op(10, []any{1, 10}) {
		t.Error("expected 10 between 1..10 (inclusive)")
	}
	// numeric
	if !op(float64(5), []any{float64(1), float64(10)}) {
		t.Error("expected float 5 between 1..10")
	}
}

// --- PolicyEvaluator ---

func newTestPolicy(effect PolicyEffect, priority int, active bool, conditions map[string]any) *Policy {
	p := NewPolicy(uuid.New(), effect)
	p.SetPriority(priority)
	p.SetConditions(conditions)
	if !active {
		p.Toggle()
	}
	return p
}

func TestPolicyEvaluator_SingleAllow(t *testing.T) {
	t.Parallel()

	e := &PolicyEvaluator{}
	p := newTestPolicy(PolicyAllow, 1, true, map[string]any{
		"user.role_name": "admin",
	})
	ctx := EvaluationContext{
		Attrs: map[string]map[string]any{
			"user": {"role_name": "admin"},
		},
	}
	eff, ok := e.Evaluate([]*Policy{p}, ctx)
	if !ok || eff != PolicyAllow {
		t.Errorf("expected (ALLOW, true), got (%s, %v)", eff, ok)
	}
}

func TestPolicyEvaluator_SingleDeny(t *testing.T) {
	t.Parallel()

	e := &PolicyEvaluator{}
	p := newTestPolicy(PolicyDeny, 1, true, map[string]any{
		"user.role_name": "guest",
	})
	ctx := EvaluationContext{
		Attrs: map[string]map[string]any{
			"user": {"role_name": "guest"},
		},
	}
	eff, ok := e.Evaluate([]*Policy{p}, ctx)
	if !ok || eff != PolicyDeny {
		t.Errorf("expected (DENY, true), got (%s, %v)", eff, ok)
	}
}

func TestPolicyEvaluator_DenyWinsOverAllow(t *testing.T) {
	t.Parallel()

	e := &PolicyEvaluator{}
	allow := newTestPolicy(PolicyAllow, 10, true, map[string]any{
		"user.role_name": "admin",
	})
	deny := newTestPolicy(PolicyDeny, 1, true, map[string]any{
		"user.role_name": "admin",
	})
	ctx := EvaluationContext{
		Attrs: map[string]map[string]any{
			"user": {"role_name": "admin"},
		},
	}
	eff, ok := e.Evaluate([]*Policy{allow, deny}, ctx)
	if !ok || eff != PolicyDeny {
		t.Errorf("expected DENY to win, got (%s, %v)", eff, ok)
	}
}

func TestPolicyEvaluator_InactivePolicySkipped(t *testing.T) {
	t.Parallel()

	e := &PolicyEvaluator{}
	p := newTestPolicy(PolicyAllow, 1, false, map[string]any{
		"user.role_name": "admin",
	})
	ctx := EvaluationContext{
		Attrs: map[string]map[string]any{
			"user": {"role_name": "admin"},
		},
	}
	_, ok := e.Evaluate([]*Policy{p}, ctx)
	if ok {
		t.Error("expected no match for inactive policy")
	}
}

func TestPolicyEvaluator_NoMatch(t *testing.T) {
	t.Parallel()

	e := &PolicyEvaluator{}
	p := newTestPolicy(PolicyAllow, 1, true, map[string]any{
		"user.role_name": "admin",
	})
	ctx := EvaluationContext{
		Attrs: map[string]map[string]any{
			"user": {"role_name": "guest"},
		},
	}
	_, ok := e.Evaluate([]*Policy{p}, ctx)
	if ok {
		t.Error("expected no match when conditions don't match")
	}
}

func TestPolicyEvaluator_ANDSemantics(t *testing.T) {
	t.Parallel()

	e := &PolicyEvaluator{}
	p := newTestPolicy(PolicyAllow, 1, true, map[string]any{
		"user.role_name":  "admin",
		"user.department": "engineering",
	})
	ctx := EvaluationContext{
		Attrs: map[string]map[string]any{
			"user": {"role_name": "admin", "department": "sales"},
		},
	}
	_, ok := e.Evaluate([]*Policy{p}, ctx)
	if ok {
		t.Error("expected no match when only one condition matches (AND semantics)")
	}
}

func TestPolicyEvaluator_EmptyConditions_AlwaysMatches(t *testing.T) {
	t.Parallel()

	e := &PolicyEvaluator{}
	p := newTestPolicy(PolicyAllow, 1, true, map[string]any{})
	ctx := EvaluationContext{
		Attrs: map[string]map[string]any{},
	}
	eff, ok := e.Evaluate([]*Policy{p}, ctx)
	if !ok || eff != PolicyAllow {
		t.Errorf("expected empty conditions to always match, got (%s, %v)", eff, ok)
	}
}

func TestPolicyEvaluator_DynamicReference(t *testing.T) {
	t.Parallel()

	e := &PolicyEvaluator{}
	p := newTestPolicy(PolicyAllow, 1, true, map[string]any{
		"user.relation_names_any": "$target.user.relation_names",
	})
	ctx := EvaluationContext{
		Attrs: map[string]map[string]any{
			"user":   {"relation_names": []any{"friend", "coworker"}},
			"target": {"user.relation_names": []any{"coworker", "manager"}},
		},
	}
	eff, ok := e.Evaluate([]*Policy{p}, ctx)
	if !ok || eff != PolicyAllow {
		t.Errorf("expected dynamic ref match, got (%s, %v)", eff, ok)
	}
}

func TestPolicyEvaluator_DynamicReference_NoMatch(t *testing.T) {
	t.Parallel()

	e := &PolicyEvaluator{}
	p := newTestPolicy(PolicyAllow, 1, true, map[string]any{
		"user.relation_names_any": "$target.user.relation_names",
	})
	ctx := EvaluationContext{
		Attrs: map[string]map[string]any{
			"user":   {"relation_names": []any{"friend"}},
			"target": {"user.relation_names": []any{"manager"}},
		},
	}
	_, ok := e.Evaluate([]*Policy{p}, ctx)
	if ok {
		t.Error("expected no match when dynamic ref resolves but no intersection")
	}
}

func TestPolicyEvaluator_NoPolicies(t *testing.T) {
	t.Parallel()

	e := &PolicyEvaluator{}
	ctx := EvaluationContext{Attrs: map[string]map[string]any{}}

	_, ok := e.Evaluate(nil, ctx)
	if ok {
		t.Error("expected no match for nil policies")
	}

	_, ok = e.Evaluate([]*Policy{}, ctx)
	if ok {
		t.Error("expected no match for empty policies")
	}
}
