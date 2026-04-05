package domain

import (
	"strconv"
	"strings"

	"github.com/google/uuid"
)

type Condition struct {
	id          uuid.UUID
	ruleGroupID uuid.UUID
	attribute   string
	operator    string
	value       string
}

func NewCondition(attribute, operator, value string) Condition {
	return Condition{
		id:        uuid.New(),
		attribute: attribute,
		operator:  operator,
		value:     value,
	}
}

func ReconstructCondition(id, ruleGroupID uuid.UUID, attribute, operator, value string) Condition {
	return Condition{
		id:          id,
		ruleGroupID: ruleGroupID,
		attribute:   attribute,
		operator:    operator,
		value:       value,
	}
}

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

func (c Condition) ID() uuid.UUID         { return c.id }
func (c Condition) RuleGroupID() uuid.UUID { return c.ruleGroupID }
func (c Condition) Attribute() string      { return c.attribute }
func (c Condition) Operator() string       { return c.operator }
func (c Condition) Value() string          { return c.value }

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
