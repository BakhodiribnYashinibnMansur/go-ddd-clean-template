package command

import (
	"context"
	"errors"
	"testing"
	"time"

	"gct/internal/context/iam/authz/domain"

	"github.com/google/uuid"
)

var errDB = errors.New("database connection failed")

// --- Role error paths ---

func TestCreateRoleHandler_SaveError(t *testing.T) {
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

func (m *errMockRoleRepository) Save(ctx context.Context, role *domain.Role) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	return m.mockRoleRepository.Save(ctx, role)
}

func (m *errMockRoleRepository) Update(ctx context.Context, role *domain.Role) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	return m.mockRoleRepository.Update(ctx, role)
}

func (m *errMockRoleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	return m.mockRoleRepository.Delete(ctx, id)
}

func TestUpdateRoleHandler_UpdateError(t *testing.T) {
	roleID := uuid.New()
	existingRole := domain.ReconstructRole(roleID, time.Now(), time.Now(), nil, "admin", nil, nil)

	repo := &errMockRoleRepository{
		mockRoleRepository: &mockRoleRepository{
			findByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Role, error) {
				if id == roleID {
					return existingRole, nil
				}
				return nil, domain.ErrRoleNotFound
			},
		},
		updateErr: errDB,
	}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewUpdateRoleHandler(repo, eventBus, log)

	newName := "updated"
	cmd := UpdateRoleCommand{ID: roleID, Name: &newName}

	err := handler.Handle(context.Background(), cmd)
	if !errors.Is(err, errDB) {
		t.Fatalf("expected errDB, got: %v", err)
	}
}

func TestDeleteRoleHandler_DeleteError(t *testing.T) {
	repo := &errMockRoleRepository{
		mockRoleRepository: &mockRoleRepository{},
		deleteErr:          errDB,
	}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewDeleteRoleHandler(repo, eventBus, log)

	cmd := DeleteRoleCommand{ID: uuid.New()}

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
	repo := &mockPermissionRepository{
		saveFn: func(_ context.Context, _ *domain.Permission) error {
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
	repo := &mockPermissionRepository{
		deleteFn: func(_ context.Context, _ uuid.UUID) error {
			return errDB
		},
	}
	log := &mockLogger{}

	handler := NewDeletePermissionHandler(repo, log)

	cmd := DeletePermissionCommand{ID: uuid.New()}

	err := handler.Handle(context.Background(), cmd)
	if !errors.Is(err, errDB) {
		t.Fatalf("expected errDB, got: %v", err)
	}
}

// --- Policy error paths ---

func TestCreatePolicyHandler_SaveError(t *testing.T) {
	repo := &mockPolicyRepository{
		saveFn: func(_ context.Context, _ *domain.Policy) error {
			return errDB
		},
	}
	log := &mockLogger{}

	handler := NewCreatePolicyHandler(repo, log)

	cmd := CreatePolicyCommand{
		PermissionID: uuid.New(),
		Effect:       domain.PolicyAllow,
		Priority:     1,
	}

	err := handler.Handle(context.Background(), cmd)
	if !errors.Is(err, errDB) {
		t.Fatalf("expected errDB, got: %v", err)
	}
}

func TestUpdatePolicyHandler_UpdateError(t *testing.T) {
	policyID := uuid.New()
	existingPolicy := domain.ReconstructPolicy(
		policyID, time.Now(), time.Now(), nil,
		uuid.New(), domain.PolicyAllow, 1, true, nil,
	)

	repo := &mockPolicyRepository{
		findByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Policy, error) {
			if id == policyID {
				return existingPolicy, nil
			}
			return nil, domain.ErrPolicyNotFound
		},
		updateFn: func(_ context.Context, _ *domain.Policy) error {
			return errDB
		},
	}
	log := &mockLogger{}

	handler := NewUpdatePolicyHandler(repo, log)

	newEffect := domain.PolicyDeny
	cmd := UpdatePolicyCommand{ID: policyID, Effect: &newEffect}

	err := handler.Handle(context.Background(), cmd)
	if !errors.Is(err, errDB) {
		t.Fatalf("expected errDB, got: %v", err)
	}
}

func TestDeletePolicyHandler_DeleteError(t *testing.T) {
	repo := &mockPolicyRepository{
		deleteFn: func(_ context.Context, _ uuid.UUID) error {
			return errDB
		},
	}
	log := &mockLogger{}

	handler := NewDeletePolicyHandler(repo, log)

	cmd := DeletePolicyCommand{ID: uuid.New()}

	err := handler.Handle(context.Background(), cmd)
	if !errors.Is(err, errDB) {
		t.Fatalf("expected errDB, got: %v", err)
	}
}

func TestTogglePolicyHandler_UpdateError(t *testing.T) {
	policyID := uuid.New()
	existingPolicy := domain.ReconstructPolicy(
		policyID, time.Now(), time.Now(), nil,
		uuid.New(), domain.PolicyAllow, 1, true, nil,
	)

	repo := &mockPolicyRepository{
		findByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Policy, error) {
			if id == policyID {
				return existingPolicy, nil
			}
			return nil, domain.ErrPolicyNotFound
		},
		updateFn: func(_ context.Context, _ *domain.Policy) error {
			return errDB
		},
	}
	log := &mockLogger{}

	handler := NewTogglePolicyHandler(repo, log)

	cmd := TogglePolicyCommand{ID: policyID}

	err := handler.Handle(context.Background(), cmd)
	if !errors.Is(err, errDB) {
		t.Fatalf("expected errDB, got: %v", err)
	}
}

// --- Scope error paths ---

func TestCreateScopeHandler_SaveError(t *testing.T) {
	repo := &mockScopeRepository{
		saveFn: func(_ context.Context, _ domain.Scope) error {
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
	repo := &mockRolePermissionRepository{
		assignFn: func(_ context.Context, _, _ uuid.UUID) error {
			return errDB
		},
	}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewAssignPermissionHandler(repo, eventBus, log)

	cmd := AssignPermissionCommand{
		RoleID:       uuid.New(),
		PermissionID: uuid.New(),
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
	repo := &mockPermissionScopeRepository{
		assignFn: func(_ context.Context, _ uuid.UUID, _, _ string) error {
			return errDB
		},
	}
	log := &mockLogger{}

	handler := NewAssignScopeHandler(repo, log)

	cmd := AssignScopeCommand{
		PermissionID: uuid.New(),
		Path:         "/fail",
		Method:       "POST",
	}

	err := handler.Handle(context.Background(), cmd)
	if !errors.Is(err, errDB) {
		t.Fatalf("expected errDB, got: %v", err)
	}
}
