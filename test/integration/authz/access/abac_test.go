package access

import (
	"context"
	"testing"

	"gct/internal/authz/application/command"
	"gct/internal/authz/application/query"
	"gct/internal/authz/domain"
	shared "gct/internal/shared/domain"
)

// ---------------------------------------------------------------------------
// ABAC integration tests
// ---------------------------------------------------------------------------

// TestIntegration_ABAC_RBACPassNoPolicies_Allowed verifies that when RBAC grants
// access and no policies exist, the request is allowed (no policies = RBAC sufficient).
func TestIntegration_ABAC_RBACPassNoPolicies_Allowed(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	seedRoleWithScope(t, bc, "viewer", "docs.read", "/api/v1/docs", "GET")

	roles, _ := bc.ListRoles.Handle(ctx, query.ListRolesQuery{Pagination: shared.Pagination{Limit: 10}})
	roleID := roles.Roles[0].ID

	allowed, err := bc.CheckAccess.Handle(ctx, query.CheckAccessQuery{
		RoleID:  roleID,
		Path:    "/api/v1/docs",
		Method:  "GET",
		EvalCtx: domain.EvaluationContext{Attrs: map[string]map[string]any{}},
	})
	if err != nil {
		t.Fatalf("CheckAccess: %v", err)
	}
	if !allowed {
		t.Error("expected access allowed when RBAC passes and no policies exist")
	}
}

// TestIntegration_ABAC_DenyPolicy_Denied verifies that a matching DENY policy
// blocks access even when RBAC would allow it.
func TestIntegration_ABAC_DenyPolicy_Denied(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	seedRoleWithScope(t, bc, "operator", "reports.read", "/api/v1/reports", "GET")

	roles, _ := bc.ListRoles.Handle(ctx, query.ListRolesQuery{Pagination: shared.Pagination{Limit: 10}})
	perms, _ := bc.ListPermissions.Handle(ctx, query.ListPermissionsQuery{Pagination: shared.Pagination{Limit: 10}})
	roleID := roles.Roles[0].ID
	permID := perms.Permissions[0].ID

	// Create DENY policy: deny when env.ip is in the blocked list.
	if err := bc.CreatePolicy.Handle(ctx, command.CreatePolicyCommand{
		PermissionID: permID,
		Effect:       "DENY",
		Priority:     10,
		Conditions:   map[string]any{"env.ip_in": []any{"192.168.1.100"}},
	}); err != nil {
		t.Fatalf("CreatePolicy: %v", err)
	}

	// Access with matching IP should be denied.
	allowed, err := bc.CheckAccess.Handle(ctx, query.CheckAccessQuery{
		RoleID: roleID,
		Path:   "/api/v1/reports",
		Method: "GET",
		EvalCtx: domain.EvaluationContext{Attrs: map[string]map[string]any{
			"env": {"ip": "192.168.1.100"},
		}},
	})
	if err != nil {
		t.Fatalf("CheckAccess: %v", err)
	}
	if allowed {
		t.Error("expected access denied when DENY policy condition matches")
	}
}

// TestIntegration_ABAC_DenyPolicy_DifferentIP_Allowed verifies that a DENY policy
// does not block access when the condition does not match.
func TestIntegration_ABAC_DenyPolicy_DifferentIP_Allowed(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	seedRoleWithScope(t, bc, "operator", "reports.read", "/api/v1/reports", "GET")

	roles, _ := bc.ListRoles.Handle(ctx, query.ListRolesQuery{Pagination: shared.Pagination{Limit: 10}})
	perms, _ := bc.ListPermissions.Handle(ctx, query.ListPermissionsQuery{Pagination: shared.Pagination{Limit: 10}})
	roleID := roles.Roles[0].ID
	permID := perms.Permissions[0].ID

	// Create DENY policy: deny when env.ip is in the blocked list.
	if err := bc.CreatePolicy.Handle(ctx, command.CreatePolicyCommand{
		PermissionID: permID,
		Effect:       "DENY",
		Priority:     10,
		Conditions:   map[string]any{"env.ip_in": []any{"192.168.1.100"}},
	}); err != nil {
		t.Fatalf("CreatePolicy: %v", err)
	}

	// Access with a different IP should be allowed (condition doesn't match).
	allowed, err := bc.CheckAccess.Handle(ctx, query.CheckAccessQuery{
		RoleID: roleID,
		Path:   "/api/v1/reports",
		Method: "GET",
		EvalCtx: domain.EvaluationContext{Attrs: map[string]map[string]any{
			"env": {"ip": "10.0.0.1"},
		}},
	})
	if err != nil {
		t.Fatalf("CheckAccess: %v", err)
	}
	if !allowed {
		t.Error("expected access allowed when DENY policy condition does not match")
	}
}

// TestIntegration_ABAC_AllowPolicy_Matches verifies that an ALLOW policy with a
// matching condition grants access. The role_name attribute is injected by CheckAccess
// from the database.
func TestIntegration_ABAC_AllowPolicy_Matches(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	seedRoleWithScope(t, bc, "auditor", "audit.read", "/api/v1/audit", "GET")

	roles, _ := bc.ListRoles.Handle(ctx, query.ListRolesQuery{Pagination: shared.Pagination{Limit: 10}})
	perms, _ := bc.ListPermissions.Handle(ctx, query.ListPermissionsQuery{Pagination: shared.Pagination{Limit: 10}})
	roleID := roles.Roles[0].ID
	permID := perms.Permissions[0].ID

	// Create ALLOW policy: allow when user.role_name equals "auditor".
	if err := bc.CreatePolicy.Handle(ctx, command.CreatePolicyCommand{
		PermissionID: permID,
		Effect:       "ALLOW",
		Priority:     10,
		Conditions:   map[string]any{"user.role_name": "auditor"},
	}); err != nil {
		t.Fatalf("CreatePolicy: %v", err)
	}

	allowed, err := bc.CheckAccess.Handle(ctx, query.CheckAccessQuery{
		RoleID:  roleID,
		Path:    "/api/v1/audit",
		Method:  "GET",
		EvalCtx: domain.EvaluationContext{Attrs: map[string]map[string]any{}},
	})
	if err != nil {
		t.Fatalf("CheckAccess: %v", err)
	}
	if !allowed {
		t.Error("expected access allowed when ALLOW policy condition matches via injected role_name")
	}
}

// TestIntegration_ABAC_RBACFail_PoliciesNotConsulted verifies that when RBAC denies
// access (wrong path), policies are never consulted and access is denied.
func TestIntegration_ABAC_RBACFail_PoliciesNotConsulted(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	seedRoleWithScope(t, bc, "narrow", "items.read", "/api/v1/items", "GET")

	roles, _ := bc.ListRoles.Handle(ctx, query.ListRolesQuery{Pagination: shared.Pagination{Limit: 10}})
	roleID := roles.Roles[0].ID

	// Try accessing a different path — RBAC should deny before policies are checked.
	allowed, err := bc.CheckAccess.Handle(ctx, query.CheckAccessQuery{
		RoleID:  roleID,
		Path:    "/api/v1/other",
		Method:  "GET",
		EvalCtx: domain.EvaluationContext{Attrs: map[string]map[string]any{}},
	})
	if err != nil {
		t.Fatalf("CheckAccess: %v", err)
	}
	if allowed {
		t.Error("expected access denied by RBAC for wrong path, policies should not be consulted")
	}
}

// TestIntegration_ABAC_InactivePolicy_Ignored verifies that a toggled-off (inactive)
// DENY policy is skipped during evaluation, allowing access to proceed.
func TestIntegration_ABAC_InactivePolicy_Ignored(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	seedRoleWithScope(t, bc, "staff", "data.read", "/api/v1/data", "GET")

	roles, _ := bc.ListRoles.Handle(ctx, query.ListRolesQuery{Pagination: shared.Pagination{Limit: 10}})
	perms, _ := bc.ListPermissions.Handle(ctx, query.ListPermissionsQuery{Pagination: shared.Pagination{Limit: 10}})
	roleID := roles.Roles[0].ID
	permID := perms.Permissions[0].ID

	// Create DENY policy that would block access.
	if err := bc.CreatePolicy.Handle(ctx, command.CreatePolicyCommand{
		PermissionID: permID,
		Effect:       "DENY",
		Priority:     10,
		Conditions:   map[string]any{"env.ip_in": []any{"192.168.1.100"}},
	}); err != nil {
		t.Fatalf("CreatePolicy: %v", err)
	}

	// Toggle the policy off.
	policies, err := bc.ListPolicies.Handle(ctx, query.ListPoliciesQuery{Pagination: shared.Pagination{Limit: 10}})
	if err != nil {
		t.Fatalf("ListPolicies: %v", err)
	}
	policyID := policies.Policies[0].ID

	if err := bc.TogglePolicy.Handle(ctx, command.TogglePolicyCommand{ID: policyID}); err != nil {
		t.Fatalf("TogglePolicy: %v", err)
	}

	// Access with matching IP should be allowed because the policy is inactive.
	allowed, err := bc.CheckAccess.Handle(ctx, query.CheckAccessQuery{
		RoleID: roleID,
		Path:   "/api/v1/data",
		Method: "GET",
		EvalCtx: domain.EvaluationContext{Attrs: map[string]map[string]any{
			"env": {"ip": "192.168.1.100"},
		}},
	})
	if err != nil {
		t.Fatalf("CheckAccess: %v", err)
	}
	if !allowed {
		t.Error("expected access allowed when DENY policy is inactive")
	}
}
