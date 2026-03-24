package domain

import (
	"time"

	"gct/internal/shared/infrastructure/validation"

	"github.com/google/uuid"
)

// Role represents a user role.
type Role struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// RoleFilter represents filter for role queries
type RoleFilter struct {
	ID   *uuid.UUID `json:"id,omitempty"`
	Name *string    `json:"name,omitempty"`
}

// RolesFilter represents filter for multiple roles with pagination
type RolesFilter struct {
	RoleFilter
	Pagination *Pagination `json:"pagination,omitempty"`
}

// Permission represents a system permission.
type Permission struct {
	ID          uuid.UUID  `json:"id"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// PermissionFilter represents filter for permission queries
type PermissionFilter struct {
	ID   *uuid.UUID `json:"id,omitempty"`
	Name *string    `json:"name,omitempty"`
}

// PermissionsFilter represents filter for multiple permissions with pagination
type PermissionsFilter struct {
	PermissionFilter
	Pagination *Pagination `json:"pagination,omitempty"`
}

// PolicyEffect enum
type PolicyEffect string

const (
	PolicyEffectAllow PolicyEffect = "ALLOW"
	PolicyEffectDeny  PolicyEffect = "DENY"
)

func (e PolicyEffect) IsValid() bool {
	return validation.IsEnumValid(e, []PolicyEffect{
		PolicyEffectAllow,
		PolicyEffectDeny,
	})
}

// Policy represents an ABAC policy
type Policy struct {
	ID           uuid.UUID      `json:"id"`
	PermissionID uuid.UUID      `json:"permission_id"`
	Effect       PolicyEffect   `json:"effect"`
	Priority     int            `json:"priority"`
	Active       bool           `json:"active"`
	Conditions   map[string]any `json:"conditions"`
	CreatedAt    time.Time      `json:"created_at"`
}

// PolicyFilter represents filter for policy queries
type PolicyFilter struct {
	ID           *uuid.UUID `json:"id,omitempty"`
	PermissionID *uuid.UUID `json:"permission_id,omitempty"`
	Active       *bool      `json:"active,omitempty"`
}

// PoliciesFilter represents filter for multiple policies with pagination
type PoliciesFilter struct {
	PolicyFilter
	Pagination *Pagination `json:"pagination,omitempty"`
}

// RelationType enum
type RelationType string

const (
	RelationTypeUnrevealed RelationType = "UNREVEALED"
	RelationTypeBranch     RelationType = "BRANCH"
	RelationTypeRegion     RelationType = "REGION"
)

func (r RelationType) IsValid() bool {
	return validation.IsEnumValid(r, []RelationType{
		RelationTypeUnrevealed,
		RelationTypeBranch,
		RelationTypeRegion,
	})
}

// Relation represents an organizational relation (branch/region).
type Relation struct {
	ID        uuid.UUID    `json:"id"`
	Type      RelationType `json:"type"`
	Name      string       `json:"name"`
	CreatedAt time.Time    `json:"created_at"`
}

// RelationFilter represents filter for relation queries
type RelationFilter struct {
	ID   *uuid.UUID `json:"id,omitempty"`
	Type *string    `json:"type,omitempty"`
	Name *string    `json:"name,omitempty"`
}

// RelationsFilter represents filter for multiple relations with pagination
type RelationsFilter struct {
	RelationFilter
	Pagination *Pagination `json:"pagination,omitempty"`
}

// Scope represents an API scope.
type Scope struct {
	Path      string    `json:"path"`
	Method    string    `json:"method"`
	CreatedAt time.Time `json:"created_at"`
}

// ScopeFilter for queries
type ScopeFilter struct {
	Path   *string `json:"path,omitempty"`
	Method *string `json:"method,omitempty"`
}

// ScopesFilter for listing
type ScopesFilter struct {
	ScopeFilter
	Pagination *Pagination `json:"pagination,omitempty"`
}
