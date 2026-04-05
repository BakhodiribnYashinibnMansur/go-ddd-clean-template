package command

import (
	"context"

	"gct/internal/platform/application"
	apperrors "gct/internal/platform/infrastructure/errors"
	"gct/internal/platform/infrastructure/logger"
	"gct/internal/platform/infrastructure/pgxutil"
	"gct/internal/context/ops/systemerror/domain"

	"github.com/google/uuid"
)

// CreateSystemErrorCommand captures a structured error record for observability and incident triage.
// Code and Severity are required; all other fields are optional context enrichments.
// Pointer fields use nil-means-absent semantics — the handler only sets them on the aggregate when non-nil.
type CreateSystemErrorCommand struct {
	Code        string
	Message     string
	StackTrace  *string
	Metadata    map[string]string
	Severity    string
	ServiceName *string
	RequestID   *uuid.UUID
	UserID      *uuid.UUID
	IPAddress   *string
	Path        *string
	Method      *string
}

// CreateSystemErrorHandler persists system error records and emits domain events for downstream alerting.
// Event bus failures are logged but swallowed so that error recording is never blocked by event delivery.
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

// Handle constructs a SystemError aggregate from the command, enriches it with optional context fields,
// persists it, and publishes domain events. Returns only repository-level errors.
func (h *CreateSystemErrorHandler) Handle(ctx context.Context, cmd CreateSystemErrorCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "CreateSystemErrorHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "CreateSystemError", "system_error")()

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
		h.logger.Errorc(ctx, "repository save failed", logger.F{Op: "CreateSystemError", Entity: "system_error", Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	if err := h.eventBus.Publish(ctx, se.Events()...); err != nil {
		h.logger.Warnc(ctx, "event publish failed", logger.F{Op: "CreateSystemError", Entity: "system_error", Err: err}.KV()...)
	}

	return nil
}
