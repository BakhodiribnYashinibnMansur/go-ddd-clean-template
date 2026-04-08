// Package subscriber hosts the activity log BC's event handlers.
// Each file corresponds to a producing BC (e.g. on_user_events.go reacts to
// the user BC's Published Language V2 events that carry field-level changes).
package subscriber

import (
	"context"
	"fmt"

	contractevents "gct/internal/contract/events"
	"gct/internal/context/ops/supporting/activitylog/application/command"
	"gct/internal/context/ops/supporting/activitylog/domain"
	"gct/internal/kernel/application"
	shareddomain "gct/internal/kernel/domain"
	"gct/internal/kernel/infrastructure/logger"

	"github.com/google/uuid"
)

// userV2Topics lists the V2 user events this subscriber reacts to.
var userV2Topics = []string{
	contractevents.EventUserCreatedV2,
	contractevents.EventUserProfileUpdatedV2,
	contractevents.EventUserDeletedV2,
	contractevents.EventUserRoleChangedV2,
	contractevents.EventUserApprovedV2,
	contractevents.EventUserPasswordChangedV2,
}

// UserEventSubscriber writes activity log entries for user V2 events.
type UserEventSubscriber struct {
	createBatch *command.CreateActivityLogBatchHandler
	log         logger.Log
}

// NewUserEventSubscriber builds the subscriber.
func NewUserEventSubscriber(createBatch *command.CreateActivityLogBatchHandler, log logger.Log) *UserEventSubscriber {
	return &UserEventSubscriber{createBatch: createBatch, log: log}
}

// Register wires each V2 topic to the subscriber's handler.
func (s *UserEventSubscriber) Register(bus application.EventBus) error {
	for _, topic := range userV2Topics {
		t := topic
		if err := bus.Subscribe(t, func(ctx context.Context, event shareddomain.DomainEvent) error {
			return s.handle(ctx, event)
		}); err != nil {
			return fmt.Errorf("user_event_subscriber.register: subscribe %s: %w", t, err)
		}
	}
	return nil
}

func (s *UserEventSubscriber) handle(ctx context.Context, event shareddomain.DomainEvent) error {
	var entries []*domain.ActivityLogEntry

	switch e := event.(type) {
	case contractevents.UserCreatedV2:
		entries = changesToEntries(e.ActorID, "user.created", "user", event.AggregateID(), e.Changes)

	case contractevents.UserProfileUpdatedV2:
		entries = changesToEntries(e.ActorID, "user.updated", "user", event.AggregateID(), e.Changes)

	case contractevents.UserDeletedV2:
		entries = append(entries, domain.NewActivityLogEntry(
			e.ActorID, "user.deleted", "user", event.AggregateID(),
			nil, nil, nil, nil,
		))

	case contractevents.UserRoleChangedV2:
		entries = changesToEntries(e.ActorID, "user.role_changed", "user", event.AggregateID(), e.Changes)

	case contractevents.UserApprovedV2:
		entries = changesToEntries(e.ActorID, "user.approved", "user", event.AggregateID(), e.Changes)

	case contractevents.UserPasswordChangedV2:
		entries = changesToEntries(e.ActorID, "user.password_changed", "user", event.AggregateID(), e.Changes)
	}

	if len(entries) == 0 {
		return nil
	}

	if err := s.createBatch.Handle(ctx, command.CreateActivityLogBatchCommand{Entries: entries}); err != nil {
		s.log.Warnc(ctx, "activity log subscriber: failed to persist activity log",
			"event", event.EventName(), "error", err)
		return nil
	}
	return nil
}

func changesToEntries(
	actorID uuid.UUID,
	action, entityType string,
	entityID uuid.UUID,
	changes []contractevents.FieldChange,
) []*domain.ActivityLogEntry {
	entries := make([]*domain.ActivityLogEntry, 0, len(changes))
	for _, c := range changes {
		fn := c.FieldName
		ov := c.OldValue
		nv := c.NewValue
		entries = append(entries, domain.NewActivityLogEntry(
			actorID, action, entityType, entityID,
			&fn, &ov, &nv, nil,
		))
	}
	return entries
}
