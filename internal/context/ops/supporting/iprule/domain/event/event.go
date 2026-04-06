package event

import (
	"time"

	"github.com/google/uuid"
)

// IPRuleCreated is a domain event raised when a new IP rule is created.
// Subscribers can use this to invalidate cached firewall rules or trigger real-time enforcement updates.
type IPRuleCreated struct {
	aggregateID uuid.UUID
	occurredAt  time.Time
	IPAddress   string
	Action      string
}

func NewIPRuleCreated(id uuid.UUID, ipAddress, action string) IPRuleCreated {
	return IPRuleCreated{
		aggregateID: id,
		occurredAt:  time.Now(),
		IPAddress:   ipAddress,
		Action:      action,
	}
}

func (e IPRuleCreated) EventName() string     { return "iprule.created" }
func (e IPRuleCreated) OccurredAt() time.Time  { return e.occurredAt }
func (e IPRuleCreated) AggregateID() uuid.UUID { return e.aggregateID }
