package integrationmod

import (
	"context"
	"testing"

	"gct/internal/integration"
	"gct/internal/integration/application/command"
	"gct/internal/integration/application/query"
	"gct/internal/integration/domain"
	"gct/internal/shared/infrastructure/eventbus"
	"gct/internal/shared/infrastructure/logger"
	"gct/test/integration/common/setup"
)

func newTestBC(t *testing.T) *integration.BoundedContext {
	t.Helper()
	eb := eventbus.NewInMemoryEventBus()
	l := logger.New("error")
	return integration.NewBoundedContext(setup.TestPG.Pool, eb, l)
}

func TestIntegration_CreateAndGetIntegration(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	err := bc.CreateIntegration.Handle(ctx, command.CreateCommand{
		Name:       "test-integration",
		Type:       "webhook",
		APIKey:     "key-123",
		WebhookURL: "https://example.com/hook",
		Enabled:    true,
		Config:     map[string]string{"channel": "#general"},
	})
	if err != nil {
		t.Fatalf("CreateIntegration: %v", err)
	}

	result, err := bc.ListIntegrations.Handle(ctx, query.ListQuery{
		Filter: domain.IntegrationFilter{Limit: 10},
	})
	if err != nil {
		t.Fatalf("ListIntegrations: %v", err)
	}
	if result.Total != 1 {
		t.Fatalf("expected 1 integration, got %d", result.Total)
	}

	ig := result.Integrations[0]
	if ig.Name != "test-integration" {
		t.Errorf("expected name test-integration, got %s", ig.Name)
	}
	if ig.Type != "webhook" {
		t.Errorf("expected type webhook, got %s", ig.Type)
	}

	view, err := bc.GetIntegration.Handle(ctx, query.GetQuery{ID: ig.ID})
	if err != nil {
		t.Fatalf("GetIntegration: %v", err)
	}
	if view.ID != ig.ID {
		t.Errorf("ID mismatch: %s vs %s", view.ID, ig.ID)
	}
}

func TestIntegration_UpdateIntegration(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	err := bc.CreateIntegration.Handle(ctx, command.CreateCommand{
		Name:       "original",
		Type:       "slack",
		APIKey:     "key-abc",
		WebhookURL: "https://example.com/original",
		Enabled:    true,
		Config:     map[string]string{},
	})
	if err != nil {
		t.Fatalf("CreateIntegration: %v", err)
	}

	list, _ := bc.ListIntegrations.Handle(ctx, query.ListQuery{
		Filter: domain.IntegrationFilter{Limit: 10},
	})
	igID := list.Integrations[0].ID

	newName := "updated-integration"
	newURL := "https://example.com/updated"
	err = bc.UpdateIntegration.Handle(ctx, command.UpdateCommand{
		ID:         igID,
		Name:       &newName,
		WebhookURL: &newURL,
	})
	if err != nil {
		t.Fatalf("UpdateIntegration: %v", err)
	}

	view, _ := bc.GetIntegration.Handle(ctx, query.GetQuery{ID: igID})
	if view.Name != "updated-integration" {
		t.Errorf("name not updated, got %s", view.Name)
	}
	if view.WebhookURL != "https://example.com/updated" {
		t.Errorf("webhook URL not updated, got %s", view.WebhookURL)
	}
}

func TestIntegration_DeleteIntegration(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	err := bc.CreateIntegration.Handle(ctx, command.CreateCommand{
		Name:       "to-delete",
		Type:       "email",
		APIKey:     "key-del",
		WebhookURL: "https://example.com/delete",
		Enabled:    true,
		Config:     map[string]string{},
	})
	if err != nil {
		t.Fatalf("CreateIntegration: %v", err)
	}

	list, _ := bc.ListIntegrations.Handle(ctx, query.ListQuery{
		Filter: domain.IntegrationFilter{Limit: 10},
	})
	igID := list.Integrations[0].ID

	err = bc.DeleteIntegration.Handle(ctx, command.DeleteCommand{ID: igID})
	if err != nil {
		t.Fatalf("DeleteIntegration: %v", err)
	}

	list2, _ := bc.ListIntegrations.Handle(ctx, query.ListQuery{
		Filter: domain.IntegrationFilter{Limit: 10},
	})
	if list2.Total != 0 {
		t.Errorf("expected 0 integrations after delete, got %d", list2.Total)
	}
}
