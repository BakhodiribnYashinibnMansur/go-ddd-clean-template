package domain

import (
	"time"

	"github.com/google/uuid"
)

// SessionRevokeRequested is raised when a user requests revocation of a single session.
// The User BC subscribes to this event and performs the actual revocation on its aggregate.
type SessionRevokeRequested struct {
	aggregateID uuid.UUID
	occurredAt  time.Time
	SessionID   uuid.UUID
}

func NewSessionRevokeRequested(userID, sessionID uuid.UUID) SessionRevokeRequested {
	return SessionRevokeRequested{
		aggregateID: userID,
		occurredAt:  time.Now(),
		SessionID:   sessionID,
	}
}

func (e SessionRevokeRequested) EventName() string      { return "session.revoke_requested" }
func (e SessionRevokeRequested) OccurredAt() time.Time  { return e.occurredAt }
func (e SessionRevokeRequested) AggregateID() uuid.UUID { return e.aggregateID }

// SessionRevokeAllRequested is raised when a user requests revocation of all their sessions.
// The User BC subscribes to this event and revokes every session on the user aggregate.
type SessionRevokeAllRequested struct {
	aggregateID uuid.UUID
	occurredAt  time.Time
}

func NewSessionRevokeAllRequested(userID uuid.UUID) SessionRevokeAllRequested {
	return SessionRevokeAllRequested{
		aggregateID: userID,
		occurredAt:  time.Now(),
	}
}

func (e SessionRevokeAllRequested) EventName() string      { return "session.revoke_all_requested" }
func (e SessionRevokeAllRequested) OccurredAt() time.Time  { return e.occurredAt }
func (e SessionRevokeAllRequested) AggregateID() uuid.UUID { return e.aggregateID }
