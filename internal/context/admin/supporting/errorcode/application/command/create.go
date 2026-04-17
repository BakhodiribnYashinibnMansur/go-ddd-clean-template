package command

import (
	"context"

	errcodeentity "gct/internal/context/admin/supporting/errorcode/domain/entity"
	errcoderepo "gct/internal/context/admin/supporting/errorcode/domain/repository"
	shareddomain "gct/internal/kernel/domain"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
	"gct/internal/kernel/outbox"
)

// CreateErrorCodeCommand represents an intent to register a new standardized error code in the system.
// Code is the stable identifier returned to API consumers (e.g., "AUTH_001"). Retryable and RetryAfter
// instruct clients whether and when to retry; Suggestion provides human-readable remediation guidance.
type CreateErrorCodeCommand struct {
	Code       string
	Message    string
	MessageUz  string
	MessageRu  string
	HTTPStatus int
	Category   string
	Severity   string
	Retryable  bool
	RetryAfter int
	Suggestion string
}

// CreateErrorCodeHandler orchestrates error code creation and emits domain events for downstream consumers.
type CreateErrorCodeHandler struct {
	repo      errcoderepo.ErrorCodeRepository
	committer *outbox.EventCommitter
	logger    logger.Log
}

// NewCreateErrorCodeHandler wires dependencies for error code creation.
func NewCreateErrorCodeHandler(
	repo errcoderepo.ErrorCodeRepository,
	committer *outbox.EventCommitter,
	logger logger.Log,
) *CreateErrorCodeHandler {
	return &CreateErrorCodeHandler{
		repo:      repo,
		committer: committer,
		logger:    logger,
	}
}

// Handle persists a new error code and publishes its domain events.
// Returns nil on success; propagates repository errors (e.g., duplicate code) to the caller.
func (h *CreateErrorCodeHandler) Handle(ctx context.Context, cmd CreateErrorCodeCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "CreateErrorCodeHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "CreateErrorCode", "error_code")()

	ec := errcodeentity.NewErrorCode(
		cmd.Code, cmd.Message, cmd.HTTPStatus,
		cmd.Category, cmd.Severity,
		cmd.Retryable, cmd.RetryAfter, cmd.Suggestion,
	)
	ec.SetTranslations(cmd.MessageUz, cmd.MessageRu)

	return h.committer.Commit(ctx, func(ctx context.Context, q shareddomain.Querier) error {
		if err := h.repo.Save(ctx, q, ec); err != nil {
			h.logger.Errorc(ctx, "repository save failed", logger.F{Op: "CreateErrorCode", Entity: "error_code", EntityID: cmd.Code, Err: err}.KV()...)
			return apperrors.MapToServiceError(err)
		}
		return nil
	}, ec.Events)
}
