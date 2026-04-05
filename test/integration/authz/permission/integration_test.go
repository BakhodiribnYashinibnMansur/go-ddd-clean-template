package permission

import (
	"context"
	"testing"

	"gct/internal/context/iam/authz"
	"gct/internal/context/iam/authz/application/command"
	"gct/internal/context/iam/authz/application/query"
	"gct/internal/context/iam/authz/domain"
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

func TestIntegration_CreateAndListPermissions(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	desc := "Read articles"
	err := bc.CreatePermission.Handle(ctx, command.CreatePermissionCommand{
		Name:        "articles.read",
		Description: &desc,
	})
	if err != nil {
		t.Fatalf("CreatePermission: %v", err)
	}

	result, err := bc.ListPermissions.Handle(ctx, query.ListPermissionsQuery{
		Pagination: shared.Pagination{Limit: 10},
	})
	if err != nil {
		t.Fatalf("ListPermissions: %v", err)
	}
	if result.Total != 1 {
		t.Fatalf("expected 1 permission, got %d", result.Total)
	}

	perm := result.Permissions[0]
	if perm.Name != "articles.read" {
		t.Errorf("expected name articles.read, got %s", perm.Name)
	}
}

func TestIntegration_HierarchicalPermission(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	err := bc.CreatePermission.Handle(ctx, command.CreatePermissionCommand{Name: "articles"})
	if err != nil {
		t.Fatalf("CreatePermission parent: %v", err)
	}

	list, _ := bc.ListPermissions.Handle(ctx, query.ListPermissionsQuery{
		Pagination: shared.Pagination{Limit: 10},
	})
	parentID := domain.PermissionID(list.Permissions[0].ID)

	err = bc.CreatePermission.Handle(ctx, command.CreatePermissionCommand{
		Name:     "articles.write",
		ParentID: &parentID,
	})
	if err != nil {
		t.Fatalf("CreatePermission child: %v", err)
	}

	list2, _ := bc.ListPermissions.Handle(ctx, query.ListPermissionsQuery{
		Pagination: shared.Pagination{Limit: 10},
	})
	if list2.Total != 2 {
		t.Fatalf("expected 2 permissions, got %d", list2.Total)
	}
}

func TestIntegration_DeletePermission(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	err := bc.CreatePermission.Handle(ctx, command.CreatePermissionCommand{Name: "to-delete"})
	if err != nil {
		t.Fatalf("CreatePermission: %v", err)
	}

	list, _ := bc.ListPermissions.Handle(ctx, query.ListPermissionsQuery{
		Pagination: shared.Pagination{Limit: 10},
	})
	permID := domain.PermissionID(list.Permissions[0].ID)

	err = bc.DeletePermission.Handle(ctx, command.DeletePermissionCommand{ID: permID})
	if err != nil {
		t.Fatalf("DeletePermission: %v", err)
	}

	list2, _ := bc.ListPermissions.Handle(ctx, query.ListPermissionsQuery{
		Pagination: shared.Pagination{Limit: 10},
	})
	if list2.Total != 0 {
		t.Errorf("expected 0 permissions after delete, got %d", list2.Total)
	}
}
