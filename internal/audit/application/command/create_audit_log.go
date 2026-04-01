package command

import (
	"context"

	"gct/internal/audit/domain"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"

	"github.com/google/uuid"
)

// CreateAuditLogCommand captures a security-relevant event for immutable storage.
// Most fields are optional pointers because not all actions have an associated user, resource, or policy.
// Metadata carries arbitrary key-value context (e.g., changed fields, request IDs) for post-hoc forensic analysis.
type CreateAuditLogCommand struct {
	UserID       *uuid.UUID
	SessionID    *uuid.UUID
	Action       domain.AuditAction
	ResourceType *string
	ResourceID   *uuid.UUID
	Platform     *string
	IPAddress    *string
	UserAgent    *string
	Permission   *string
	PolicyID     *uuid.UUID
	Decision     *string
	Success      bool
	ErrorMessage *string
	Metadata     map[string]string
}

// CreateAuditLogHandler persists audit log entries and emits domain events for downstream consumers.
// Event publish failures are logged but swallowed — audit persistence is the critical path, not event delivery.
type CreateAuditLogHandler struct {
	repo     domain.AuditLogRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewCreateAuditLogHandler wires dependencies for audit log creation.
func NewCreateAuditLogHandler(
	repo domain.AuditLogRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *CreateAuditLogHandler {
	return &CreateAuditLogHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle constructs the audit log aggregate from the command, persists it, and publishes domain events.
// Returns nil on success; propagates repository errors to the caller.
func (h *CreateAuditLogHandler) Handle(ctx context.Context, cmd CreateAuditLogCommand) error {
	auditLog := domain.NewAuditLog(
		cmd.UserID,
		cmd.SessionID,
		cmd.Action,
		cmd.ResourceType,
		cmd.ResourceID,
		cmd.Platform,
		cmd.IPAddress,
		cmd.UserAgent,
		cmd.Permission,
		cmd.PolicyID,
		cmd.Decision,
		cmd.Success,
		cmd.ErrorMessage,
		cmd.Metadata,
	)

	if err := h.repo.Save(ctx, auditLog); err != nil {
		h.logger.Errorf("failed to save audit log: %v", err)
		return err
	}

	if err := h.eventBus.Publish(ctx, auditLog.Events()...); err != nil {
		h.logger.Errorf("failed to publish audit log events: %v", err)
	}

	return nil
}
