package domain

import (
	"time"

	shared "gct/internal/platform/domain"

	"github.com/google/uuid"
)

// Permission is a child entity of Role that groups a set of API scopes under a named capability.
// Permissions form a tree via parentID — a nil parentID indicates a root-level permission.
// Scopes are managed as an ordered slice; duplicates are the caller's responsibility to prevent.
type Permission struct {
	shared.BaseEntity
	parentID    *uuid.UUID
	name        string
	description *string
	scopes      []Scope
}

// NewPermission creates a new Permission with a generated ID.
func NewPermission(name string, parentID *uuid.UUID) *Permission {
	return &Permission{
		BaseEntity: shared.NewBaseEntity(),
		parentID:   parentID,
		name:       name,
		scopes:     make([]Scope, 0),
	}
}

// ReconstructPermission rebuilds a Permission from persisted data.
func ReconstructPermission(
	id uuid.UUID,
	createdAt, updatedAt time.Time,
	deletedAt *time.Time,
	parentID *uuid.UUID,
	name string,
	description *string,
	scopes []Scope,
) *Permission {
	if scopes == nil {
		scopes = make([]Scope, 0)
	}
	return &Permission{
		BaseEntity:  shared.NewBaseEntityWithID(id, createdAt, updatedAt, deletedAt),
		parentID:    parentID,
		name:        name,
		description: description,
		scopes:      scopes,
	}
}

// ParentID returns the parent permission ID.
func (p *Permission) ParentID() *uuid.UUID { return p.parentID }

// Name returns the permission name.
func (p *Permission) Name() string { return p.name }

// Description returns the permission description.
func (p *Permission) Description() *string { return p.description }

// Scopes returns the permission's scopes.
func (p *Permission) Scopes() []Scope { return p.scopes }

// Rename changes the permission name.
func (p *Permission) Rename(name string) {
	p.name = name
	p.Touch()
}

// SetDescription updates the permission description.
func (p *Permission) SetDescription(desc *string) {
	p.description = desc
	p.Touch()
}

// AddScope adds a scope to the permission.
func (p *Permission) AddScope(scope Scope) {
	p.scopes = append(p.scopes, scope)
	p.Touch()
}

// RemoveScope removes a scope identified by its path+method composite key.
// Returns ErrScopeNotFound if no matching scope exists in the permission's scope list.
func (p *Permission) RemoveScope(path, method string) error {
	for i, s := range p.scopes {
		if s.Path == path && s.Method == method {
			p.scopes = append(p.scopes[:i], p.scopes[i+1:]...)
			p.Touch()
			return nil
		}
	}
	return ErrScopeNotFound
}
