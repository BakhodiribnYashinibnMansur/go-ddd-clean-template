package domain

import (
	"time"

	"github.com/google/uuid"
)

// AuditLogCreated is a domain event emitted every time a new audit record is persisted.
// Downstream handlers can use this for real-time alerting (e.g., on ACCESS_DENIED actions)
// or to fan out to external SIEM systems.
type AuditLogCreated struct {
	aggregateID uuid.UUID
	occurredAt  time.Time
	Action      AuditAction
}

// NewAuditLogCreated creates a new AuditLogCreated event.
func NewAuditLogCreated(auditLogID uuid.UUID, action AuditAction) AuditLogCreated {
	return AuditLogCreated{
		aggregateID: auditLogID,
		occurredAt:  time.Now(),
		Action:      action,
	}
}

func (e AuditLogCreated) EventName() string      { return "audit_log.created" }
func (e AuditLogCreated) OccurredAt() time.Time  { return e.occurredAt }
func (e AuditLogCreated) AggregateID() uuid.UUID { return e.aggregateID }
