package access

import (
	"context"
	"testing"

	"gct/internal/context/iam/generic/authz/application/command"
	"gct/internal/context/iam/generic/authz/application/query"
	"gct/internal/context/iam/generic/authz/domain"
	shared "gct/internal/kernel/domain"
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
	roleID := domain.RoleID(roles.Roles[0].ID)

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
	roleID := domain.RoleID(roles.Roles[0].ID)
	permID := domain.PermissionID(perms.Permissions[0].ID)

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
	roleID := domain.RoleID(roles.Roles[0].ID)
	permID := domain.PermissionID(perms.Permissions[0].ID)

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
	roleID := domain.RoleID(roles.Roles[0].ID)
	permID := domain.PermissionID(perms.Permissions[0].ID)

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
	roleID := domain.RoleID(roles.Roles[0].ID)

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
	roleID := domain.RoleID(roles.Roles[0].ID)
	permID := domain.PermissionID(perms.Permissions[0].ID)

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
	policyID := domain.PolicyID(policies.Policies[0].ID)

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

// ---------------------------------------------------------------------------
// Time-based ABAC policies (between operator)
// ---------------------------------------------------------------------------

// TestIntegration_ABAC_TimeBetween_Allowed verifies that access is granted
// when the current time falls within the allowed window (e.g., 02:00–17:00).
func TestIntegration_ABAC_TimeBetween_Allowed(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	seedRoleWithScope(t, bc, "worker", "tasks.manage", "/api/v1/tasks", "POST")

	roles, _ := bc.ListRoles.Handle(ctx, query.ListRolesQuery{Pagination: shared.Pagination{Limit: 10}})
	perms, _ := bc.ListPermissions.Handle(ctx, query.ListPermissionsQuery{Pagination: shared.Pagination{Limit: 10}})
	roleID := domain.RoleID(roles.Roles[0].ID)
	permID := domain.PermissionID(perms.Permissions[0].ID)

	// ALLOW only between 02:00 and 17:00.
	if err := bc.CreatePolicy.Handle(ctx, command.CreatePolicyCommand{
		PermissionID: permID,
		Effect:       "ALLOW",
		Priority:     10,
		Conditions:   map[string]any{"env.time_between": []any{"02:00", "17:00"}},
	}); err != nil {
		t.Fatalf("CreatePolicy: %v", err)
	}

	// Request at 10:30 — within range.
	allowed, err := bc.CheckAccess.Handle(ctx, query.CheckAccessQuery{
		RoleID: roleID,
		Path:   "/api/v1/tasks",
		Method: "POST",
		EvalCtx: domain.EvaluationContext{Attrs: map[string]map[string]any{
			"user": {},
			"env":  {"time": "10:30"},
		}},
	})
	if err != nil {
		t.Fatalf("CheckAccess: %v", err)
	}
	if !allowed {
		t.Error("expected allowed — 10:30 is within 02:00–17:00")
	}
}

// TestIntegration_ABAC_TimeBetween_DenyOutsideHours verifies that a DENY policy
// blocks access during off-hours (e.g., deny between 17:00 and 23:59).
func TestIntegration_ABAC_TimeBetween_DenyOutsideHours(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	seedRoleWithScope(t, bc, "worker", "tasks.manage", "/api/v1/tasks", "POST")

	roles, _ := bc.ListRoles.Handle(ctx, query.ListRolesQuery{Pagination: shared.Pagination{Limit: 10}})
	perms, _ := bc.ListPermissions.Handle(ctx, query.ListPermissionsQuery{Pagination: shared.Pagination{Limit: 10}})
	roleID := domain.RoleID(roles.Roles[0].ID)
	permID := domain.PermissionID(perms.Permissions[0].ID)

	// DENY between 17:00 and 23:59 (off-hours).
	if err := bc.CreatePolicy.Handle(ctx, command.CreatePolicyCommand{
		PermissionID: permID,
		Effect:       "DENY",
		Priority:     10,
		Conditions:   map[string]any{"env.time_between": []any{"17:00", "23:59"}},
	}); err != nil {
		t.Fatalf("CreatePolicy: %v", err)
	}

	// 23:00 is within off-hours → DENY matches → blocked.
	allowed, err := bc.CheckAccess.Handle(ctx, query.CheckAccessQuery{
		RoleID: roleID,
		Path:   "/api/v1/tasks",
		Method: "POST",
		EvalCtx: domain.EvaluationContext{Attrs: map[string]map[string]any{
			"user": {},
			"env":  {"time": "23:00"},
		}},
	})
	if err != nil {
		t.Fatalf("CheckAccess: %v", err)
	}
	if allowed {
		t.Error("expected denied — 23:00 is within off-hours 17:00–23:59")
	}

	// 10:30 is outside off-hours → DENY doesn't match → RBAC stands → allowed.
	allowed, err = bc.CheckAccess.Handle(ctx, query.CheckAccessQuery{
		RoleID: roleID,
		Path:   "/api/v1/tasks",
		Method: "POST",
		EvalCtx: domain.EvaluationContext{Attrs: map[string]map[string]any{
			"user": {},
			"env":  {"time": "10:30"},
		}},
	})
	if err != nil {
		t.Fatalf("CheckAccess: %v", err)
	}
	if !allowed {
		t.Error("expected allowed — 10:30 is outside off-hours")
	}
}

// TestIntegration_ABAC_TimeBetween_WorkingHours verifies the 02:00–05:00 working
// window pattern using a DENY policy for outside-hours.
func TestIntegration_ABAC_TimeBetween_WorkingHours(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	seedRoleWithScope(t, bc, "guard", "gates.open", "/api/v1/gates", "POST")

	roles, _ := bc.ListRoles.Handle(ctx, query.ListRolesQuery{Pagination: shared.Pagination{Limit: 10}})
	perms, _ := bc.ListPermissions.Handle(ctx, query.ListPermissionsQuery{Pagination: shared.Pagination{Limit: 10}})
	roleID := domain.RoleID(roles.Roles[0].ID)
	permID := domain.PermissionID(perms.Permissions[0].ID)

	// Pattern: "only allow 02:00–05:00" = two DENY policies covering the outside.
	// DENY 00:00–01:59 (before window).
	if err := bc.CreatePolicy.Handle(ctx, command.CreatePolicyCommand{
		PermissionID: permID,
		Effect:       "DENY",
		Priority:     10,
		Conditions:   map[string]any{"env.time_between": []any{"00:00", "01:59"}},
	}); err != nil {
		t.Fatalf("CreatePolicy: %v", err)
	}
	// DENY 05:01–23:59 (after window).
	if err := bc.CreatePolicy.Handle(ctx, command.CreatePolicyCommand{
		PermissionID: permID,
		Effect:       "DENY",
		Priority:     10,
		Conditions:   map[string]any{"env.time_between": []any{"05:01", "23:59"}},
	}); err != nil {
		t.Fatalf("CreatePolicy: %v", err)
	}

	tests := []struct {
		time    string
		allowed bool
	}{
		{"02:00", true},  // lower bound — inside window
		{"03:30", true},  // middle of window
		{"05:00", true},  // upper bound — inside window
		{"01:59", false}, // just before window
		{"05:01", false}, // just after window
		{"12:00", false}, // afternoon — outside
		{"00:00", false}, // midnight — outside
	}

	for _, tt := range tests {
		t.Run(tt.time, func(t *testing.T) {
			allowed, err := bc.CheckAccess.Handle(ctx, query.CheckAccessQuery{
				RoleID: roleID,
				Path:   "/api/v1/gates",
				Method: "POST",
				EvalCtx: domain.EvaluationContext{Attrs: map[string]map[string]any{
					"user": {},
					"env":  {"time": tt.time},
				}},
			})
			if err != nil {
				t.Fatalf("CheckAccess: %v", err)
			}
			if allowed != tt.allowed {
				t.Errorf("time %s: expected allowed=%v, got %v", tt.time, tt.allowed, allowed)
			}
		})
	}
}
