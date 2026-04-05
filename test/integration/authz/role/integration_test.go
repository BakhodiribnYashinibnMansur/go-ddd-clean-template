package role

import (
	"context"
	"testing"

	"gct/internal/context/iam/authz"
	"gct/internal/context/iam/authz/application/command"
	"gct/internal/context/iam/authz/application/query"
	shared "gct/internal/kernel/domain"
	"gct/internal/kernel/infrastructure/eventbus"
	"gct/internal/kernel/infrastructure/logger"
	"gct/test/integration/common/setup"
)

func newTestBC(t *testing.T) *authz.BoundedContext {
	t.Helper()
	eb := eventbus.NewInMemoryEventBus()
	l := logger.New("error")
	return authz.NewBoundedContext(setup.TestPG.Pool, eb, l)
}

func TestIntegration_CreateAndGetRole(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	desc := "Admin role"
	err := bc.CreateRole.Handle(ctx, command.CreateRoleCommand{
		Name:        "admin",
		Description: &desc,
	})
	if err != nil {
		t.Fatalf("CreateRole: %v", err)
	}

	result, err := bc.ListRoles.Handle(ctx, query.ListRolesQuery{
		Pagination: shared.Pagination{Limit: 10},
	})
	if err != nil {
		t.Fatalf("ListRoles: %v", err)
	}
	if result.Total != 1 {
		t.Fatalf("expected 1 role, got %d", result.Total)
	}

	role := result.Roles[0]
	if role.Name != "admin" {
		t.Errorf("expected name admin, got %s", role.Name)
	}
	if role.Description == nil || *role.Description != "Admin role" {
		t.Error("expected description 'Admin role'")
	}

	view, err := bc.GetRole.Handle(ctx, query.GetRoleQuery{ID: role.ID})
	if err != nil {
		t.Fatalf("GetRole: %v", err)
	}
	if view.ID != role.ID {
		t.Errorf("ID mismatch: %s vs %s", view.ID, role.ID)
	}
}

func TestIntegration_DeleteRole(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	err := bc.CreateRole.Handle(ctx, command.CreateRoleCommand{Name: "to-delete"})
	if err != nil {
		t.Fatalf("CreateRole: %v", err)
	}

	list, _ := bc.ListRoles.Handle(ctx, query.ListRolesQuery{
		Pagination: shared.Pagination{Limit: 10},
	})
	roleID := list.Roles[0].ID

	err = bc.DeleteRole.Handle(ctx, command.DeleteRoleCommand{ID: roleID})
	if err != nil {
		t.Fatalf("DeleteRole: %v", err)
	}

	list2, _ := bc.ListRoles.Handle(ctx, query.ListRolesQuery{
		Pagination: shared.Pagination{Limit: 10},
	})
	if list2.Total != 0 {
		t.Errorf("expected 0 roles after delete, got %d", list2.Total)
	}
}

func TestIntegration_AssignPermissionToRole(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	err := bc.CreateRole.Handle(ctx, command.CreateRoleCommand{Name: "editor"})
	if err != nil {
		t.Fatalf("CreateRole: %v", err)
	}

	err = bc.CreatePermission.Handle(ctx, command.CreatePermissionCommand{Name: "articles.write"})
	if err != nil {
		t.Fatalf("CreatePermission: %v", err)
	}

	roles, _ := bc.ListRoles.Handle(ctx, query.ListRolesQuery{
		Pagination: shared.Pagination{Limit: 10},
	})
	perms, _ := bc.ListPermissions.Handle(ctx, query.ListPermissionsQuery{
		Pagination: shared.Pagination{Limit: 10},
	})

	err = bc.AssignPermission.Handle(ctx, command.AssignPermissionCommand{
		RoleID:       roles.Roles[0].ID,
		PermissionID: perms.Permissions[0].ID,
	})
	if err != nil {
		t.Fatalf("AssignPermission: %v", err)
	}
}
