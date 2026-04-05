// Package subscriber hosts the audit BC's event handlers. Each file
// corresponds to a producing BC (e.g. on_user_events.go reacts to the user
// BC's Published Language). The subscriber is wired into the event bus in
// the BC's bc.go via RegisterSubscribers.
package subscriber

import (
	"context"
	"fmt"

	auditcmd "gct/internal/context/iam/supporting/audit/application/command"
	auditdomain "gct/internal/context/iam/supporting/audit/domain"
	"gct/internal/kernel/application"
	shareddomain "gct/internal/kernel/domain"
	"gct/internal/kernel/infrastructure/logger"
)

// userEventTopics lists the user BC event names this subscriber reacts to.
// The names follow the Published Language convention. We also subscribe to
// the legacy un-versioned names so existing publishers (which still emit
// internal domain events) are captured during the migration.
var userEventTopics = []struct {
	Name   string
	Action auditdomain.AuditAction
}{
	{Name: "user.created", Action: auditdomain.AuditActionUserCreate},
	{Name: "user.created.v1", Action: auditdomain.AuditActionUserCreate},
	{Name: "user.signed_in", Action: auditdomain.AuditActionLogin},
	{Name: "user.signed_in.v1", Action: auditdomain.AuditActionLogin},
	{Name: "user.deactivated", Action: auditdomain.AuditActionUserDelete},
	{Name: "user.deactivated.v1", Action: auditdomain.AuditActionUserDelete},
	{Name: "user.password_changed", Action: auditdomain.AuditActionPasswordChange},
	{Name: "user.password_changed.v1", Action: auditdomain.AuditActionPasswordChange},
	{Name: "user.role_changed", Action: auditdomain.AuditActionRoleAssign},
	{Name: "user.role_changed.v1", Action: auditdomain.AuditActionRoleAssign},
}

// UserEventSubscriber writes an audit log entry for every user-facing event
// it receives on the bus. It never imports the user BC — all data it needs
// arrives through the Published Language contracts / DomainEvent metadata.
type UserEventSubscriber struct {
	create *auditcmd.CreateAuditLogHandler
	log    logger.Log
}

// NewUserEventSubscriber builds the subscriber.
func NewUserEventSubscriber(create *auditcmd.CreateAuditLogHandler, log logger.Log) *UserEventSubscriber {
	return &UserEventSubscriber{create: create, log: log}
}

// Register wires each topic to the subscriber's handler. Call once from the
// composition root after the BC is constructed.
func (s *UserEventSubscriber) Register(bus application.EventBus) error {
	for _, t := range userEventTopics {
		topic := t // capture
		if err := bus.Subscribe(topic.Name, func(ctx context.Context, event shareddomain.DomainEvent) error {
			return s.handle(ctx, topic.Action, event)
		}); err != nil {
			return fmt.Errorf("user_event_subscriber.register: subscribe %s: %w", topic.Name, err)
		}
	}
	return nil
}

func (s *UserEventSubscriber) handle(ctx context.Context, action auditdomain.AuditAction, event shareddomain.DomainEvent) error {
	userID := event.AggregateID()
	cmd := auditcmd.CreateAuditLogCommand{
		UserID:  &userID,
		Action:  action,
		Success: true,
		Metadata: map[string]string{
			"source_event": event.EventName(),
		},
	}
	if err := s.create.Handle(ctx, cmd); err != nil {
		s.log.Warnc(ctx, "audit subscriber: failed to persist audit log",
			"event", event.EventName(), "error", err)
		// Do not propagate — a failed audit side-effect must not abort the
		// originating command. The subscriber is best-effort.
		return nil
	}
	return nil
}
