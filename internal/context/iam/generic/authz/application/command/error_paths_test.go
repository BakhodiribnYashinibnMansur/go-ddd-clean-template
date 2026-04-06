package command

import (
	"context"
	"errors"
	"testing"
	"time"

	authzentity "gct/internal/context/iam/generic/authz/domain/entity"

	"github.com/google/uuid"
)

var errDB = errors.New("database connection failed")

// --- Role error paths ---

func TestCreateRoleHandler_SaveError(t *testing.T) {
	t.Parallel()

	repo := &mockRoleRepository{}
	// Override Save to return an error
	errRepo := &errMockRoleRepository{mockRoleRepository: repo, saveErr: errDB}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewCreateRoleHandler(errRepo, eventBus, log)

	cmd := CreateRoleCommand{Name: "fail_role"}

	err := handler.Handle(context.Background(), cmd)
	if !errors.Is(err, errDB) {
		t.Fatalf("expected errDB, got: %v", err)
	}
}

type errMockRoleRepository struct {
	*mockRoleRepository
	saveErr   error
	updateErr error
	deleteErr error
}

func (m *errMockRoleRepository) Save(ctx context.Context, role *authzentity.Role) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	return m.mockRoleRepository.Save(ctx, role)
}

func (m *errMockRoleRepository) Update(ctx context.Context, role *authzentity.Role) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	return m.mockRoleRepository.Update(ctx, role)
}

func (m *errMockRoleRepository) Delete(ctx context.Context, id authzentity.RoleID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	return m.mockRoleRepository.Delete(ctx, id)
}

func TestUpdateRoleHandler_UpdateError(t *testing.T) {
	t.Parallel()

	roleID := authzentity.NewRoleID()
	existingRole := authzentity.ReconstructRole(roleID.UUID(), time.Now(), time.Now(), nil, "admin", nil, nil)

	repo := &errMockRoleRepository{
		mockRoleRepository: &mockRoleRepository{
			findByIDFn: func(_ context.Context, id authzentity.RoleID) (*authzentity.Role, error) {
				if id == roleID {
					return existingRole, nil
				}
				return nil, authzentity.ErrRoleNotFound
			},
		},
		updateErr: errDB,
	}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewUpdateRoleHandler(repo, eventBus, log)

	newName := "updated"
	cmd := UpdateRoleCommand{ID: authzentity.RoleID(roleID), Name: &newName}

	err := handler.Handle(context.Background(), cmd)
	if !errors.Is(err, errDB) {
		t.Fatalf("expected errDB, got: %v", err)
	}
}

func TestDeleteRoleHandler_DeleteError(t *testing.T) {
	t.Parallel()

	repo := &errMockRoleRepository{
		mockRoleRepository: &mockRoleRepository{},
		deleteErr:          errDB,
	}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewDeleteRoleHandler(repo, eventBus, log)

	cmd := DeleteRoleCommand{ID: authzentity.RoleID(uuid.New())}

	err := handler.Handle(context.Background(), cmd)
	if !errors.Is(err, errDB) {
		t.Fatalf("expected errDB, got: %v", err)
	}

	if len(eventBus.publishedEvents) != 0 {
		t.Error("expected no events to be published on delete error")
	}
}

// --- Permission error paths ---

func TestCreatePermissionHandler_SaveError(t *testing.T) {
	t.Parallel()

	repo := &mockPermissionRepository{
		saveFn: func(_ context.Context, _ *authzentity.Permission) error {
			return errDB
		},
	}
	log := &mockLogger{}

	handler := NewCreatePermissionHandler(repo, log)

	cmd := CreatePermissionCommand{Name: "fail_perm"}

	err := handler.Handle(context.Background(), cmd)
	if !errors.Is(err, errDB) {
		t.Fatalf("expected errDB, got: %v", err)
	}
}

func TestDeletePermissionHandler_DeleteError(t *testing.T) {
	t.Parallel()

	repo := &mockPermissionRepository{
		deleteFn: func(_ context.Context, _ authzentity.PermissionID) error {
			return errDB
		},
	}
	log := &mockLogger{}

	handler := NewDeletePermissionHandler(repo, log)

	cmd := DeletePermissionCommand{ID: authzentity.PermissionID(uuid.New())}

	err := handler.Handle(context.Background(), cmd)
	if !errors.Is(err, errDB) {
		t.Fatalf("expected errDB, got: %v", err)
	}
}

// --- Policy error paths ---

func TestCreatePolicyHandler_SaveError(t *testing.T) {
	t.Parallel()

	repo := &mockPolicyRepository{
		saveFn: func(_ context.Context, _ *authzentity.Policy) error {
			return errDB
		},
	}
	log := &mockLogger{}

	handler := NewCreatePolicyHandler(repo, log)

	cmd := CreatePolicyCommand{
		PermissionID: authzentity.PermissionID(uuid.New()),
		Effect:       authzentity.PolicyAllow,
		Priority:     1,
	}

	err := handler.Handle(context.Background(), cmd)
	if !errors.Is(err, errDB) {
		t.Fatalf("expected errDB, got: %v", err)
	}
}

func TestUpdatePolicyHandler_UpdateError(t *testing.T) {
	t.Parallel()

	policyID := authzentity.NewPolicyID()
	existingPolicy := authzentity.ReconstructPolicy(
		policyID.UUID(), time.Now(), time.Now(), nil,
		uuid.New(), authzentity.PolicyAllow, 1, true, nil,
	)

	repo := &mockPolicyRepository{
		findByIDFn: func(_ context.Context, id authzentity.PolicyID) (*authzentity.Policy, error) {
			if id == policyID {
				return existingPolicy, nil
			}
			return nil, authzentity.ErrPolicyNotFound
		},
		updateFn: func(_ context.Context, _ *authzentity.Policy) error {
			return errDB
		},
	}
	log := &mockLogger{}

	handler := NewUpdatePolicyHandler(repo, log)

	newEffect := authzentity.PolicyDeny
	cmd := UpdatePolicyCommand{ID: authzentity.PolicyID(policyID), Effect: &newEffect}

	err := handler.Handle(context.Background(), cmd)
	if !errors.Is(err, errDB) {
		t.Fatalf("expected errDB, got: %v", err)
	}
}

func TestDeletePolicyHandler_DeleteError(t *testing.T) {
	t.Parallel()

	repo := &mockPolicyRepository{
		deleteFn: func(_ context.Context, _ authzentity.PolicyID) error {
			return errDB
		},
	}
	log := &mockLogger{}

	handler := NewDeletePolicyHandler(repo, log)

	cmd := DeletePolicyCommand{ID: authzentity.PolicyID(uuid.New())}

	err := handler.Handle(context.Background(), cmd)
	if !errors.Is(err, errDB) {
		t.Fatalf("expected errDB, got: %v", err)
	}
}

func TestTogglePolicyHandler_UpdateError(t *testing.T) {
	t.Parallel()

	policyID := authzentity.NewPolicyID()
	existingPolicy := authzentity.ReconstructPolicy(
		policyID.UUID(), time.Now(), time.Now(), nil,
		uuid.New(), authzentity.PolicyAllow, 1, true, nil,
	)

	repo := &mockPolicyRepository{
		findByIDFn: func(_ context.Context, id authzentity.PolicyID) (*authzentity.Policy, error) {
			if id == policyID {
				return existingPolicy, nil
			}
			return nil, authzentity.ErrPolicyNotFound
		},
		updateFn: func(_ context.Context, _ *authzentity.Policy) error {
			return errDB
		},
	}
	log := &mockLogger{}

	handler := NewTogglePolicyHandler(repo, log)

	cmd := TogglePolicyCommand{ID: authzentity.PolicyID(policyID)}

	err := handler.Handle(context.Background(), cmd)
	if !errors.Is(err, errDB) {
		t.Fatalf("expected errDB, got: %v", err)
	}
}

// --- Scope error paths ---

func TestCreateScopeHandler_SaveError(t *testing.T) {
	t.Parallel()

	repo := &mockScopeRepository{
		saveFn: func(_ context.Context, _ authzentity.Scope) error {
			return errDB
		},
	}
	log := &mockLogger{}

	handler := NewCreateScopeHandler(repo, log)

	cmd := CreateScopeCommand{Path: "/fail", Method: "GET"}

	err := handler.Handle(context.Background(), cmd)
	if !errors.Is(err, errDB) {
		t.Fatalf("expected errDB, got: %v", err)
	}
}

func TestDeleteScopeHandler_DeleteError(t *testing.T) {
	t.Parallel()

	repo := &mockScopeRepository{
		deleteFn: func(_ context.Context, _, _ string) error {
			return errDB
		},
	}
	log := &mockLogger{}

	handler := NewDeleteScopeHandler(repo, log)

	cmd := DeleteScopeCommand{Path: "/fail", Method: "DELETE"}

	err := handler.Handle(context.Background(), cmd)
	if !errors.Is(err, errDB) {
		t.Fatalf("expected errDB, got: %v", err)
	}
}

// --- Assign error paths ---

func TestAssignPermissionHandler_AssignError(t *testing.T) {
	t.Parallel()

	repo := &mockRolePermissionRepository{
		assignFn: func(_ context.Context, _ authzentity.RoleID, _ authzentity.PermissionID) error {
			return errDB
		},
	}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewAssignPermissionHandler(repo, eventBus, log)

	cmd := AssignPermissionCommand{
		RoleID:       authzentity.RoleID(uuid.New()),
		PermissionID: authzentity.PermissionID(uuid.New()),
	}

	err := handler.Handle(context.Background(), cmd)
	if !errors.Is(err, errDB) {
		t.Fatalf("expected errDB, got: %v", err)
	}

	if len(eventBus.publishedEvents) != 0 {
		t.Error("expected no events to be published on assign error")
	}
}

func TestAssignScopeHandler_AssignError(t *testing.T) {
	t.Parallel()

	repo := &mockPermissionScopeRepository{
		assignFn: func(_ context.Context, _ authzentity.PermissionID, _, _ string) error {
			return errDB
		},
	}
	log := &mockLogger{}

	handler := NewAssignScopeHandler(repo, log)

	cmd := AssignScopeCommand{
		PermissionID: authzentity.PermissionID(uuid.New()),
		Path:         "/fail",
		Method:       "POST",
	}

	err := handler.Handle(context.Background(), cmd)
	if !errors.Is(err, errDB) {
		t.Fatalf("expected errDB, got: %v", err)
	}
}
