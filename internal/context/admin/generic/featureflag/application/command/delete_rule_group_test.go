package command

import (
	"context"
	"errors"
	"testing"
	"time"

	ffentity "gct/internal/context/admin/generic/featureflag/domain/entity"

	"gct/internal/kernel/outbox"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestDeleteRuleGroupHandler_Handle(t *testing.T) {
	t.Parallel()

	rgID := ffentity.NewRuleGroupID()
	flagID := ffentity.NewFeatureFlagID()
	rg := ffentity.ReconstructRuleGroup(rgID.UUID(), flagID.UUID(), "test-rg", "true", 1, time.Now(), time.Now(), nil)

	rgRepo := &mockRuleGroupRepo{
		findFn: func(_ context.Context, id ffentity.RuleGroupID) (*ffentity.RuleGroup, error) {
			if id == rgID {
				return rg, nil
			}
			return nil, ffentity.ErrRuleGroupNotFound
		},
	}
	eb := &mockEventBus{}
	handler := NewDeleteRuleGroupHandler(rgRepo, outbox.NewEventCommitter(nil, nil, eb, &mockLogger{}), &mockLogger{})

	err := handler.Handle(context.Background(), DeleteRuleGroupCommand{ID: ffentity.RuleGroupID(rgID)})
	require.NoError(t, err)

	if rgRepo.deleted != rgID {
		t.Errorf("expected deleted ID %s, got %s", rgID, rgRepo.deleted)
	}
	if len(eb.published) == 0 {
		t.Error("expected FlagUpdated event to be published")
	}
	if eb.published[0].EventName() != "featureflag.updated" {
		t.Errorf("expected event name featureflag.updated, got %s", eb.published[0].EventName())
	}
}

func TestDeleteRuleGroupHandler_Handle_NotFound(t *testing.T) {
	t.Parallel()

	rgRepo := &mockRuleGroupRepo{} // default returns ErrRuleGroupNotFound
	handler := NewDeleteRuleGroupHandler(rgRepo, outbox.NewEventCommitter(nil, nil, &mockEventBus{}, &mockLogger{}), &mockLogger{})

	err := handler.Handle(context.Background(), DeleteRuleGroupCommand{ID: ffentity.RuleGroupID(uuid.New())})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ffentity.ErrRuleGroupNotFound) {
		t.Fatalf("expected ErrRuleGroupNotFound, got: %v", err)
	}
}

func TestDeleteRuleGroupHandler_Handle_DeleteRepoError(t *testing.T) {
	t.Parallel()

	rgID := ffentity.NewRuleGroupID()
	flagID := ffentity.NewFeatureFlagID()
	rg := ffentity.ReconstructRuleGroup(rgID.UUID(), flagID.UUID(), "test-rg", "true", 1, time.Now(), time.Now(), nil)

	repoErr := errors.New("delete failed")
	rgRepo := &mockRuleGroupRepo{
		findFn: func(_ context.Context, _ ffentity.RuleGroupID) (*ffentity.RuleGroup, error) {
			return rg, nil
		},
		deleteFn: func(_ context.Context, _ ffentity.RuleGroupID) error {
			return repoErr
		},
	}
	handler := NewDeleteRuleGroupHandler(rgRepo, outbox.NewEventCommitter(nil, nil, &mockEventBus{}, &mockLogger{}), &mockLogger{})

	err := handler.Handle(context.Background(), DeleteRuleGroupCommand{ID: ffentity.RuleGroupID(rgID)})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got: %v", err)
	}
}
