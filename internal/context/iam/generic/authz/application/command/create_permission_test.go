package command

import (
	"context"
	"testing"

	"gct/internal/context/iam/generic/authz/domain"
	shared "gct/internal/kernel/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// --- Mock PermissionRepository ---

type mockPermissionRepository struct {
	savedPerm   *domain.Permission
	updatedPerm *domain.Permission
	findByIDFn  func(ctx context.Context, id domain.PermissionID) (*domain.Permission, error)
	saveFn      func(ctx context.Context, perm *domain.Permission) error
	deleteFn    func(ctx context.Context, id domain.PermissionID) error
}

func (m *mockPermissionRepository) Save(ctx context.Context, perm *domain.Permission) error {
	if m.saveFn != nil {
		return m.saveFn(ctx, perm)
	}
	m.savedPerm = perm
	return nil
}

func (m *mockPermissionRepository) FindByID(ctx context.Context, id domain.PermissionID) (*domain.Permission, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, domain.ErrPermissionNotFound
}

func (m *mockPermissionRepository) Update(ctx context.Context, perm *domain.Permission) error {
	m.updatedPerm = perm
	return nil
}

func (m *mockPermissionRepository) Delete(ctx context.Context, id domain.PermissionID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func (m *mockPermissionRepository) List(ctx context.Context, pagination shared.Pagination) ([]*domain.Permission, int64, error) {
	return nil, 0, nil
}

// --- Tests ---

func TestCreatePermissionHandler_Success(t *testing.T) {
	t.Parallel()

	repo := &mockPermissionRepository{}
	log := &mockLogger{}

	handler := NewCreatePermissionHandler(repo, log)

	desc := "Read access"
	cmd := CreatePermissionCommand{
		Name:        "read",
		Description: &desc,
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	if repo.savedPerm == nil {
		t.Fatal("expected permission to be saved")
	}

	if repo.savedPerm.Name() != "read" {
		t.Errorf("expected name 'read', got '%s'", repo.savedPerm.Name())
	}

	if repo.savedPerm.Description() == nil || *repo.savedPerm.Description() != "Read access" {
		t.Error("expected description 'Read access'")
	}

	if repo.savedPerm.ParentID() != nil {
		t.Error("expected nil parent ID")
	}
}

func TestCreatePermissionHandler_WithParent(t *testing.T) {
	t.Parallel()

	repo := &mockPermissionRepository{}
	log := &mockLogger{}

	handler := NewCreatePermissionHandler(repo, log)

	parentUUID := uuid.New()
	parentID := domain.PermissionID(parentUUID)
	cmd := CreatePermissionCommand{
		Name:     "read_users",
		ParentID: &parentID,
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	if repo.savedPerm == nil {
		t.Fatal("expected permission to be saved")
	}

	if repo.savedPerm.ParentID() == nil {
		t.Fatal("expected non-nil parent ID")
	}

	if *repo.savedPerm.ParentID() != parentUUID {
		t.Errorf("expected parent ID %s, got %s", parentUUID, *repo.savedPerm.ParentID())
	}
}

func TestCreatePermissionHandler_WithoutDescription(t *testing.T) {
	t.Parallel()

	repo := &mockPermissionRepository{}
	log := &mockLogger{}

	handler := NewCreatePermissionHandler(repo, log)

	cmd := CreatePermissionCommand{
		Name: "write",
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	if repo.savedPerm == nil {
		t.Fatal("expected permission to be saved")
	}

	if repo.savedPerm.Description() != nil {
		t.Error("expected nil description")
	}
}
