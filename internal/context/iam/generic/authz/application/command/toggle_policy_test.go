package command

import (
	"context"
	"errors"
	"testing"
	"time"

	authzentity "gct/internal/context/iam/generic/authz/domain/entity"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestTogglePolicyHandler_ToggleActive(t *testing.T) {
	t.Parallel()

	policyID := authzentity.NewPolicyID()
	permID := authzentity.NewPermissionID()
	// Start with active=true
	existingPolicy := authzentity.ReconstructPolicy(
		policyID.UUID(), time.Now(), time.Now(), nil,
		permID.UUID(), authzentity.PolicyAllow, 1, true, nil,
	)

	repo := &mockPolicyRepository{
		findByIDFn: func(_ context.Context, id authzentity.PolicyID) (*authzentity.Policy, error) {
			if id == policyID {
				return existingPolicy, nil
			}
			return nil, authzentity.ErrPolicyNotFound
		},
	}
	log := &mockLogger{}

	handler := NewTogglePolicyHandler(repo, log)

	cmd := TogglePolicyCommand{ID: authzentity.PolicyID(policyID)}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	if repo.updatedPolicy == nil {
		t.Fatal("expected policy to be updated")
	}

	// Should now be inactive
	if repo.updatedPolicy.IsActive() {
		t.Error("expected policy to be inactive after toggle")
	}
}

func TestTogglePolicyHandler_ToggleInactiveToActive(t *testing.T) {
	t.Parallel()

	policyID := authzentity.NewPolicyID()
	permID := authzentity.NewPermissionID()
	// Start with active=false
	existingPolicy := authzentity.ReconstructPolicy(
		policyID.UUID(), time.Now(), time.Now(), nil,
		permID.UUID(), authzentity.PolicyDeny, 5, false, nil,
	)

	repo := &mockPolicyRepository{
		findByIDFn: func(_ context.Context, id authzentity.PolicyID) (*authzentity.Policy, error) {
			if id == policyID {
				return existingPolicy, nil
			}
			return nil, authzentity.ErrPolicyNotFound
		},
	}
	log := &mockLogger{}

	handler := NewTogglePolicyHandler(repo, log)

	cmd := TogglePolicyCommand{ID: authzentity.PolicyID(policyID)}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	if repo.updatedPolicy == nil {
		t.Fatal("expected policy to be updated")
	}

	if !repo.updatedPolicy.IsActive() {
		t.Error("expected policy to be active after toggle from inactive")
	}
}

func TestTogglePolicyHandler_NotFound(t *testing.T) {
	t.Parallel()

	repo := &mockPolicyRepository{} // default returns ErrPolicyNotFound
	log := &mockLogger{}

	handler := NewTogglePolicyHandler(repo, log)

	cmd := TogglePolicyCommand{ID: authzentity.PolicyID(uuid.New())}

	err := handler.Handle(context.Background(), cmd)
	if !errors.Is(err, authzentity.ErrPolicyNotFound) {
		t.Fatalf("expected ErrPolicyNotFound, got: %v", err)
	}

	if repo.updatedPolicy != nil {
		t.Error("expected no policy to be updated when not found")
	}
}
