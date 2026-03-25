package command

import (
	"context"

	"gct/internal/audit/domain"
	"gct/internal/shared/infrastructure/logger"

	"github.com/google/uuid"
)

// CreateEndpointHistoryCommand holds the input for creating a new endpoint history entry.
type CreateEndpointHistoryCommand struct {
	UserID     *uuid.UUID
	Endpoint   string
	Method     string
	StatusCode int
	Latency    int
	IPAddress  *string
	UserAgent  *string
}

// CreateEndpointHistoryHandler handles the CreateEndpointHistoryCommand.
type CreateEndpointHistoryHandler struct {
	repo   domain.EndpointHistoryRepository
	logger logger.Log
}

// NewCreateEndpointHistoryHandler creates a new CreateEndpointHistoryHandler.
func NewCreateEndpointHistoryHandler(
	repo domain.EndpointHistoryRepository,
	logger logger.Log,
) *CreateEndpointHistoryHandler {
	return &CreateEndpointHistoryHandler{
		repo:   repo,
		logger: logger,
	}
}

// Handle executes the CreateEndpointHistoryCommand.
func (h *CreateEndpointHistoryHandler) Handle(ctx context.Context, cmd CreateEndpointHistoryCommand) error {
	entry := domain.NewEndpointHistory(
		cmd.UserID,
		cmd.Endpoint,
		cmd.Method,
		cmd.StatusCode,
		cmd.Latency,
		cmd.IPAddress,
		cmd.UserAgent,
	)

	if err := h.repo.Save(ctx, entry); err != nil {
		h.logger.Errorf("failed to save endpoint history: %v", err)
		return err
	}

	return nil
}
