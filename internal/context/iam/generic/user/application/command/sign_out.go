package command

import (
	"context"
	"time"

	userentity "gct/internal/context/iam/generic/user/domain/entity"
	userrepo "gct/internal/context/iam/generic/user/domain/repository"
	"gct/internal/kernel/application"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
	"gct/internal/kernel/infrastructure/security/audit"
	"gct/internal/kernel/infrastructure/security/revocation"

	"github.com/google/uuid"
)

// SignOutCommand holds the input for user sign-out.
type SignOutCommand struct {
	UserID    userentity.UserID
	SessionID userentity.SessionID
	IP        string
	UserAgent string
}

// SignOutHandler handles the SignOutCommand.
type SignOutHandler struct {
	repo        userrepo.UserRepository
	eventBus    application.EventBus
	logger      commandLogger
	auditLogger audit.Logger
	revStore    *revocation.Store
}

// NewSignOutHandler creates a new SignOutHandler.
func NewSignOutHandler(
	repo userrepo.UserRepository,
	eventBus application.EventBus,
	logger commandLogger,
) *SignOutHandler {
	return &SignOutHandler{
		repo:        repo,
		eventBus:    eventBus,
		logger:      logger,
		auditLogger: audit.NoopLogger{},
	}
}

// WithSecurityDeps injects Phase S1 security dependencies.
func (h *SignOutHandler) WithSecurityDeps(al audit.Logger, rs *revocation.Store) *SignOutHandler {
	if al != nil {
		h.auditLogger = al
	}
	h.revStore = rs
	return h
}

// defaultRevocationTTL is the safe fallback TTL for revocation entries when
// the remaining access-token lifetime is unknown.
const defaultRevocationTTL = 15 * time.Minute

// Handle executes the SignOutCommand.
func (h *SignOutHandler) Handle(ctx context.Context, cmd SignOutCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "SignOutHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "SignOut", "user")()

	user, err := h.repo.FindByID(ctx, cmd.UserID)
	if err != nil {
		return apperrors.MapToServiceError(err)
	}

	if err := user.RevokeSession(cmd.SessionID.UUID()); err != nil {
		return apperrors.MapToServiceError(err)
	}

	if err := h.repo.Update(ctx, user); err != nil {
		h.logger.Errorc(ctx, "repository update failed", logger.F{Op: "SignOut", Entity: "user", EntityID: cmd.UserID, Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	// Phase S1: revoke session in Redis so in-flight access tokens are
	// immediately rejected without waiting for the DB to propagate.
	if h.revStore != nil {
		if rErr := h.revStore.Revoke(ctx, cmd.SessionID.UUID().String(), defaultRevocationTTL); rErr != nil {
			h.logger.Warnc(ctx, "revocation store write failed",
				logger.F{Op: "SignOut", Entity: "user", Err: rErr}.KV()...)
		}
	}

	sid := uuid.UUID(cmd.SessionID)
	uid := uuid.UUID(cmd.UserID)
	h.auditLogger.Log(ctx, audit.Entry{
		Event: audit.EventSessionRevoked, IPAddress: cmd.IP, UserAgent: cmd.UserAgent,
		SessionID: &sid, UserID: &uid,
	})

	return nil
}
