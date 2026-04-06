package query

import (
	"context"
	"errors"
	"gct/internal/kernel/infrastructure/logger"
	"testing"

	authzentity "gct/internal/context/iam/generic/authz/domain/entity"
	authzrepo "gct/internal/context/iam/generic/authz/domain/repository"
	shared "gct/internal/kernel/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestListPoliciesHandler_WithResults(t *testing.T) {
	t.Parallel()

	permID := authzentity.NewPermissionID()
	repo := &mockAuthzReadRepository{
		listPoliciesFn: func(_ context.Context, _ shared.Pagination) ([]*authzrepo.PolicyView, int64, error) {
			return []*authzrepo.PolicyView{
				{
					ID:           authzentity.NewPolicyID(),
					PermissionID: permID,
					Effect:       "ALLOW",
					Priority:     10,
					Active:       true,
					Conditions:   map[string]any{"ip_range": "10.0.0.0/8"},
				},
				{
					ID:           authzentity.NewPolicyID(),
					PermissionID: permID,
					Effect:       "DENY",
					Priority:     20,
					Active:       false,
					Conditions:   nil,
				},
			}, 2, nil
		},
	}

	handler := NewListPoliciesHandler(repo, logger.Noop())
	result, err := handler.Handle(context.Background(), ListPoliciesQuery{
		Pagination: shared.Pagination{Limit: 10, Offset: 0},
	})
	require.NoError(t, err)

	if len(result.Policies) != 2 {
		t.Fatalf("expected 2 policies, got %d", len(result.Policies))
	}
	if result.Total != 2 {
		t.Errorf("expected total 2, got %d", result.Total)
	}
	if result.Policies[0].Effect != "ALLOW" {
		t.Errorf("expected first policy effect 'ALLOW', got '%s'", result.Policies[0].Effect)
	}
	if result.Policies[0].Priority != 10 {
		t.Errorf("expected first policy priority 10, got %d", result.Policies[0].Priority)
	}
	if !result.Policies[0].Active {
		t.Error("expected first policy to be active")
	}
	if result.Policies[0].PermissionID != uuid.UUID(permID) {
		t.Errorf("expected permission_id %s, got %s", permID, result.Policies[0].PermissionID)
	}
	if result.Policies[1].Effect != "DENY" {
		t.Errorf("expected second policy effect 'DENY', got '%s'", result.Policies[1].Effect)
	}
	if result.Policies[1].Active {
		t.Error("expected second policy to be inactive")
	}
}

func TestListPoliciesHandler_Empty(t *testing.T) {
	t.Parallel()

	repo := &mockAuthzReadRepository{
		listPoliciesFn: func(_ context.Context, _ shared.Pagination) ([]*authzrepo.PolicyView, int64, error) {
			return []*authzrepo.PolicyView{}, 0, nil
		},
	}

	handler := NewListPoliciesHandler(repo, logger.Noop())
	result, err := handler.Handle(context.Background(), ListPoliciesQuery{
		Pagination: shared.Pagination{Limit: 10, Offset: 0},
	})
	require.NoError(t, err)

	if len(result.Policies) != 0 {
		t.Errorf("expected 0 policies, got %d", len(result.Policies))
	}
	if result.Total != 0 {
		t.Errorf("expected total 0, got %d", result.Total)
	}
}

func TestListPoliciesHandler_RepoError(t *testing.T) {
	t.Parallel()

	repoErr := errors.New("policy query failed")
	repo := &mockAuthzReadRepository{
		listPoliciesFn: func(_ context.Context, _ shared.Pagination) ([]*authzrepo.PolicyView, int64, error) {
			return nil, 0, repoErr
		},
	}

	handler := NewListPoliciesHandler(repo, logger.Noop())
	_, err := handler.Handle(context.Background(), ListPoliciesQuery{
		Pagination: shared.Pagination{Limit: 10, Offset: 0},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, repoErr) {
		t.Errorf("expected repo error, got: %v", err)
	}
}
