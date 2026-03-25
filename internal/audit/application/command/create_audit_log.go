package command

import (
	"context"

	"gct/internal/audit/domain"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"

	"github.com/google/uuid"
)

// CreateAuditLogCommand holds the input for creating a new audit log entry.
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
	Metadata     map[string]any
}

// CreateAuditLogHandler handles the CreateAuditLogCommand.
type CreateAuditLogHandler struct {
	repo     domain.AuditLogRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewCreateAuditLogHandler creates a new CreateAuditLogHandler.
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

// Handle executes the CreateAuditLogCommand.
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
