package command

import (
	"context"

	userentity "gct/internal/context/iam/generic/user/domain/entity"
	userevent "gct/internal/context/iam/generic/user/domain/event"
	userrepo "gct/internal/context/iam/generic/user/domain/repository"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
	"gct/internal/kernel/outbox"
)

// UpdateUserCommand represents a partial update to a user's profile fields.
// Pointer fields use nil-means-unchanged semantics. Phone, password, and role are excluded —
// use dedicated commands (ChangeRole, etc.) for those privileged mutations.
type UpdateUserCommand struct {
	ID         userentity.UserID
	Email      *string
	Username   *string
	Attributes map[string]string
}

// UpdateUserHandler applies partial profile updates via a load-reconstruct-save cycle.
// Because the User aggregate uses unexported fields, the handler reconstructs the entity with merged values.
type UpdateUserHandler struct {
	repo      userrepo.UserRepository
	committer *outbox.EventCommitter
	logger    commandLogger
}

// NewUpdateUserHandler creates a new UpdateUserHandler.
func NewUpdateUserHandler(
	repo userrepo.UserRepository,
	committer *outbox.EventCommitter,
	logger commandLogger,
) *UpdateUserHandler {
	return &UpdateUserHandler{
		repo:      repo,
		committer: committer,
		logger:    logger,
	}
}

// Handle loads the user, merges changed fields with existing data, reconstructs the aggregate, and persists it.
// Calls Touch() to update the modification timestamp. Returns domain or repository errors to the caller.
func (h *UpdateUserHandler) Handle(ctx context.Context, cmd UpdateUserCommand) (err error) {
	ctx, end := pgxutil.AppSpan(ctx, "UpdateUserHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "UpdateUser", "user")()

	user, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return apperrors.MapToServiceError(err)
	}

	// Build options for the updated fields and reconstruct.
	// Since the domain entity uses unexported fields, we reconstruct with
	// the updated values while preserving existing data.
	email := user.Email()
	if cmd.Email != nil {
		e, err := userentity.NewEmail(*cmd.Email)
		if err != nil {
			return apperrors.MapToServiceError(err)
		}
		email = &e
	}

	username := user.Username()
	if cmd.Username != nil {
		username = cmd.Username
	}

	attributes := user.Attributes()
	if cmd.Attributes != nil {
		attributes = cmd.Attributes
	}

	updated := userentity.ReconstructUser(
		user.ID(),
		user.CreatedAt(),
		user.UpdatedAt(),
		user.DeletedAt(),
		user.Phone(),
		email,
		username,
		user.Password(),
		user.RoleID(),
		attributes,
		user.IsActive(),
		user.IsApproved(),
		user.LastSeen(),
		user.Sessions(),
	)
	updated.Touch()
	updated.AddEvent(userevent.NewUserProfileUpdated(updated.ID()))

	return h.committer.Commit(ctx, func(ctx context.Context) error {
		if err := h.repo.Update(ctx, updated); err != nil {
			h.logger.Errorc(ctx, "repository update failed", logger.F{Op: "UpdateUser", Entity: "user", EntityID: cmd.ID, Err: err}.KV()...)
			return apperrors.MapToServiceError(err)
		}
		return nil
	}, updated.Events)
}
