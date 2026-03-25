package domain

import (
	"time"

	"github.com/google/uuid"
)

// AuditLogCreated is raised when a new audit log entry is created.
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
