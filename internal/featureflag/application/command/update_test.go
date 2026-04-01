package command

import (
	"context"
	"errors"
	"testing"

	"gct/internal/featureflag/domain"

	"github.com/google/uuid"
)

func TestUpdateHandler_Handle(t *testing.T) {
	flagID := uuid.New()
	flagRepo := &mockFeatureFlagRepo{
		findFn: func(_ context.Context, id uuid.UUID) (*domain.FeatureFlag, error) {
			if id == flagID {
				return newReconstructedFlag(flagID), nil
			}
			return nil, domain.ErrFeatureFlagNotFound
		},
	}
	eb := &mockEventBus{}
	handler := NewUpdateHandler(flagRepo, eb, &mockLogger{})

	newName := "updated-name"
	newKey := "updated_key"
	cmd := UpdateCommand{
		ID:   flagID,
		Name: &newName,
		Key:  &newKey,
	}

	err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if flagRepo.updated == nil {
		t.Fatal("expected feature flag to be updated")
	}
	if flagRepo.updated.Name() != "updated-name" {
		t.Errorf("expected name updated-name, got %s", flagRepo.updated.Name())
	}
	if flagRepo.updated.Key() != "updated_key" {
		t.Errorf("expected key updated_key, got %s", flagRepo.updated.Key())
	}
	if len(eb.published) == 0 {
		t.Error("expected events to be published")
	}
}

func TestUpdateHandler_Handle_PartialUpdate(t *testing.T) {
	flagID := uuid.New()
	flagRepo := &mockFeatureFlagRepo{
		findFn: func(_ context.Context, _ uuid.UUID) (*domain.FeatureFlag, error) {
			return newReconstructedFlag(flagID), nil
		},
	}
	handler := NewUpdateHandler(flagRepo, &mockEventBus{}, &mockLogger{})

	newDesc := "new description"
	err := handler.Handle(context.Background(), UpdateCommand{
		ID:          flagID,
		Description: &newDesc,
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if flagRepo.updated == nil {
		t.Fatal("expected feature flag to be updated")
	}
	if flagRepo.updated.Description() != "new description" {
		t.Errorf("expected description 'new description', got %s", flagRepo.updated.Description())
	}
	// Other fields should remain unchanged
	if flagRepo.updated.Name() != "test-flag" {
		t.Errorf("expected name test-flag (unchanged), got %s", flagRepo.updated.Name())
	}
}

func TestUpdateHandler_Handle_NotFound(t *testing.T) {
	flagRepo := &mockFeatureFlagRepo{} // default returns ErrFeatureFlagNotFound
	handler := NewUpdateHandler(flagRepo, &mockEventBus{}, &mockLogger{})

	err := handler.Handle(context.Background(), UpdateCommand{ID: uuid.New()})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrFeatureFlagNotFound) {
		t.Fatalf("expected ErrFeatureFlagNotFound, got: %v", err)
	}
}

func TestUpdateHandler_Handle_UpdateRepoError(t *testing.T) {
	flagID := uuid.New()
	repoErr := errors.New("update failed")
	flagRepo := &mockFeatureFlagRepo{
		findFn: func(_ context.Context, _ uuid.UUID) (*domain.FeatureFlag, error) {
			return newReconstructedFlag(flagID), nil
		},
		updateFn: func(_ context.Context, _ *domain.FeatureFlag) error {
			return repoErr
		},
	}
	handler := NewUpdateHandler(flagRepo, &mockEventBus{}, &mockLogger{})

	err := handler.Handle(context.Background(), UpdateCommand{ID: flagID})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got: %v", err)
	}
}

func TestUpdateHandler_Handle_ToggleActive(t *testing.T) {
	flagID := uuid.New()
	flagRepo := &mockFeatureFlagRepo{
		findFn: func(_ context.Context, _ uuid.UUID) (*domain.FeatureFlag, error) {
			return newReconstructedFlag(flagID), nil
		},
	}
	handler := NewUpdateHandler(flagRepo, &mockEventBus{}, &mockLogger{})

	isActive := false
	err := handler.Handle(context.Background(), UpdateCommand{
		ID:       flagID,
		IsActive: &isActive,
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if flagRepo.updated.IsActive() {
		t.Error("expected flag to be inactive after update")
	}
}
