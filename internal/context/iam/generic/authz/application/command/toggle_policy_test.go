package command

import (
	"context"
	"errors"
	"testing"
	"time"

	"gct/internal/context/iam/generic/authz/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestTogglePolicyHandler_ToggleActive(t *testing.T) {
	t.Parallel()

	policyID := uuid.New()
	permID := uuid.New()
	// Start with active=true
	existingPolicy := domain.ReconstructPolicy(
		policyID, time.Now(), time.Now(), nil,
		permID, domain.PolicyAllow, 1, true, nil,
	)

	repo := &mockPolicyRepository{
		findByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Policy, error) {
			if id == policyID {
				return existingPolicy, nil
			}
			return nil, domain.ErrPolicyNotFound
		},
	}
	log := &mockLogger{}

	handler := NewTogglePolicyHandler(repo, log)

	cmd := TogglePolicyCommand{ID: domain.PolicyID(policyID)}

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

	policyID := uuid.New()
	permID := uuid.New()
	// Start with active=false
	existingPolicy := domain.ReconstructPolicy(
		policyID, time.Now(), time.Now(), nil,
		permID, domain.PolicyDeny, 5, false, nil,
	)

	repo := &mockPolicyRepository{
		findByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Policy, error) {
			if id == policyID {
				return existingPolicy, nil
			}
			return nil, domain.ErrPolicyNotFound
		},
	}
	log := &mockLogger{}

	handler := NewTogglePolicyHandler(repo, log)

	cmd := TogglePolicyCommand{ID: domain.PolicyID(policyID)}

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

	cmd := TogglePolicyCommand{ID: domain.PolicyID(uuid.New())}

	err := handler.Handle(context.Background(), cmd)
	if !errors.Is(err, domain.ErrPolicyNotFound) {
		t.Fatalf("expected ErrPolicyNotFound, got: %v", err)
	}

	if repo.updatedPolicy != nil {
		t.Error("expected no policy to be updated when not found")
	}
}
