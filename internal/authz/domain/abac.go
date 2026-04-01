package domain

import (
	"fmt"
	"sort"
	"strings"
)

// EvaluationContext holds the attribute bags used during ABAC policy evaluation.
// Each top-level key is a namespace (e.g., "user", "env", "target") and the nested
// map contains the attributes for that namespace.
type EvaluationContext struct {
	Attrs map[string]map[string]any
}

// Resolve looks up a dotted key in the context. The first dot separates the namespace
// from the field key; any further dots are part of the field key itself.
// Returns (nil, false) if the key has no dot, the namespace is missing, or the field is absent.
func (ec EvaluationContext) Resolve(key string) (any, bool) {
	idx := strings.IndexByte(key, '.')
	if idx < 0 {
		return nil, false
	}
	ns := key[:idx]
	field := key[idx+1:]
	bucket, ok := ec.Attrs[ns]
	if !ok {
		return nil, false
	}
	val, ok := bucket[field]
	return val, ok
}

// ---------------------------------------------------------------------------
// Operator registry
// ---------------------------------------------------------------------------

// OperatorFunc compares a resolved field value against a condition value.
type OperatorFunc func(fieldVal, conditionVal any) bool

// operators is the global registry populated in init().
var operators = map[string]OperatorFunc{}

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

// ---------------------------------------------------------------------------
// Operator implementations
// ---------------------------------------------------------------------------

func opEquals(a, b any) bool {
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

func opNotEquals(a, b any) bool {
	return !opEquals(a, b)
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

func opGt(a, b any) bool  { return compareValues(a, b) > 0 }
func opGte(a, b any) bool { return compareValues(a, b) >= 0 }
func opLt(a, b any) bool  { return compareValues(a, b) < 0 }
func opLte(a, b any) bool { return compareValues(a, b) <= 0 }

func opBetween(fieldVal, conditionVal any) bool {
	bounds, ok := toSlice(conditionVal)
	if !ok || len(bounds) != 2 {
		return false
	}
	return compareValues(fieldVal, bounds[0]) >= 0 && compareValues(fieldVal, bounds[1]) <= 0
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// toSlice converts a value to []any. Handles []any and []string.
func toSlice(v any) ([]any, bool) {
	switch s := v.(type) {
	case []any:
		return s, true
	case []string:
		out := make([]any, len(s))
		for i, item := range s {
			out[i] = item
		}
		return out, true
	}
	return nil, false
}

// compareValues compares two values, returning -1, 0, or 1. It attempts numeric
// comparison first; if either value is not numeric, it falls back to string comparison.
func compareValues(a, b any) int {
	fa, okA := toFloat64(a)
	fb, okB := toFloat64(b)
	if okA && okB {
		if fa < fb {
			return -1
		}
		if fa > fb {
			return 1
		}
		return 0
	}
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

// toFloat64 attempts to convert a value to float64 for numeric comparison.
func toFloat64(v any) (float64, bool) {
	switch n := v.(type) {
	case int:
		return float64(n), true
	case int8:
		return float64(n), true
	case int16:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case float32:
		return float64(n), true
	case float64:
		return n, true
	}
	return 0, false
}

// ---------------------------------------------------------------------------
// Condition key parser
// ---------------------------------------------------------------------------

// compoundOperators are multi-word operators that must be checked before single-word ones.
var compoundOperators = []string{"not_equals", "not_in"}

// parseConditionKey splits a dotted condition key into namespace, field, and operator.
//
//	"user.role_name"           → ("user", "role_name", "equals")
//	"env.ip_not_in"            → ("env", "ip", "not_in")
//	"user.relation_names_any"  → ("user", "relation_names", "any")
func parseConditionKey(key string) (namespace, field, operator string) {
	dotIdx := strings.IndexByte(key, '.')
	if dotIdx < 0 {
		return "", key, "equals"
	}
	namespace = key[:dotIdx]
	rest := key[dotIdx+1:]

	// Try compound operators first.
	for _, cop := range compoundOperators {
		suffix := "_" + cop
		if strings.HasSuffix(rest, suffix) {
			field = rest[:len(rest)-len(suffix)]
			operator = cop
			return
		}
	}

	// Try single-word operators from the registry.
	lastUnderscore := strings.LastIndexByte(rest, '_')
	if lastUnderscore > 0 {
		candidate := rest[lastUnderscore+1:]
		if _, ok := operators[candidate]; ok {
			field = rest[:lastUnderscore]
			operator = candidate
			return
		}
	}

	// Default to equals.
	field = rest
	operator = "equals"
	return
}

// ---------------------------------------------------------------------------
// Dynamic reference resolution
// ---------------------------------------------------------------------------

// resolveDynamicRef resolves condition values that start with "$target." by looking
// them up in the evaluation context. Non-string values or strings without the prefix
// are returned unchanged.
func resolveDynamicRef(val any, ctx EvaluationContext) any {
	s, ok := val.(string)
	if !ok {
		return val
	}
	if !strings.HasPrefix(s, "$") {
		return val
	}
	ref := s[1:] // strip the leading $
	if resolved, found := ctx.Resolve(ref); found {
		return resolved
	}
	return val
}

// ---------------------------------------------------------------------------
// Condition evaluation
// ---------------------------------------------------------------------------

// evaluateConditions checks whether all conditions in the map hold true against
// the given context. Empty conditions always match (return true). Uses AND semantics.
func evaluateConditions(conditions map[string]string, ctx EvaluationContext) bool {
	if len(conditions) == 0 {
		return true
	}
	for key, condVal := range conditions {
		ns, fieldName, op := parseConditionKey(key)
		resolvedKey := ns + "." + fieldName
		fieldVal, ok := ctx.Resolve(resolvedKey)
		if !ok {
			return false
		}

		resolved := resolveDynamicRef(condVal, ctx)

		opFunc, opOk := operators[op]
		if !opOk {
			return false
		}
		if !opFunc(fieldVal, resolved) {
			return false
		}
	}
	return true
}

// ---------------------------------------------------------------------------
// Policy evaluator
// ---------------------------------------------------------------------------

// PolicyEvaluator evaluates a list of ABAC policies against an evaluation context.
// It implements explicit-deny-wins semantics: any matching DENY policy overrides
// all ALLOW policies regardless of priority.
type PolicyEvaluator struct{}

// Evaluate processes policies and returns the resulting effect. It returns ("", false)
// when no policy matches. Active policies are sorted by priority (descending) and
// evaluated with AND semantics on conditions.
func (e *PolicyEvaluator) Evaluate(policies []*Policy, ctx EvaluationContext) (PolicyEffect, bool) {
	if len(policies) == 0 {
		return "", false
	}

	// Filter active policies.
	active := make([]*Policy, 0, len(policies))
	for _, p := range policies {
		if p.IsActive() {
			active = append(active, p)
		}
	}
	if len(active) == 0 {
		return "", false
	}

	// Sort by priority descending.
	sort.Slice(active, func(i, j int) bool {
		return active[i].Priority() > active[j].Priority()
	})

	// Evaluate and collect matches.
	var matched []*Policy
	for _, p := range active {
		if evaluateConditions(p.Conditions(), ctx) {
			matched = append(matched, p)
		}
	}

	if len(matched) == 0 {
		return "", false
	}

	// Explicit deny wins.
	for _, p := range matched {
		if p.Effect() == PolicyDeny {
			return PolicyDeny, true
		}
	}

	return PolicyAllow, true
}
