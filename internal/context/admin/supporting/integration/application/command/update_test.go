package command

import (
	"context"
	"testing"

	integentity "gct/internal/context/admin/supporting/integration/domain/entity"

	"gct/internal/kernel/outbox"

	"github.com/stretchr/testify/require"
)

func TestUpdateHandler_Handle(t *testing.T) {
	t.Parallel()

	i, _ := integentity.NewIntegration("Slack", "messaging", "old-key", "https://old.com", true, nil)

	repo := &mockIntegrationRepo{
		findFn: func(_ context.Context, id integentity.IntegrationID) (*integentity.Integration, error) {
			if id == i.TypedID() {
				return i, nil
			}
			return nil, integentity.ErrIntegrationNotFound
		},
	}
	eb := &mockEventBus{}
	handler := NewUpdateHandler(repo, outbox.NewEventCommitter(nil, nil, eb, &mockLogger{}), &mockLogger{})

	newName := "Slack Updated"
	newKey := "new-key"
	newEnabled := false
	cmd := UpdateCommand{
		ID:      integentity.IntegrationID(i.ID()),
		Name:    &newName,
		APIKey:  &newKey,
		Enabled: &newEnabled,
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)
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
	t.Parallel()

	i, _ := integentity.NewIntegration("Name", "type", "key", "https://url.com", true, map[string]string{"k": "v"})

	repo := &mockIntegrationRepo{
		findFn: func(_ context.Context, _ integentity.IntegrationID) (*integentity.Integration, error) {
			return i, nil
		},
	}
	handler := NewUpdateHandler(repo, outbox.NewEventCommitter(nil, nil, &mockEventBus{}, &mockLogger{}), &mockLogger{})

	newURL := "https://new.com"
	err := handler.Handle(context.Background(), UpdateCommand{
		ID:         integentity.IntegrationID(i.ID()),
		WebhookURL: &newURL,
	})
	require.NoError(t, err)
	if repo.updated.WebhookURL() != "https://new.com" {
		t.Errorf("expected webhookURL https://new.com, got %s", repo.updated.WebhookURL())
	}
	if repo.updated.Name() != "Name" {
		t.Error("name should remain unchanged")
	}
}

func TestUpdateHandler_WithConfig(t *testing.T) {
	t.Parallel()

	i, _ := integentity.NewIntegration("Name", "type", "key", "url", true, nil)

	repo := &mockIntegrationRepo{
		findFn: func(_ context.Context, _ integentity.IntegrationID) (*integentity.Integration, error) {
			return i, nil
		},
	}
	handler := NewUpdateHandler(repo, outbox.NewEventCommitter(nil, nil, &mockEventBus{}, &mockLogger{}), &mockLogger{})

	newConfig := map[string]string{"channel": "#alerts", "priority": "high"}
	err := handler.Handle(context.Background(), UpdateCommand{
		ID:     integentity.IntegrationID(i.ID()),
		Config: &newConfig,
	})
	require.NoError(t, err)
	if repo.updated.Config()["channel"] != "#alerts" {
		t.Errorf("expected config channel #alerts, got %v", repo.updated.Config()["channel"])
	}
}

func TestUpdateHandler_NotFound(t *testing.T) {
	t.Parallel()

	repo := &mockIntegrationRepo{}
	handler := NewUpdateHandler(repo, outbox.NewEventCommitter(nil, nil, &mockEventBus{}, &mockLogger{}), &mockLogger{})

	err := handler.Handle(context.Background(), UpdateCommand{ID: integentity.NewIntegrationID()})
	if err == nil {
		t.Fatal("expected error for not found")
	}
}
