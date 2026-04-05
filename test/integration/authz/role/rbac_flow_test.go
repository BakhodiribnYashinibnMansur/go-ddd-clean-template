package role

import (
	"context"
	"testing"

	"gct/internal/context/iam/authz/application/command"
	"gct/internal/context/iam/authz/application/query"
	"gct/internal/context/iam/authz/domain"
	shared "gct/internal/kernel/domain"
	"gct/test/integration/common/setup"
)

// TestIntegration_RBACFullFlow exercises the complete RBAC lifecycle:
//  1. Create a Role
//  2. Create a Permission
//  3. Create a Scope and assign it to the Permission
//  4. Assign the Permission to the Role
//  5. Verify the assignment exists in the role_permission table
//  6. Revoke the Permission from the Role (via direct SQL, no command handler exists)
//  7. Verify the association is removed
func TestIntegration_RBACFullFlow(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	// --- Step 1: Create a Role ---
	roleDesc := "Full-access administrator"
	err := bc.CreateRole.Handle(ctx, command.CreateRoleCommand{
		Name:        "admin",
		Description: &roleDesc,
	})
	if err != nil {
		t.Fatalf("CreateRole: %v", err)
	}

	roles, err := bc.ListRoles.Handle(ctx, query.ListRolesQuery{
		Pagination: shared.Pagination{Limit: 10},
	})
	if err != nil {
		t.Fatalf("ListRoles: %v", err)
	}
	if roles.Total != 1 {
		t.Fatalf("expected 1 role, got %d", roles.Total)
	}
	roleID := domain.RoleID(roles.Roles[0].ID)

	// --- Step 2: Create a Permission ---
	err = bc.CreatePermission.Handle(ctx, command.CreatePermissionCommand{
		Name: "users.manage",
	})
	if err != nil {
		t.Fatalf("CreatePermission: %v", err)
	}

	perms, err := bc.ListPermissions.Handle(ctx, query.ListPermissionsQuery{
		Pagination: shared.Pagination{Limit: 10},
	})
	if err != nil {
		t.Fatalf("ListPermissions: %v", err)
	}
	if perms.Total != 1 {
		t.Fatalf("expected 1 permission, got %d", perms.Total)
	}
	permID := domain.PermissionID(perms.Permissions[0].ID)

	// --- Step 3: Create a Scope and assign it to the Permission ---
	err = bc.CreateScope.Handle(ctx, command.CreateScopeCommand{
		Path:   "/api/v1/users",
		Method: "GET",
	})
	if err != nil {
		t.Fatalf("CreateScope: %v", err)
	}

	err = bc.AssignScope.Handle(ctx, command.AssignScopeCommand{
		PermissionID: permID,
		Path:         "/api/v1/users",
		Method:       "GET",
	})
	if err != nil {
		t.Fatalf("AssignScope: %v", err)
	}

	// Verify scope was created.
	scopes, err := bc.ListScopes.Handle(ctx, query.ListScopesQuery{
		Pagination: shared.Pagination{Limit: 10},
	})
	if err != nil {
		t.Fatalf("ListScopes: %v", err)
	}
	if scopes.Total != 1 {
		t.Fatalf("expected 1 scope, got %d", scopes.Total)
	}
	if scopes.Scopes[0].Path != "/api/v1/users" {
		t.Errorf("expected scope path /api/v1/users, got %s", scopes.Scopes[0].Path)
	}
	if scopes.Scopes[0].Method != "GET" {
		t.Errorf("expected scope method GET, got %s", scopes.Scopes[0].Method)
	}

	// --- Step 4: Assign Permission to Role ---
	err = bc.AssignPermission.Handle(ctx, command.AssignPermissionCommand{
		RoleID:       roleID,
		PermissionID: permID,
	})
	if err != nil {
		t.Fatalf("AssignPermission: %v", err)
	}

	// --- Step 5: Verify the association exists ---
	var count int
	err = setup.TestPG.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM role_permission WHERE role_id = $1 AND permission_id = $2`,
		roleID, permID,
	).Scan(&count)
	if err != nil {
		t.Fatalf("query role_permission: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 role_permission row, got %d", count)
	}

	// --- Step 6: Revoke the Permission from the Role ---
	// No RevokePermission command handler is exposed on the BoundedContext,
	// so we revoke via direct SQL against the join table.
	_, err = setup.TestPG.Pool.Exec(ctx,
		`DELETE FROM role_permission WHERE role_id = $1 AND permission_id = $2`,
		roleID, permID,
	)
	if err != nil {
		t.Fatalf("revoke role_permission: %v", err)
	}

	// --- Step 7: Verify the association is removed ---
	err = setup.TestPG.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM role_permission WHERE role_id = $1 AND permission_id = $2`,
		roleID, permID,
	).Scan(&count)
	if err != nil {
		t.Fatalf("query role_permission after revoke: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 role_permission rows after revoke, got %d", count)
	}
}

// TestIntegration_RBACMultiplePermissionsOnRole verifies that a role can hold
// multiple permissions and that revoking one does not affect the others.
func TestIntegration_RBACMultiplePermissionsOnRole(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	// Create a role.
	err := bc.CreateRole.Handle(ctx, command.CreateRoleCommand{Name: "editor"})
	if err != nil {
		t.Fatalf("CreateRole: %v", err)
	}
	roles, _ := bc.ListRoles.Handle(ctx, query.ListRolesQuery{
		Pagination: shared.Pagination{Limit: 10},
	})
	roleID := domain.RoleID(roles.Roles[0].ID)

	// Create two permissions.
	for _, name := range []string{"articles.read", "articles.write"} {
		err = bc.CreatePermission.Handle(ctx, command.CreatePermissionCommand{Name: name})
		if err != nil {
			t.Fatalf("CreatePermission(%s): %v", name, err)
		}
	}
	perms, _ := bc.ListPermissions.Handle(ctx, query.ListPermissionsQuery{
		Pagination: shared.Pagination{Limit: 10},
	})
	if perms.Total != 2 {
		t.Fatalf("expected 2 permissions, got %d", perms.Total)
	}

	// Assign both permissions to the role.
	for _, p := range perms.Permissions {
		err = bc.AssignPermission.Handle(ctx, command.AssignPermissionCommand{
			RoleID:       roleID,
			PermissionID: domain.PermissionID(p.ID),
		})
		if err != nil {
			t.Fatalf("AssignPermission(%s): %v", p.Name, err)
		}
	}

	// Verify both are assigned.
	var count int
	err = setup.TestPG.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM role_permission WHERE role_id = $1`, roleID,
	).Scan(&count)
	if err != nil {
		t.Fatalf("query role_permission count: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 role_permission rows, got %d", count)
	}

	// Revoke one permission.
	_, err = setup.TestPG.Pool.Exec(ctx,
		`DELETE FROM role_permission WHERE role_id = $1 AND permission_id = $2`,
		roleID, perms.Permissions[0].ID,
	)
	if err != nil {
		t.Fatalf("revoke one permission: %v", err)
	}

	// Verify only one remains.
	err = setup.TestPG.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM role_permission WHERE role_id = $1`, roleID,
	).Scan(&count)
	if err != nil {
		t.Fatalf("query role_permission after partial revoke: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 role_permission row after revoking one, got %d", count)
	}
}
