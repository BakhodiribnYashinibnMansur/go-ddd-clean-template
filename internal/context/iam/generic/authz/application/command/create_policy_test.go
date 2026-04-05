package command

import (
	"context"
	"testing"

	"gct/internal/context/iam/generic/authz/domain"
	shared "gct/internal/kernel/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// --- Mock PolicyRepository ---

type mockPolicyRepository struct {
	savedPolicy   *domain.Policy
	updatedPolicy *domain.Policy
	findByIDFn    func(ctx context.Context, id uuid.UUID) (*domain.Policy, error)
	saveFn        func(ctx context.Context, policy *domain.Policy) error
	updateFn      func(ctx context.Context, policy *domain.Policy) error
	deleteFn      func(ctx context.Context, id uuid.UUID) error
}

func (m *mockPolicyRepository) Save(ctx context.Context, policy *domain.Policy) error {
	if m.saveFn != nil {
		return m.saveFn(ctx, policy)
	}
	m.savedPolicy = policy
	return nil
}

func (m *mockPolicyRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Policy, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, domain.ErrPolicyNotFound
}

func (m *mockPolicyRepository) Update(ctx context.Context, policy *domain.Policy) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, policy)
	}
	m.updatedPolicy = policy
	return nil
}

func (m *mockPolicyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func (m *mockPolicyRepository) List(ctx context.Context, pagination shared.Pagination) ([]*domain.Policy, int64, error) {
	return nil, 0, nil
}

func (m *mockPolicyRepository) FindByPermissionID(ctx context.Context, permissionID uuid.UUID) ([]*domain.Policy, error) {
	return nil, nil
}

// --- Tests ---

func TestCreatePolicyHandler_AllowEffect(t *testing.T) {
	t.Parallel()

	repo := &mockPolicyRepository{}
	log := &mockLogger{}

	handler := NewCreatePolicyHandler(repo, log)

	permID := uuid.New()
	cmd := CreatePolicyCommand{
		PermissionID: domain.PermissionID(permID),
		Effect:       domain.PolicyAllow,
		Priority:     10,
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	if repo.savedPolicy == nil {
		t.Fatal("expected policy to be saved")
	}

	if repo.savedPolicy.Effect() != domain.PolicyAllow {
		t.Errorf("expected effect ALLOW, got '%s'", repo.savedPolicy.Effect())
	}

	if repo.savedPolicy.Priority() != 10 {
		t.Errorf("expected priority 10, got %d", repo.savedPolicy.Priority())
	}

	if repo.savedPolicy.PermissionID() != permID {
		t.Errorf("expected permission ID %s, got %s", permID, repo.savedPolicy.PermissionID())
	}

	if !repo.savedPolicy.IsActive() {
		t.Error("expected policy to be active by default")
	}
}

func TestCreatePolicyHandler_DenyEffect(t *testing.T) {
	t.Parallel()

	repo := &mockPolicyRepository{}
	log := &mockLogger{}

	handler := NewCreatePolicyHandler(repo, log)

	cmd := CreatePolicyCommand{
		PermissionID: domain.PermissionID(uuid.New()),
		Effect:       domain.PolicyDeny,
		Priority:     5,
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	if repo.savedPolicy == nil {
		t.Fatal("expected policy to be saved")
	}

	if repo.savedPolicy.Effect() != domain.PolicyDeny {
		t.Errorf("expected effect DENY, got '%s'", repo.savedPolicy.Effect())
	}
}

func TestCreatePolicyHandler_WithConditions(t *testing.T) {
	t.Parallel()

	repo := &mockPolicyRepository{}
	log := &mockLogger{}

	handler := NewCreatePolicyHandler(repo, log)

	conditions := map[string]any{
		"ip_range": "10.0.0.0/8",
		"max_age":  "3600",
	}

	cmd := CreatePolicyCommand{
		PermissionID: domain.PermissionID(uuid.New()),
		Effect:       domain.PolicyAllow,
		Priority:     1,
		Conditions:   conditions,
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	if repo.savedPolicy == nil {
		t.Fatal("expected policy to be saved")
	}

	conds := repo.savedPolicy.Conditions()
	if conds == nil {
		t.Fatal("expected conditions to be set")
	}

	if conds["ip_range"] != "10.0.0.0/8" {
		t.Errorf("expected ip_range '10.0.0.0/8', got '%v'", conds["ip_range"])
	}

	if conds["max_age"] != "3600" {
		t.Errorf("expected max_age 3600, got '%v'", conds["max_age"])
	}
}

func TestCreatePolicyHandler_NilConditions(t *testing.T) {
	t.Parallel()

	repo := &mockPolicyRepository{}
	log := &mockLogger{}

	handler := NewCreatePolicyHandler(repo, log)

	cmd := CreatePolicyCommand{
		PermissionID: domain.PermissionID(uuid.New()),
		Effect:       domain.PolicyAllow,
		Priority:     0,
		Conditions:   nil,
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	if repo.savedPolicy == nil {
		t.Fatal("expected policy to be saved")
	}

	// Conditions should remain as initialized empty map (nil was not passed through)
	conds := repo.savedPolicy.Conditions()
	if conds == nil {
		t.Error("expected non-nil conditions map")
	}
	if len(conds) != 0 {
		t.Errorf("expected empty conditions map, got %d entries", len(conds))
	}
}
