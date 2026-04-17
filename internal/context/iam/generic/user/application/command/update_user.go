package command

import (
	"context"
	"fmt"
	"sort"

	userentity "gct/internal/context/iam/generic/user/domain/entity"
	userevent "gct/internal/context/iam/generic/user/domain/event"
	userrepo "gct/internal/context/iam/generic/user/domain/repository"
	shareddomain "gct/internal/kernel/domain"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
	"gct/internal/kernel/outbox"

	"github.com/google/uuid"
)

// UpdateUserCommand represents a partial update to a user's profile fields.
// Pointer fields use nil-means-unchanged semantics. Phone, password, and role are excluded —
// use dedicated commands (ChangeRole, etc.) for those privileged mutations.
type UpdateUserCommand struct {
	ID         userentity.UserID
	ActorID    uuid.UUID
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

	// Compute field-level changes before raising events.
	var changes []userevent.FieldChange

	if cmd.Email != nil {
		oldVal := ""
		if user.Email() != nil {
			oldVal = user.Email().Value()
		}
		newVal := ""
		if email != nil {
			newVal = email.Value()
		}
		if oldVal != newVal {
			changes = append(changes, userevent.FieldChange{FieldName: "email", OldValue: oldVal, NewValue: newVal})
		}
	}

	if cmd.Username != nil {
		oldVal := ""
		if user.Username() != nil {
			oldVal = *user.Username()
		}
		newVal := ""
		if username != nil {
			newVal = *username
		}
		if oldVal != newVal {
			changes = append(changes, userevent.FieldChange{FieldName: "username", OldValue: oldVal, NewValue: newVal})
		}
	}

	if cmd.Attributes != nil {
		changes = append(changes, diffAttributes(user.Attributes(), attributes)...)
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

	if len(changes) > 0 {
		updated.AddEvent(userevent.NewUserProfileUpdatedWithChanges(updated.ID(), cmd.ActorID, changes))
	}

	return h.committer.Commit(ctx, func(ctx context.Context, q shareddomain.Querier) error {
		if err := h.repo.Update(ctx, q, updated); err != nil {
			h.logger.Errorc(ctx, "repository update failed", logger.F{Op: "UpdateUser", Entity: "user", EntityID: cmd.ID, Err: err}.KV()...)
			return apperrors.MapToServiceError(err)
		}
		return nil
	}, updated.Events)
}

// diffAttributes compares old and new attribute maps, returning a FieldChange
// for every added, removed, or modified key. Keys are sorted for deterministic output.
func diffAttributes(old, new map[string]string) []userevent.FieldChange {
	allKeys := make(map[string]struct{})
	for k := range old {
		allKeys[k] = struct{}{}
	}
	for k := range new {
		allKeys[k] = struct{}{}
	}

	sorted := make([]string, 0, len(allKeys))
	for k := range allKeys {
		sorted = append(sorted, k)
	}
	sort.Strings(sorted)

	var changes []userevent.FieldChange
	for _, k := range sorted {
		oldVal := old[k]
		newVal := new[k]
		if oldVal != newVal {
			changes = append(changes, userevent.FieldChange{
				FieldName: fmt.Sprintf("attributes.%s", k),
				OldValue:  oldVal,
				NewValue:  newVal,
			})
		}
	}
	return changes
}
