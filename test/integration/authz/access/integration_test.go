package access

import (
	"context"
	"testing"

	"gct/internal/context/iam/authz"
	"gct/internal/context/iam/authz/application/command"
	"gct/internal/context/iam/authz/application/query"
	"gct/internal/context/iam/authz/domain"
	shared "gct/internal/platform/domain"
	"gct/internal/platform/infrastructure/eventbus"
	"gct/internal/platform/infrastructure/logger"
	"gct/test/integration/common/setup"
)

func newTestBC(t *testing.T) *authz.BoundedContext {
	t.Helper()
	eb := eventbus.NewInMemoryEventBus()
	l := logger.New("error")
	return authz.NewBoundedContext(setup.TestPG.Pool, eb, l)
}

// seedRoleWithScope creates a role, permission, scope, and wires them together.
// Returns the role ID so callers can use it for CheckAccess.
func seedRoleWithScope(t *testing.T, bc *authz.BoundedContext, roleName, permName, scopePath, scopeMethod string) {
	t.Helper()
	ctx := context.Background()

	if err := bc.CreateRole.Handle(ctx, command.CreateRoleCommand{Name: roleName}); err != nil {
		t.Fatalf("CreateRole(%s): %v", roleName, err)
	}
	if err := bc.CreatePermission.Handle(ctx, command.CreatePermissionCommand{Name: permName}); err != nil {
		t.Fatalf("CreatePermission(%s): %v", permName, err)
	}
	if err := bc.CreateScope.Handle(ctx, command.CreateScopeCommand{Path: scopePath, Method: scopeMethod}); err != nil {
		t.Fatalf("CreateScope(%s %s): %v", scopeMethod, scopePath, err)
	}

	roles, _ := bc.ListRoles.Handle(ctx, query.ListRolesQuery{Pagination: shared.Pagination{Limit: 100}})
	perms, _ := bc.ListPermissions.Handle(ctx, query.ListPermissionsQuery{Pagination: shared.Pagination{Limit: 100}})

	var roleID, permID = roles.Roles[len(roles.Roles)-1].ID, perms.Permissions[len(perms.Permissions)-1].ID

	if err := bc.AssignPermission.Handle(ctx, command.AssignPermissionCommand{RoleID: roleID, PermissionID: permID}); err != nil {
		t.Fatalf("AssignPermission: %v", err)
	}
	if err := bc.AssignScope.Handle(ctx, command.AssignScopeCommand{PermissionID: permID, Path: scopePath, Method: scopeMethod}); err != nil {
		t.Fatalf("AssignScope: %v", err)
	}
}

// ---------------------------------------------------------------------------
// CheckAccess — exact match
// ---------------------------------------------------------------------------

func TestIntegration_CheckAccess_ExactMatch_Allowed(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	seedRoleWithScope(t, bc, "editor", "articles.read", "/api/v1/articles", "GET")

	roles, _ := bc.ListRoles.Handle(ctx, query.ListRolesQuery{Pagination: shared.Pagination{Limit: 10}})
	roleID := roles.Roles[0].ID

	allowed, err := bc.CheckAccess.Handle(ctx, query.CheckAccessQuery{
		RoleID:  roleID,
		Path:    "/api/v1/articles",
		Method:  "GET",
		EvalCtx: domain.EvaluationContext{Attrs: map[string]map[string]any{}},
	})
	if err != nil {
		t.Fatalf("CheckAccess: %v", err)
	}
	if !allowed {
		t.Error("expected access to be allowed for exact match")
	}
}

func TestIntegration_CheckAccess_ExactMatch_DeniedWrongPath(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	seedRoleWithScope(t, bc, "editor", "articles.read", "/api/v1/articles", "GET")

	roles, _ := bc.ListRoles.Handle(ctx, query.ListRolesQuery{Pagination: shared.Pagination{Limit: 10}})
	roleID := roles.Roles[0].ID

	allowed, err := bc.CheckAccess.Handle(ctx, query.CheckAccessQuery{
		RoleID:  roleID,
		Path:    "/api/v1/users",
		Method:  "GET",
		EvalCtx: domain.EvaluationContext{Attrs: map[string]map[string]any{}},
	})
	if err != nil {
		t.Fatalf("CheckAccess: %v", err)
	}
	if allowed {
		t.Error("expected access to be denied for wrong path")
	}
}

func TestIntegration_CheckAccess_ExactMatch_DeniedWrongMethod(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	seedRoleWithScope(t, bc, "editor", "articles.read", "/api/v1/articles", "GET")

	roles, _ := bc.ListRoles.Handle(ctx, query.ListRolesQuery{Pagination: shared.Pagination{Limit: 10}})
	roleID := roles.Roles[0].ID

	allowed, err := bc.CheckAccess.Handle(ctx, query.CheckAccessQuery{
		RoleID:  roleID,
		Path:    "/api/v1/articles",
		Method:  "POST",
		EvalCtx: domain.EvaluationContext{Attrs: map[string]map[string]any{}},
	})
	if err != nil {
		t.Fatalf("CheckAccess: %v", err)
	}
	if allowed {
		t.Error("expected access to be denied for wrong method")
	}
}

// ---------------------------------------------------------------------------
// CheckAccess — super_admin bypass
// ---------------------------------------------------------------------------

func TestIntegration_CheckAccess_SuperAdminBypass(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	// Create super_admin with NO scopes — should still have full access.
	if err := bc.CreateRole.Handle(ctx, command.CreateRoleCommand{Name: "super_admin"}); err != nil {
		t.Fatalf("CreateRole: %v", err)
	}

	roles, _ := bc.ListRoles.Handle(ctx, query.ListRolesQuery{Pagination: shared.Pagination{Limit: 10}})
	roleID := roles.Roles[0].ID

	allowed, err := bc.CheckAccess.Handle(ctx, query.CheckAccessQuery{
		RoleID:  roleID,
		Path:    "/api/v1/anything/at/all",
		Method:  "DELETE",
		EvalCtx: domain.EvaluationContext{Attrs: map[string]map[string]any{}},
	})
	if err != nil {
		t.Fatalf("CheckAccess: %v", err)
	}
	if !allowed {
		t.Error("expected super_admin to bypass all access checks")
	}
}

// ---------------------------------------------------------------------------
// CheckAccess — wildcard method
// ---------------------------------------------------------------------------

func TestIntegration_CheckAccess_WildcardMethod(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	seedRoleWithScope(t, bc, "admin", "users.all", "/api/v1/users", "*")

	roles, _ := bc.ListRoles.Handle(ctx, query.ListRolesQuery{Pagination: shared.Pagination{Limit: 10}})
	roleID := roles.Roles[0].ID

	for _, method := range []string{"GET", "POST", "PUT", "PATCH", "DELETE"} {
		t.Run(method, func(t *testing.T) {
			allowed, err := bc.CheckAccess.Handle(ctx, query.CheckAccessQuery{
				RoleID:  roleID,
				Path:    "/api/v1/users",
				Method:  method,
				EvalCtx: domain.EvaluationContext{Attrs: map[string]map[string]any{}},
			})
			if err != nil {
				t.Fatalf("CheckAccess: %v", err)
			}
			if !allowed {
				t.Errorf("expected wildcard method scope to allow %s", method)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// CheckAccess — prefix wildcard path
// ---------------------------------------------------------------------------

func TestIntegration_CheckAccess_PrefixWildcardPath(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	seedRoleWithScope(t, bc, "manager", "users.manage", "/api/v1/users*", "GET")

	roles, _ := bc.ListRoles.Handle(ctx, query.ListRolesQuery{Pagination: shared.Pagination{Limit: 10}})
	roleID := roles.Roles[0].ID

	tests := []struct {
		path    string
		allowed bool
	}{
		{"/api/v1/users", true},
		{"/api/v1/users/123", true},
		{"/api/v1/users/123/sessions", true},
		{"/api/v1/roles", false},
		{"/api/v1/articles", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			allowed, err := bc.CheckAccess.Handle(ctx, query.CheckAccessQuery{
				RoleID:  roleID,
				Path:    tt.path,
				Method:  "GET",
				EvalCtx: domain.EvaluationContext{Attrs: map[string]map[string]any{}},
			})
			if err != nil {
				t.Fatalf("CheckAccess: %v", err)
			}
			if allowed != tt.allowed {
				t.Errorf("path %s: expected allowed=%v, got %v", tt.path, tt.allowed, allowed)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// CheckAccess — role with no permissions
// ---------------------------------------------------------------------------

func TestIntegration_CheckAccess_RoleNoPermissions(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	if err := bc.CreateRole.Handle(ctx, command.CreateRoleCommand{Name: "empty_role"}); err != nil {
		t.Fatalf("CreateRole: %v", err)
	}

	roles, _ := bc.ListRoles.Handle(ctx, query.ListRolesQuery{Pagination: shared.Pagination{Limit: 10}})
	roleID := roles.Roles[0].ID

	allowed, err := bc.CheckAccess.Handle(ctx, query.CheckAccessQuery{
		RoleID:  roleID,
		Path:    "/api/v1/users",
		Method:  "GET",
		EvalCtx: domain.EvaluationContext{Attrs: map[string]map[string]any{}},
	})
	if err != nil {
		t.Fatalf("CheckAccess: %v", err)
	}
	if allowed {
		t.Error("expected access denied for role with no permissions")
	}
}

// ---------------------------------------------------------------------------
// CheckAccess — multiple scopes on one role
// ---------------------------------------------------------------------------

func TestIntegration_CheckAccess_MultipleScopes(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	// Create role with two different permission+scope combos.
	if err := bc.CreateRole.Handle(ctx, command.CreateRoleCommand{Name: "multi"}); err != nil {
		t.Fatalf("CreateRole: %v", err)
	}
	if err := bc.CreatePermission.Handle(ctx, command.CreatePermissionCommand{Name: "users.read"}); err != nil {
		t.Fatalf("CreatePermission: %v", err)
	}
	if err := bc.CreatePermission.Handle(ctx, command.CreatePermissionCommand{Name: "articles.read"}); err != nil {
		t.Fatalf("CreatePermission: %v", err)
	}
	if err := bc.CreateScope.Handle(ctx, command.CreateScopeCommand{Path: "/api/v1/users", Method: "GET"}); err != nil {
		t.Fatalf("CreateScope: %v", err)
	}
	if err := bc.CreateScope.Handle(ctx, command.CreateScopeCommand{Path: "/api/v1/articles", Method: "GET"}); err != nil {
		t.Fatalf("CreateScope: %v", err)
	}

	roles, _ := bc.ListRoles.Handle(ctx, query.ListRolesQuery{Pagination: shared.Pagination{Limit: 10}})
	perms, _ := bc.ListPermissions.Handle(ctx, query.ListPermissionsQuery{Pagination: shared.Pagination{Limit: 100}})

	roleID := roles.Roles[0].ID

	for _, p := range perms.Permissions {
		if err := bc.AssignPermission.Handle(ctx, command.AssignPermissionCommand{RoleID: roleID, PermissionID: p.ID}); err != nil {
			t.Fatalf("AssignPermission(%s): %v", p.Name, err)
		}
		var scopePath string
		if p.Name == "users.read" {
			scopePath = "/api/v1/users"
		} else {
			scopePath = "/api/v1/articles"
		}
		if err := bc.AssignScope.Handle(ctx, command.AssignScopeCommand{PermissionID: p.ID, Path: scopePath, Method: "GET"}); err != nil {
			t.Fatalf("AssignScope(%s): %v", p.Name, err)
		}
	}

	tests := []struct {
		path    string
		method  string
		allowed bool
	}{
		{"/api/v1/users", "GET", true},
		{"/api/v1/articles", "GET", true},
		{"/api/v1/users", "POST", false},
		{"/api/v1/roles", "GET", false},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			allowed, err := bc.CheckAccess.Handle(ctx, query.CheckAccessQuery{
				RoleID:  roleID,
				Path:    tt.path,
				Method:  tt.method,
				EvalCtx: domain.EvaluationContext{Attrs: map[string]map[string]any{}},
			})
			if err != nil {
				t.Fatalf("CheckAccess: %v", err)
			}
			if allowed != tt.allowed {
				t.Errorf("expected allowed=%v, got %v", tt.allowed, allowed)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// CheckAccess — nonexistent role returns error
// ---------------------------------------------------------------------------

func TestIntegration_CheckAccess_NonexistentRole(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	_, err := bc.CheckAccess.Handle(ctx, query.CheckAccessQuery{
		RoleID:  [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		Path:    "/api/v1/users",
		Method:  "GET",
		EvalCtx: domain.EvaluationContext{Attrs: map[string]map[string]any{}},
	})
	if err == nil {
		t.Error("expected error for nonexistent role, got nil")
	}
}

// ---------------------------------------------------------------------------
// Full lifecycle: create → assign → check → revoke → re-check
// ---------------------------------------------------------------------------

func TestIntegration_CheckAccess_FullLifecycle(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	// 1. Create role + permission + scope.
	seedRoleWithScope(t, bc, "lifecycle", "users.view", "/api/v1/users", "GET")

	roles, _ := bc.ListRoles.Handle(ctx, query.ListRolesQuery{Pagination: shared.Pagination{Limit: 10}})
	roleID := roles.Roles[0].ID

	// 2. Access should be granted.
	allowed, err := bc.CheckAccess.Handle(ctx, query.CheckAccessQuery{
		RoleID: roleID, Path: "/api/v1/users", Method: "GET",
		EvalCtx: domain.EvaluationContext{Attrs: map[string]map[string]any{}},
	})
	if err != nil {
		t.Fatalf("CheckAccess: %v", err)
	}
	if !allowed {
		t.Fatal("expected access allowed after assignment")
	}

	// 3. Delete the role entirely.
	if err := bc.DeleteRole.Handle(ctx, command.DeleteRoleCommand{ID: roleID}); err != nil {
		t.Fatalf("DeleteRole: %v", err)
	}

	// 4. Access check should fail (role gone).
	_, err = bc.CheckAccess.Handle(ctx, query.CheckAccessQuery{
		RoleID: roleID, Path: "/api/v1/users", Method: "GET",
		EvalCtx: domain.EvaluationContext{Attrs: map[string]map[string]any{}},
	})
	if err == nil {
		t.Error("expected error after role deletion, got nil")
	}
}

// ---------------------------------------------------------------------------
// Scope + Permission assignment — adding new scope to existing role
// ---------------------------------------------------------------------------

func TestIntegration_CheckAccess_AddScopeToExistingPermission(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	seedRoleWithScope(t, bc, "evolving", "resources.manage", "/api/v1/resources", "GET")

	roles, _ := bc.ListRoles.Handle(ctx, query.ListRolesQuery{Pagination: shared.Pagination{Limit: 10}})
	perms, _ := bc.ListPermissions.Handle(ctx, query.ListPermissionsQuery{Pagination: shared.Pagination{Limit: 10}})
	roleID := roles.Roles[0].ID
	permID := perms.Permissions[0].ID

	// Initially POST is not allowed.
	allowed, _ := bc.CheckAccess.Handle(ctx, query.CheckAccessQuery{
		RoleID: roleID, Path: "/api/v1/resources", Method: "POST",
		EvalCtx: domain.EvaluationContext{Attrs: map[string]map[string]any{}},
	})
	if allowed {
		t.Fatal("POST should not be allowed yet")
	}

	// Add POST scope.
	if err := bc.CreateScope.Handle(ctx, command.CreateScopeCommand{Path: "/api/v1/resources", Method: "POST"}); err != nil {
		t.Fatalf("CreateScope: %v", err)
	}
	if err := bc.AssignScope.Handle(ctx, command.AssignScopeCommand{PermissionID: permID, Path: "/api/v1/resources", Method: "POST"}); err != nil {
		t.Fatalf("AssignScope: %v", err)
	}

	// Now POST should be allowed.
	allowed, err := bc.CheckAccess.Handle(ctx, query.CheckAccessQuery{
		RoleID: roleID, Path: "/api/v1/resources", Method: "POST",
		EvalCtx: domain.EvaluationContext{Attrs: map[string]map[string]any{}},
	})
	if err != nil {
		t.Fatalf("CheckAccess: %v", err)
	}
	if !allowed {
		t.Error("expected POST to be allowed after adding scope")
	}

	// GET should still be allowed.
	allowed, _ = bc.CheckAccess.Handle(ctx, query.CheckAccessQuery{
		RoleID: roleID, Path: "/api/v1/resources", Method: "GET",
		EvalCtx: domain.EvaluationContext{Attrs: map[string]map[string]any{}},
	})
	if !allowed {
		t.Error("GET should still be allowed")
	}
}
