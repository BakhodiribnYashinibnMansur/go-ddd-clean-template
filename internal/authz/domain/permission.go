package domain

import (
	"time"

	shared "gct/internal/shared/domain"

	"github.com/google/uuid"
)

// Permission is a child entity of Role.
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

// RemoveScope removes a scope from the permission by path and method.
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
