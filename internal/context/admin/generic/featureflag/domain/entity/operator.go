package entity

import (
	"strconv"
	"strings"
)

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

// Matcher evaluates whether a user-supplied attribute value matches a
// condition value for a given operator.
type Matcher func(userValue, condValue string) bool

// ValidOperators is the operator registry. Adding a new operator means
// adding an entry here (or calling RegisterOperator) — no existing code
// needs to be modified. This keeps Condition.Match closed for modification
// and open for extension (OCP).
var ValidOperators = map[string]Matcher{
	OpEq:    func(u, c string) bool { return u == c },
	OpNotEq: func(u, c string) bool { return u != c },
	OpIn:    func(u, c string) bool { return containsInList(c, u) },
	OpNotIn: func(u, c string) bool { return !containsInList(c, u) },
	OpGt: func(u, c string) bool {
		return compareFloat(u, c, func(a, b float64) bool { return a > b })
	},
	OpGte: func(u, c string) bool {
		return compareFloat(u, c, func(a, b float64) bool { return a >= b })
	},
	OpLt: func(u, c string) bool {
		return compareFloat(u, c, func(a, b float64) bool { return a < b })
	},
	OpLte: func(u, c string) bool {
		return compareFloat(u, c, func(a, b float64) bool { return a <= b })
	},
	OpContains: func(u, c string) bool { return strings.Contains(u, c) },
}

func IsValidOperator(op string) bool {
	_, ok := ValidOperators[op]
	return ok
}

// RegisterOperator registers a custom operator matcher. Intended for
// plugins/extensions that want to add new operators without touching
// existing code.
func RegisterOperator(name string, matcher Matcher) {
	ValidOperators[name] = matcher
}

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
