package command

import (
	"context"
	"errors"
	"testing"
	"time"

	"gct/internal/authz/domain"

	"github.com/google/uuid"
)

func TestTogglePolicyHandler_ToggleActive(t *testing.T) {
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

	cmd := TogglePolicyCommand{ID: policyID}

	err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if repo.updatedPolicy == nil {
		t.Fatal("expected policy to be updated")
	}

	// Should now be inactive
	if repo.updatedPolicy.IsActive() {
		t.Error("expected policy to be inactive after toggle")
	}
}

func TestTogglePolicyHandler_ToggleInactiveToActive(t *testing.T) {
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

	cmd := TogglePolicyCommand{ID: policyID}

	err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if repo.updatedPolicy == nil {
		t.Fatal("expected policy to be updated")
	}

	if !repo.updatedPolicy.IsActive() {
		t.Error("expected policy to be active after toggle from inactive")
	}
}

func TestTogglePolicyHandler_NotFound(t *testing.T) {
	repo := &mockPolicyRepository{} // default returns ErrPolicyNotFound
	log := &mockLogger{}

	handler := NewTogglePolicyHandler(repo, log)

	cmd := TogglePolicyCommand{ID: uuid.New()}

	err := handler.Handle(context.Background(), cmd)
	if !errors.Is(err, domain.ErrPolicyNotFound) {
		t.Fatalf("expected ErrPolicyNotFound, got: %v", err)
	}

	if repo.updatedPolicy != nil {
		t.Error("expected no policy to be updated when not found")
	}
}
