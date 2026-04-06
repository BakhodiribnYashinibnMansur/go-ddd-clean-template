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

func TestUpdatePolicyHandler_UpdateFields(t *testing.T) {
	t.Parallel()

	policyID := authzentity.NewPolicyID()
	permID := authzentity.NewPermissionID()
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

	handler := NewUpdatePolicyHandler(repo, log)

	newEffect := authzentity.PolicyDeny
	newPriority := 99
	newConditions := map[string]any{"env": "production"}

	cmd := UpdatePolicyCommand{
		ID:         authzentity.PolicyID(policyID),
		Effect:     &newEffect,
		Priority:   &newPriority,
		Conditions: newConditions,
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	if repo.updatedPolicy == nil {
		t.Fatal("expected policy to be updated")
	}

	if repo.updatedPolicy.Effect() != authzentity.PolicyDeny {
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
	t.Parallel()

	repo := &mockPolicyRepository{} // default returns ErrPolicyNotFound
	log := &mockLogger{}

	handler := NewUpdatePolicyHandler(repo, log)

	newEffect := authzentity.PolicyAllow
	cmd := UpdatePolicyCommand{
		ID:     authzentity.PolicyID(uuid.New()),
		Effect: &newEffect,
	}

	err := handler.Handle(context.Background(), cmd)
	if !errors.Is(err, authzentity.ErrPolicyNotFound) {
		t.Fatalf("expected ErrPolicyNotFound, got: %v", err)
	}

	if repo.updatedPolicy != nil {
		t.Error("expected no policy to be updated when not found")
	}
}
