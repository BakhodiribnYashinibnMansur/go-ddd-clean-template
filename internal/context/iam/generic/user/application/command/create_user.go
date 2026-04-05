package command

import (
	"context"
	"fmt"

	"gct/internal/kernel/application"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
	"gct/internal/context/iam/generic/user/domain"

	"github.com/google/uuid"
)

// CreateUserCommand represents an admin-initiated user creation (as opposed to self-registration via SignUp).
// Phone and Password are required; all other fields are optional enrichments.
// The password is supplied in raw form and will be hashed by the domain layer before persistence.
type CreateUserCommand struct {
	Phone      string
	Password   string
	Email      *string
	Username   *string
	RoleID     *uuid.UUID
	Attributes map[string]string
}

// CreateUserHandler orchestrates user creation with domain validation (phone format, email format, password strength).
// Domain events are published after a successful save; event bus failures are logged but do not roll back the write.
type CreateUserHandler struct {
	repo     domain.UserRepository
	eventBus application.EventBus
	logger   commandLogger
}

// NewCreateUserHandler creates a new CreateUserHandler.
func NewCreateUserHandler(
	repo domain.UserRepository,
	eventBus application.EventBus,
	logger commandLogger,
) *CreateUserHandler {
	return &CreateUserHandler{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Handle validates inputs through domain value objects, constructs the User aggregate, and persists it.
// Returns domain validation errors (invalid phone, weak password) or repository errors (duplicate phone/email).
func (h *CreateUserHandler) Handle(ctx context.Context, cmd CreateUserCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "CreateUserHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "CreateUser", "user")()

	phone, err := domain.NewPhone(cmd.Phone)
	if err != nil {
		return apperrors.MapToServiceError(err)
	}

	password, err := domain.NewPasswordFromRaw(cmd.Password)
	if err != nil {
		return apperrors.MapToServiceError(err)
	}

	var opts []domain.UserOption

	if cmd.Email != nil {
		email, err := domain.NewEmail(*cmd.Email)
		if err != nil {
			return apperrors.MapToServiceError(err)
		}
		opts = append(opts, domain.WithEmail(email))
	}

	if cmd.Username != nil {
		opts = append(opts, domain.WithUsername(*cmd.Username))
	}

	if cmd.RoleID != nil {
		opts = append(opts, domain.WithRoleID(*cmd.RoleID))
	}

	if cmd.Attributes != nil {
		opts = append(opts, domain.WithAttributes(cmd.Attributes))
	}

	user, err := domain.NewUser(phone, password, opts...)
	if err != nil {
		return fmt.Errorf("create_user: %w", err)
	}

	if err := h.repo.Save(ctx, user); err != nil {
		h.logger.Errorc(ctx, "repository save failed", logger.F{Op: "CreateUser", Entity: "user", Err: err}.KV()...)
		return apperrors.MapToServiceError(err)
	}

	if err := h.eventBus.Publish(ctx, user.Events()...); err != nil {
		h.logger.Warnc(ctx, "event publish failed", logger.F{Op: "CreateUser", Entity: "user", Err: err}.KV()...)
	}

	return nil
}
