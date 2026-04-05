package command

import (
	"context"
	"testing"

	"gct/internal/context/admin/integration/domain"

	"github.com/google/uuid"
)

func TestUpdateHandler_Handle(t *testing.T) {
	i := domain.NewIntegration("Slack", "messaging", "old-key", "https://old.com", true, nil)

	repo := &mockIntegrationRepo{
		findFn: func(_ context.Context, id uuid.UUID) (*domain.Integration, error) {
			if id == i.ID() {
				return i, nil
			}
			return nil, domain.ErrIntegrationNotFound
		},
	}
	eb := &mockEventBus{}
	handler := NewUpdateHandler(repo, eb, &mockLogger{})

	newName := "Slack Updated"
	newKey := "new-key"
	newEnabled := false
	cmd := UpdateCommand{
		ID:      i.ID(),
		Name:    &newName,
		APIKey:  &newKey,
		Enabled: &newEnabled,
	}

	err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if repo.updated == nil {
		t.Fatal("expected integration to be updated")
	}
	if repo.updated.Name() != "Slack Updated" {
		t.Errorf("expected name 'Slack Updated', got %s", repo.updated.Name())
	}
	if repo.updated.APIKey() != "new-key" {
		t.Errorf("expected apiKey new-key, got %s", repo.updated.APIKey())
	}
	if repo.updated.Enabled() {
		t.Error("expected enabled false")
	}
	// Type should remain unchanged
	if repo.updated.Type() != "messaging" {
		t.Errorf("expected type messaging (unchanged), got %s", repo.updated.Type())
	}
}

func TestUpdateHandler_PartialUpdate(t *testing.T) {
	i := domain.NewIntegration("Name", "type", "key", "https://url.com", true, map[string]string{"k": "v"})

	repo := &mockIntegrationRepo{
		findFn: func(_ context.Context, _ uuid.UUID) (*domain.Integration, error) {
			return i, nil
		},
	}
	handler := NewUpdateHandler(repo, &mockEventBus{}, &mockLogger{})

	newURL := "https://new.com"
	err := handler.Handle(context.Background(), UpdateCommand{
		ID:         i.ID(),
		WebhookURL: &newURL,
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if repo.updated.WebhookURL() != "https://new.com" {
		t.Errorf("expected webhookURL https://new.com, got %s", repo.updated.WebhookURL())
	}
	if repo.updated.Name() != "Name" {
		t.Error("name should remain unchanged")
	}
}

func TestUpdateHandler_WithConfig(t *testing.T) {
	i := domain.NewIntegration("Name", "type", "key", "url", true, nil)

	repo := &mockIntegrationRepo{
		findFn: func(_ context.Context, _ uuid.UUID) (*domain.Integration, error) {
			return i, nil
		},
	}
	handler := NewUpdateHandler(repo, &mockEventBus{}, &mockLogger{})

	newConfig := map[string]string{"channel": "#alerts", "priority": "high"}
	err := handler.Handle(context.Background(), UpdateCommand{
		ID:     i.ID(),
		Config: &newConfig,
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if repo.updated.Config()["channel"] != "#alerts" {
		t.Errorf("expected config channel #alerts, got %v", repo.updated.Config()["channel"])
	}
}

func TestUpdateHandler_NotFound(t *testing.T) {
	repo := &mockIntegrationRepo{}
	handler := NewUpdateHandler(repo, &mockEventBus{}, &mockLogger{})

	err := handler.Handle(context.Background(), UpdateCommand{ID: uuid.New()})
	if err == nil {
		t.Fatal("expected error for not found")
	}
}
