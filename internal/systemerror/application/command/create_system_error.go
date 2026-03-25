package command

import (
	"context"

	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/systemerror/domain"

	"github.com/google/uuid"
)

// CreateSystemErrorCommand holds the input for recording a new system error.
type CreateSystemErrorCommand struct {
	Code        string
	Message     string
	StackTrace  *string
	Metadata    map[string]any
	Severity    string
	ServiceName *string
	RequestID   *uuid.UUID
	UserID      *uuid.UUID
	IPAddress   *string
	Path        *string
	Method      *string
}

// CreateSystemErrorHandler handles the CreateSystemErrorCommand.
type CreateSystemErrorHandler struct {
	repo     domain.SystemErrorRepository
	eventBus application.EventBus
	logger   logger.Log
}

// NewCreateSystemErrorHandler creates a new CreateSystemErrorHandler.
func NewCreateSystemErrorHandler(
	repo domain.SystemErrorRepository,
	eventBus application.EventBus,
	logger logger.Log,
) *CreateSystemErrorHandler {
	return &CreateSystemErrorHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle executes the CreateSystemErrorCommand.
func (h *CreateSystemErrorHandler) Handle(ctx context.Context, cmd CreateSystemErrorCommand) error {
	se := domain.NewSystemError(cmd.Code, cmd.Message, cmd.Severity)

	if cmd.StackTrace != nil {
		se.SetStackTrace(cmd.StackTrace)
	}
	if cmd.Metadata != nil {
		se.SetMetadata(cmd.Metadata)
	}
	if cmd.ServiceName != nil {
		se.SetServiceName(cmd.ServiceName)
	}
	if cmd.RequestID != nil {
		se.SetRequestID(cmd.RequestID)
	}
	if cmd.UserID != nil {
		se.SetUserID(cmd.UserID)
	}
	if cmd.IPAddress != nil {
		se.SetIPAddress(cmd.IPAddress)
	}
	if cmd.Path != nil {
		se.SetPath(cmd.Path)
	}
	if cmd.Method != nil {
		se.SetMethod(cmd.Method)
	}

	if err := h.repo.Save(ctx, se); err != nil {
		h.logger.Errorf("failed to save system error: %v", err)
		return err
	}

	if err := h.eventBus.Publish(ctx, se.Events()...); err != nil {
		h.logger.Errorf("failed to publish events: %v", err)
	}

	return nil
}
