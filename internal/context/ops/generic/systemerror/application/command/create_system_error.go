package command

import (
	"context"

	shareddomain "gct/internal/kernel/domain"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
	"gct/internal/kernel/outbox"
	syserrentity "gct/internal/context/ops/generic/systemerror/domain/entity"
	syserrrepo "gct/internal/context/ops/generic/systemerror/domain/repository"

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
	repo      syserrrepo.SystemErrorRepository
	committer *outbox.EventCommitter
	logger    logger.Log
}

// NewCreateSystemErrorHandler creates a new CreateSystemErrorHandler.
func NewCreateSystemErrorHandler(
	repo syserrrepo.SystemErrorRepository,
	committer *outbox.EventCommitter,
	logger logger.Log,
) *CreateSystemErrorHandler {
	return &CreateSystemErrorHandler{
		repo:      repo,
		committer: committer,
		logger:    logger,
	}
}

// Handle constructs a SystemError aggregate from the command, enriches it with optional context fields,
// persists it, and publishes domain events. Returns only repository-level errors.
func (h *CreateSystemErrorHandler) Handle(ctx context.Context, cmd CreateSystemErrorCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "CreateSystemErrorHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "CreateSystemError", "system_error")()

	se := syserrentity.NewSystemError(cmd.Code, cmd.Message, cmd.Severity)

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

	return h.committer.Commit(ctx, func(ctx context.Context, q shareddomain.Querier) error {
		if err := h.repo.Save(ctx, q, se); err != nil {
			h.logger.Errorc(ctx, "repository save failed", logger.F{Op: "CreateSystemError", Entity: "system_error", Err: err}.KV()...)
			return apperrors.MapToServiceError(err)
		}
		return nil
	}, se.Events)
}
