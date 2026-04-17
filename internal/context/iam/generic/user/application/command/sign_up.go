package command

import (
	"context"
	"fmt"

	userentity "gct/internal/context/iam/generic/user/domain/entity"
	userrepo "gct/internal/context/iam/generic/user/domain/repository"
	"gct/internal/kernel/application"
	shareddomain "gct/internal/kernel/domain"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/metrics"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// SignUpCommand holds the input for user self-registration.
type SignUpCommand struct {
	Phone    string
	Password string
	Username *string
	Email    *string
}

// SignUpHandler handles the SignUpCommand.
type SignUpHandler struct {
	repo     userrepo.UserRepository
	readRepo userrepo.UserReadRepository
	db       shareddomain.DB
	eventBus application.EventBus
	logger   commandLogger
	bm       *metrics.BusinessMetrics
}

// NewSignUpHandler creates a new SignUpHandler.
func NewSignUpHandler(
	repo userrepo.UserRepository,
	readRepo userrepo.UserReadRepository,
	db shareddomain.DB,
	eventBus application.EventBus,
	logger commandLogger,
) *SignUpHandler {
	return &SignUpHandler{
		repo:     repo,
		readRepo: readRepo,
		db:       db,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle executes the SignUpCommand.
// The user is created as active but NOT approved by default.
func (h *SignUpHandler) Handle(ctx context.Context, cmd SignUpCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "SignUpHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "SignUp", "user")()

	phone, err := userentity.NewPhone(cmd.Phone)
	if err != nil {
		return apperrors.MapToServiceError(err)
	}

	password, err := userentity.NewPasswordFromRaw(cmd.Password)
	if err != nil {
		return apperrors.MapToServiceError(err)
	}

	var opts []userentity.UserOption

	if cmd.Email != nil {
		email, err := userentity.NewEmail(*cmd.Email)
		if err != nil {
			return apperrors.MapToServiceError(err)
		}
		opts = append(opts, userentity.WithEmail(email))
	}

	if cmd.Username != nil {
		opts = append(opts, userentity.WithUsername(*cmd.Username))
	}

	// Self-registration: user is auto-approved with default role.
	user, err := userentity.NewUser(phone, password, opts...)
	if err != nil {
		return fmt.Errorf("create_user: %w", err)
	}
	user.Approve()

	if defaultRoleID, err := h.readRepo.FindDefaultRoleID(ctx); err == nil {
		user.ChangeRole(defaultRoleID)
	}

	if err := pgxutil.WithTx(ctx, h.db, func(q shareddomain.Querier) error {
		return h.repo.Save(ctx, q, user)
	}); err != nil {
		h.logger.Errorc(ctx, "repository save failed", logger.F{Op: "SignUp", Entity: "user", Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	if err := h.eventBus.Publish(ctx, user.Events()...); err != nil {
		h.logger.Warnc(ctx, "event publish failed", logger.F{Op: "SignUp", Entity: "user", Err: err}.KV()...)
	}

	h.bm.Inc(ctx, "user_signups")

	return nil
}

// WithBusinessMetrics injects business metrics into the handler.
func (h *SignUpHandler) WithBusinessMetrics(bm *metrics.BusinessMetrics) {
	h.bm = bm
}
