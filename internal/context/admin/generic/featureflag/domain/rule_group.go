package domain

import (
	"time"

	"github.com/google/uuid"
)

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

func ReconstructRuleGroup(id, flagID uuid.UUID, name, variation string, priority int, createdAt, updatedAt time.Time, conditions []Condition) *RuleGroup {
	return &RuleGroup{
		id: id, flagID: flagID, name: name, variation: variation,
		priority: priority, conditions: conditions, createdAt: createdAt, updatedAt: updatedAt,
	}
}

func (rg *RuleGroup) AddCondition(c Condition) {
	c.ruleGroupID = rg.id
	rg.conditions = append(rg.conditions, c)
}

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

func (rg *RuleGroup) ID() uuid.UUID              { return rg.id }
func (rg *RuleGroup) TypedID() RuleGroupID       { return RuleGroupID(rg.id) }
func (rg *RuleGroup) FlagID() uuid.UUID          { return rg.flagID }
func (rg *RuleGroup) TypedFlagID() FeatureFlagID { return FeatureFlagID(rg.flagID) }
func (rg *RuleGroup) Name() string               { return rg.name }
func (rg *RuleGroup) Variation() string          { return rg.variation }
func (rg *RuleGroup) Priority() int              { return rg.priority }
func (rg *RuleGroup) Conditions() []Condition    { return rg.conditions }
func (rg *RuleGroup) CreatedAt() time.Time       { return rg.createdAt }
func (rg *RuleGroup) UpdatedAt() time.Time       { return rg.updatedAt }
