package command

import (
	"context"
	"errors"
	"testing"
	"time"

	"gct/internal/featureflag/domain"

	"github.com/google/uuid"
)

func TestDeleteRuleGroupHandler_Handle(t *testing.T) {
	rgID := uuid.New()
	flagID := uuid.New()
	rg := domain.ReconstructRuleGroup(rgID, flagID, "test-rg", "true", 1, time.Now(), time.Now(), nil)

	rgRepo := &mockRuleGroupRepo{
		findFn: func(_ context.Context, id uuid.UUID) (*domain.RuleGroup, error) {
			if id == rgID {
				return rg, nil
			}
			return nil, domain.ErrRuleGroupNotFound
		},
	}
	eb := &mockEventBus{}
	handler := NewDeleteRuleGroupHandler(rgRepo, eb, &mockLogger{})

	err := handler.Handle(context.Background(), DeleteRuleGroupCommand{ID: rgID})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

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
	rgRepo := &mockRuleGroupRepo{} // default returns ErrRuleGroupNotFound
	handler := NewDeleteRuleGroupHandler(rgRepo, &mockEventBus{}, &mockLogger{})

	err := handler.Handle(context.Background(), DeleteRuleGroupCommand{ID: uuid.New()})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrRuleGroupNotFound) {
		t.Fatalf("expected ErrRuleGroupNotFound, got: %v", err)
	}
}

func TestDeleteRuleGroupHandler_Handle_DeleteRepoError(t *testing.T) {
	rgID := uuid.New()
	flagID := uuid.New()
	rg := domain.ReconstructRuleGroup(rgID, flagID, "test-rg", "true", 1, time.Now(), time.Now(), nil)

	repoErr := errors.New("delete failed")
	rgRepo := &mockRuleGroupRepo{
		findFn: func(_ context.Context, _ uuid.UUID) (*domain.RuleGroup, error) {
			return rg, nil
		},
		deleteFn: func(_ context.Context, _ uuid.UUID) error {
			return repoErr
		},
	}
	handler := NewDeleteRuleGroupHandler(rgRepo, &mockEventBus{}, &mockLogger{})

	err := handler.Handle(context.Background(), DeleteRuleGroupCommand{ID: rgID})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got: %v", err)
	}
}
