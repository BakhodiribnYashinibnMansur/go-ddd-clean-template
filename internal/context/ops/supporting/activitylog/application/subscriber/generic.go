package subscriber

import (
	"context"
	"fmt"

	"gct/internal/context/ops/supporting/activitylog/application/command"
	"gct/internal/context/ops/supporting/activitylog/domain"
	"gct/internal/kernel/application"
	shareddomain "gct/internal/kernel/domain"
	"gct/internal/kernel/infrastructure/logger"
)

// eventMapping maps an internal domain event name to the activity log action and entity type.
type eventMapping struct {
	Event      string
	Action     string
	EntityType string
}

// eventMappings is the config-driven mapping table. To log a new event, add a row here.
// metric.recorded is intentionally excluded — too high-frequency for activity logging.
var eventMappings = []eventMapping{
	// === IAM: User ===
	{Event: "user.signed_in", Action: "user.signed_in", EntityType: "user"},
	{Event: "user.created", Action: "user.signed_up", EntityType: "user"},
	{Event: "user.deactivated", Action: "user.deactivated", EntityType: "user"},
	{Event: "user.password_changed", Action: "user.password_changed", EntityType: "user"},
	{Event: "user.approved", Action: "user.approved", EntityType: "user"},
	{Event: "user.profile_updated", Action: "user.profile_updated", EntityType: "user"},
	{Event: "user.role_changed", Action: "user.role_changed", EntityType: "user"},

	// === IAM: Session ===
	{Event: "session.revoke_requested", Action: "session.revoked", EntityType: "session"},
	{Event: "session.revoke_all_requested", Action: "session.revoked_all", EntityType: "session"},

	// === IAM: Authz ===
	{Event: "authz.role_created", Action: "role.created", EntityType: "role"},
	{Event: "authz.role_deleted", Action: "role.deleted", EntityType: "role"},
	{Event: "authz.policy_updated", Action: "policy.updated", EntityType: "policy"},
	{Event: "authz.permission_granted", Action: "permission.granted", EntityType: "permission"},

	// === IAM: User Setting ===
	{Event: "usersetting.changed", Action: "usersetting.changed", EntityType: "usersetting"},

	// === Admin: Feature Flag ===
	{Event: "featureflag.toggled", Action: "featureflag.toggled", EntityType: "feature_flag"},
	{Event: "featureflag.created", Action: "featureflag.created", EntityType: "feature_flag"},
	{Event: "featureflag.updated", Action: "featureflag.updated", EntityType: "feature_flag"},
	{Event: "featureflag.deleted", Action: "featureflag.deleted", EntityType: "feature_flag"},

	// === Admin: Site Setting ===
	{Event: "sitesetting.updated", Action: "sitesetting.updated", EntityType: "site_setting"},

	// === Admin: Integration ===
	{Event: "integration.connected", Action: "integration.connected", EntityType: "integration"},

	// === Admin: Error Code ===
	{Event: "errorcode.created", Action: "errorcode.created", EntityType: "error_code"},
	{Event: "errorcode.updated", Action: "errorcode.updated", EntityType: "error_code"},
	{Event: "errorcode.deleted", Action: "errorcode.deleted", EntityType: "error_code"},

	// === Admin: Data Export ===
	{Event: "dataexport.requested", Action: "dataexport.requested", EntityType: "data_export"},
	{Event: "dataexport.completed", Action: "dataexport.completed", EntityType: "data_export"},

	// === Content: File ===
	{Event: "file.uploaded", Action: "file.uploaded", EntityType: "file"},

	// === Content: Translation ===
	{Event: "translation.updated", Action: "translation.updated", EntityType: "translation"},

	// === Content: Notification ===
	{Event: "notification.sent", Action: "notification.sent", EntityType: "notification"},

	// === Content: Announcement ===
	{Event: "announcement.published", Action: "announcement.published", EntityType: "announcement"},

	// === Ops: Rate Limit ===
	{Event: "ratelimit.changed", Action: "ratelimit.changed", EntityType: "rate_limit"},

	// === Ops: System Error ===
	{Event: "system_error.recorded", Action: "system_error.recorded", EntityType: "system_error"},
	{Event: "system_error.resolved", Action: "system_error.resolved", EntityType: "system_error"},

	// === Ops: IP Rule ===
	{Event: "iprule.created", Action: "iprule.created", EntityType: "ip_rule"},
}

// GenericSubscriber subscribes to all V1 domain events across all BCs and writes
// a single activity log entry per event. For events implementing MetadataProvider,
// the metadata string is persisted alongside the action.
type GenericSubscriber struct {
	createBatch *command.CreateActivityLogBatchHandler
	log         logger.Log
}

// NewGenericSubscriber builds the generic subscriber.
func NewGenericSubscriber(createBatch *command.CreateActivityLogBatchHandler, log logger.Log) *GenericSubscriber {
	return &GenericSubscriber{createBatch: createBatch, log: log}
}

// Register wires every mapped event topic to the subscriber's handler.
func (s *GenericSubscriber) Register(bus application.EventBus) error {
	for _, m := range eventMappings {
		mapping := m
		if err := bus.Subscribe(mapping.Event, func(ctx context.Context, event shareddomain.DomainEvent) error {
			return s.handle(ctx, mapping, event)
		}); err != nil {
			return fmt.Errorf("generic_subscriber.register: subscribe %s: %w", mapping.Event, err)
		}
	}
	return nil
}

func (s *GenericSubscriber) handle(ctx context.Context, mapping eventMapping, event shareddomain.DomainEvent) error {
	// Extract optional metadata from events that implement MetadataProvider.
	var metadata *string
	if mp, ok := event.(shareddomain.MetadataProvider); ok {
		m := mp.ActivityMetadata()
		if m != "" {
			metadata = &m
		}
	}

	entry := domain.NewActivityLogEntry(
		event.AggregateID(), // actor_id = aggregate_id for V1 events
		mapping.Action,
		mapping.EntityType,
		event.AggregateID(),
		nil, nil, nil, // no field-level changes
		metadata,
	)

	if err := s.createBatch.Handle(ctx, command.CreateActivityLogBatchCommand{
		Entries: []*domain.ActivityLogEntry{entry},
	}); err != nil {
		s.log.Warnc(ctx, "generic activity log subscriber: failed to persist",
			"event", event.EventName(), "error", err)
		return nil
	}
	return nil
}
