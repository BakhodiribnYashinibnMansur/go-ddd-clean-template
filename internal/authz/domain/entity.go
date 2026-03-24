package domain

import (
	"time"

	shared "gct/internal/shared/domain"

	"github.com/google/uuid"
)

// Role is the aggregate root for authorization roles.
type Role struct {
	shared.AggregateRoot
	name        string
	description *string
	permissions []Permission
}

// NewRole creates a new Role with a generated ID.
func NewRole(name string) *Role {
	r := &Role{
		AggregateRoot: shared.NewAggregateRoot(),
		name:          name,
		permissions:   make([]Permission, 0),
	}
	r.AddEvent(NewRoleCreated(r.ID(), name))
	return r
}

// ReconstructRole rebuilds a Role from persisted data.
func ReconstructRole(
	id uuid.UUID,
	createdAt, updatedAt time.Time,
	deletedAt *time.Time,
	name string,
	description *string,
	permissions []Permission,
) *Role {
	if permissions == nil {
		permissions = make([]Permission, 0)
	}
	return &Role{
		AggregateRoot: shared.NewAggregateRootWithID(id, createdAt, updatedAt, deletedAt),
		name:          name,
		description:   description,
		permissions:   permissions,
	}
}

// Name returns the role name.
func (r *Role) Name() string { return r.name }

// Description returns the role description.
func (r *Role) Description() *string { return r.description }

// Permissions returns the role's permissions.
func (r *Role) Permissions() []Permission { return r.permissions }

// Rename changes the role's name.
func (r *Role) Rename(name string) {
	r.name = name
	r.Touch()
}

// SetDescription updates the role's description.
func (r *Role) SetDescription(desc *string) {
	r.description = desc
	r.Touch()
}

// AddPermission adds a permission to the role.
func (r *Role) AddPermission(perm Permission) {
	r.permissions = append(r.permissions, perm)
	r.Touch()
	r.AddEvent(NewPermissionGranted(r.ID(), perm.ID()))
}

// RemovePermission removes a permission from the role by its ID.
func (r *Role) RemovePermission(permID uuid.UUID) error {
	for i, p := range r.permissions {
		if p.ID() == permID {
			r.permissions = append(r.permissions[:i], r.permissions[i+1:]...)
			r.Touch()
			return nil
		}
	}
	return ErrPermissionNotFound
}
