package entity

import (
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
	matcher, ok := ValidOperators[c.operator]
	if !ok {
		return false
	}
	return matcher(userValue, c.value)
}

func (c Condition) ID() uuid.UUID          { return c.id }
func (c Condition) RuleGroupID() uuid.UUID { return c.ruleGroupID }
func (c Condition) Attribute() string      { return c.attribute }
func (c Condition) Operator() string       { return c.operator }
func (c Condition) Value() string          { return c.value }
