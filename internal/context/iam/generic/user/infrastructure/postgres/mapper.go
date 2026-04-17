package postgres

import (
	"time"

	userentity "gct/internal/context/iam/generic/user/domain/entity"

	"github.com/google/uuid"
)

// userInsertValues holds the raw column values used to INSERT or UPDATE
// the users row. Centralising this unwrapping keeps write_repo.go free of
// VO-to-primitive boilerplate.
type userInsertValues struct {
	id           uuid.UUID
	roleID       *uuid.UUID
	username     *string
	email        *string
	phone        string
	passwordHash string
	active       bool
	isApproved   bool
	createdAt    time.Time
	updatedAt    time.Time
	deletedAt    int64
	lastSeen     *time.Time
}

// userToInsertValues extracts the primitive column values from a User
// aggregate. VOs are unwrapped once here and reused by Save/Update.
func userToInsertValues(u *userentity.User) userInsertValues {
	var email *string
	if u.Email() != nil {
		v := u.Email().Value()
		email = &v
	}

	var deletedAt int64
	if u.DeletedAt() != nil {
		deletedAt = u.DeletedAt().Unix()
	}

	return userInsertValues{
		id:           u.ID(),
		roleID:       u.RoleID(),
		username:     u.Username(),
		email:        email,
		phone:        u.Phone().Value(),
		passwordHash: u.Password().Hash(),
		active:       u.IsActive(),
		isApproved:   u.IsApproved(),
		createdAt:    u.CreatedAt(),
		updatedAt:    u.UpdatedAt(),
		deletedAt:    deletedAt,
		lastSeen:     u.LastSeen(),
	}
}
