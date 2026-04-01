package command

import (
	"context"

	"gct/internal/audit/domain"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/shared/infrastructure/pgxutil"

	"github.com/google/uuid"
)

// CreateEndpointHistoryCommand records a single HTTP request for API usage analytics and latency tracking.
// UserID is nil for unauthenticated requests. Latency is stored in milliseconds.
type CreateEndpointHistoryCommand struct {
	UserID     *uuid.UUID
	Endpoint   string
	Method     string
	StatusCode int
	Latency    int
	IPAddress  *string
	UserAgent  *string
}

// CreateEndpointHistoryHandler persists endpoint history entries for observability.
// Unlike audit logs, no domain events are emitted — this is a fire-and-forget telemetry record.
type CreateEndpointHistoryHandler struct {
	repo   domain.EndpointHistoryRepository
	logger logger.Log
}

// NewCreateEndpointHistoryHandler wires dependencies for endpoint history recording.
func NewCreateEndpointHistoryHandler(
	repo domain.EndpointHistoryRepository,
	logger logger.Log,
) *CreateEndpointHistoryHandler {
	return &CreateEndpointHistoryHandler{
		repo:   repo,
		logger: logger,
	}
}

// Handle persists the endpoint history entry.
// Returns nil on success; propagates repository errors to the caller.
func (h *CreateEndpointHistoryHandler) Handle(ctx context.Context, cmd CreateEndpointHistoryCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "CreateEndpointHistoryHandler.Handle")
	defer func() { end(err) }()

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
