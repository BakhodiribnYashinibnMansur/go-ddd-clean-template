package mock

import (
	"time"

	"gct/internal/domain"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
)

// Role generates a fake domain.Role
func Role() *domain.Role {
	return &domain.Role{
		ID:        UUID(),
		Name:      gofakeit.JobTitle(),
		CreatedAt: time.Now(),
	}
}

// Roles generates multiple fake domain.Role
func Roles(count int) []*domain.Role {
	roles := make([]*domain.Role, count)
	for i := range count {
		roles[i] = Role()
	}
	return roles
}

// RoleWithID generates a fake domain.Role with specific ID
func RoleWithID(id uuid.UUID) *domain.Role {
	role := Role()
	role.ID = id
	return role
}

// RoleWithName generates a fake domain.Role with specific name
func RoleWithName(name string) *domain.Role {
	role := Role()
	role.Name = name
	return role
}

// RoleFilter generates a fake domain.RoleFilter
func RoleFilter() *domain.RoleFilter {
	id := UUID()
	name := gofakeit.JobTitle()
	return &domain.RoleFilter{
		ID:   &id,
		Name: &name,
	}
}

// RoleFilterWithID generates a domain.RoleFilter with specific ID
func RoleFilterWithID(id uuid.UUID) *domain.RoleFilter {
	return &domain.RoleFilter{
		ID: &id,
	}
}

// RoleFilterWithName generates a domain.RoleFilter with specific name
func RoleFilterWithName(name string) *domain.RoleFilter {
	return &domain.RoleFilter{
		Name: &name,
	}
}

// RolesFilter generates a fake domain.RolesFilter
func RolesFilter() *domain.RolesFilter {
	return &domain.RolesFilter{
		RoleFilter: *RoleFilter(),
		Pagination: Pagination(),
	}
}

// RolesFilterWithPagination generates a fake domain.RolesFilter with custom pagination
func RolesFilterWithPagination(limit, offset, total int64) *domain.RolesFilter {
	return &domain.RolesFilter{
		RoleFilter: *RoleFilter(),
		Pagination: PaginationWithValues(limit, offset, total),
	}
}

// Permission generates a fake domain.Permission
func Permission() *domain.Permission {
	parentID := func() *uuid.UUID { id := UUID(); return &id }()
	if gofakeit.Bool() {
		parentID = nil // Root permission
	}

	return &domain.Permission{
		ID:        UUID(),
		ParentID:  parentID,
		Name:      gofakeit.JobDescriptor(),
		CreatedAt: time.Now(),
	}
}

// Permissions generates multiple fake domain.Permission
func Permissions(count int) []*domain.Permission {
	permissions := make([]*domain.Permission, count)
	for i := range count {
		permissions[i] = Permission()
	}
	return permissions
}

// PermissionWithID generates a fake domain.Permission with specific ID
func PermissionWithID(id uuid.UUID) *domain.Permission {
	permission := Permission()
	permission.ID = id
	return permission
}

// PermissionWithParentID generates a fake domain.Permission with specific parent ID
func PermissionWithParentID(parentID *uuid.UUID) *domain.Permission {
	permission := Permission()
	permission.ParentID = parentID
	return permission
}

// PermissionRoot generates a fake root domain.Permission (no parent)
func PermissionRoot() *domain.Permission {
	permission := Permission()
	permission.ParentID = nil
	return permission
}

// PermissionFilter generates a fake domain.PermissionFilter
func PermissionFilter() *domain.PermissionFilter {
	id := UUID()
	name := gofakeit.JobDescriptor()
	return &domain.PermissionFilter{
		ID:   &id,
		Name: &name,
	}
}

// PermissionFilterWithID generates a domain.PermissionFilter with specific ID
func PermissionFilterWithID(id uuid.UUID) *domain.PermissionFilter {
	return &domain.PermissionFilter{
		ID: &id,
	}
}

// PermissionFilterWithName generates a domain.PermissionFilter with specific name
func PermissionFilterWithName(name string) *domain.PermissionFilter {
	return &domain.PermissionFilter{
		Name: &name,
	}
}

// PermissionsFilter generates a fake domain.PermissionsFilter
func PermissionsFilter() *domain.PermissionsFilter {
	return &domain.PermissionsFilter{
		PermissionFilter: *PermissionFilter(),
		Pagination:       Pagination(),
	}
}

// PermissionsFilterWithPagination generates a fake domain.PermissionsFilter with custom pagination
func PermissionsFilterWithPagination(limit, offset, total int64) *domain.PermissionsFilter {
	return &domain.PermissionsFilter{
		PermissionFilter: *PermissionFilter(),
		Pagination:       PaginationWithValues(limit, offset, total),
	}
}

// PolicyEffect generates a fake domain.PolicyEffect
func PolicyEffect() domain.PolicyEffect {
	effects := []domain.PolicyEffect{
		domain.PolicyEffectAllow,
		domain.PolicyEffectDeny,
	}
	return effects[gofakeit.IntRange(0, len(effects)-1)]
}

// PolicyEffectAllow generates domain.PolicyEffectAllow
func PolicyEffectAllow() domain.PolicyEffect {
	return domain.PolicyEffectAllow
}

// PolicyEffectDeny generates domain.PolicyEffectDeny
func PolicyEffectDeny() domain.PolicyEffect {
	return domain.PolicyEffectDeny
}

// Policy generates a fake domain.Policy
func Policy() *domain.Policy {
	return &domain.Policy{
		ID:           UUID(),
		PermissionID: UUID(),
		Effect:       PolicyEffect(),
		Priority:     gofakeit.IntRange(1, 100),
		Active:       gofakeit.Bool(),
		Conditions: map[string]any{
			"ip":          gofakeit.IPv4Address(),
			"time":        gofakeit.Date().Format("15:04"),
			"day_of_week": gofakeit.WeekDay(),
		},
		CreatedAt: time.Now(),
	}
}

// Policies generates multiple fake domain.Policy
func Policies(count int) []*domain.Policy {
	policies := make([]*domain.Policy, count)
	for i := range count {
		policies[i] = Policy()
	}
	return policies
}

// PolicyWithID generates a fake domain.Policy with specific ID
func PolicyWithID(id uuid.UUID) *domain.Policy {
	policy := Policy()
	policy.ID = id
	return policy
}

// PolicyWithPermissionID generates a fake domain.Policy with specific permission ID
func PolicyWithPermissionID(permissionID uuid.UUID) *domain.Policy {
	policy := Policy()
	policy.PermissionID = permissionID
	return policy
}

// PolicyWithEffect generates a fake domain.Policy with specific effect
func PolicyWithEffect(effect domain.PolicyEffect) *domain.Policy {
	policy := Policy()
	policy.Effect = effect
	return policy
}

// PolicyActive generates a fake active domain.Policy
func PolicyActive() *domain.Policy {
	policy := Policy()
	policy.Active = true
	return policy
}

// PolicyInactive generates a fake inactive domain.Policy
func PolicyInactive() *domain.Policy {
	policy := Policy()
	policy.Active = false
	return policy
}

// PolicyFilter generates a fake domain.PolicyFilter
func PolicyFilter() *domain.PolicyFilter {
	id := UUID()
	permissionID := UUID()
	active := gofakeit.Bool()
	return &domain.PolicyFilter{
		ID:           &id,
		PermissionID: &permissionID,
		Active:       &active,
	}
}

// PolicyFilterWithID generates a domain.PolicyFilter with specific ID
func PolicyFilterWithID(id uuid.UUID) *domain.PolicyFilter {
	return &domain.PolicyFilter{
		ID: &id,
	}
}

// PolicyFilterWithPermissionID generates a domain.PolicyFilter with specific permission ID
func PolicyFilterWithPermissionID(permissionID uuid.UUID) *domain.PolicyFilter {
	return &domain.PolicyFilter{
		PermissionID: &permissionID,
	}
}

// PolicyFilterWithActive generates a domain.PolicyFilter with specific active status
func PolicyFilterWithActive(active bool) *domain.PolicyFilter {
	return &domain.PolicyFilter{
		Active: &active,
	}
}

// PoliciesFilter generates a fake domain.PoliciesFilter
func PoliciesFilter() *domain.PoliciesFilter {
	return &domain.PoliciesFilter{
		PolicyFilter: *PolicyFilter(),
		Pagination:   Pagination(),
	}
}

// PoliciesFilterWithPagination generates a fake domain.PoliciesFilter with custom pagination
func PoliciesFilterWithPagination(limit, offset, total int64) *domain.PoliciesFilter {
	return &domain.PoliciesFilter{
		PolicyFilter: *PolicyFilter(),
		Pagination:   PaginationWithValues(limit, offset, total),
	}
}

// Relation generates a fake domain.Relation
func Relation() *domain.Relation {
	return &domain.Relation{
		ID:        UUID(),
		Type:      domain.RelationTypeBranch,
		Name:      gofakeit.Company(),
		CreatedAt: time.Now(),
	}
}

// Relations generates multiple fake domain.Relation
func Relations(count int) []*domain.Relation {
	relations := make([]*domain.Relation, count)
	for i := range count {
		relations[i] = Relation()
	}
	return relations
}

// Scope generates a fake domain.Scope
func Scope() *domain.Scope {
	return &domain.Scope{
		Path:      gofakeit.URL(),
		Method:    gofakeit.HTTPMethod(),
		CreatedAt: time.Now(),
	}
}

// Scopes generates multiple fake domain.Scope
func Scopes(count int) []*domain.Scope {
	scopes := make([]*domain.Scope, count)
	for i := range count {
		scopes[i] = Scope()
	}
	return scopes
}
