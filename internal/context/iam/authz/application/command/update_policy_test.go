package command

import (
	"context"
	"errors"
	"testing"
	"time"

	"gct/internal/context/iam/authz/domain"

	"github.com/google/uuid"
)

func TestUpdatePolicyHandler_UpdateFields(t *testing.T) {
	policyID := uuid.New()
	permID := uuid.New()
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

	handler := NewUpdatePolicyHandler(repo, log)

	newEffect := domain.PolicyDeny
	newPriority := 99
	newConditions := map[string]any{"env": "production"}

	cmd := UpdatePolicyCommand{
		ID:         policyID,
		Effect:     &newEffect,
		Priority:   &newPriority,
		Conditions: newConditions,
	}

	err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if repo.updatedPolicy == nil {
		t.Fatal("expected policy to be updated")
	}

	if repo.updatedPolicy.Effect() != domain.PolicyDeny {
		t.Errorf("expected effect DENY, got '%s'", repo.updatedPolicy.Effect())
	}

	if repo.updatedPolicy.Priority() != 99 {
		t.Errorf("expected priority 99, got %d", repo.updatedPolicy.Priority())
	}

	if repo.updatedPolicy.Conditions()["env"] != "production" {
		t.Errorf("expected condition env=production, got '%v'", repo.updatedPolicy.Conditions()["env"])
	}
}

func TestUpdatePolicyHandler_NotFound(t *testing.T) {
	repo := &mockPolicyRepository{} // default returns ErrPolicyNotFound
	log := &mockLogger{}

	handler := NewUpdatePolicyHandler(repo, log)

	newEffect := domain.PolicyAllow
	cmd := UpdatePolicyCommand{
		ID:     uuid.New(),
		Effect: &newEffect,
	}

	err := handler.Handle(context.Background(), cmd)
	if !errors.Is(err, domain.ErrPolicyNotFound) {
		t.Fatalf("expected ErrPolicyNotFound, got: %v", err)
	}

	if repo.updatedPolicy != nil {
		t.Error("expected no policy to be updated when not found")
	}
}
