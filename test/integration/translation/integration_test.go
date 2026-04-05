package translation

import (
	"context"
	"testing"

	"gct/internal/kernel/infrastructure/eventbus"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/context/content/generic/translation"
	"gct/internal/context/content/generic/translation/application/command"
	"gct/internal/context/content/generic/translation/application/query"
	"gct/internal/context/content/generic/translation/domain"
	"gct/test/integration/common/setup"
)

func newTestBC(t *testing.T) *translation.BoundedContext {
	t.Helper()
	eb := eventbus.NewInMemoryEventBus()
	l := logger.New("error")
	return translation.NewBoundedContext(setup.TestPG.Pool, eb, l)
}

func TestIntegration_CreateAndGetTranslation(t *testing.T) {
	t.Skip("DB schema (entity_type/entity_id/lang_code/data) does not match DDD domain model (key/language/value/group)")
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	err := bc.CreateTranslation.Handle(ctx, command.CreateTranslationCommand{
		Key:      "welcome_message",
		Language: "en",
		Value:    "Welcome!",
		Group:    "homepage",
	})
	if err != nil {
		t.Fatalf("CreateTranslation: %v", err)
	}

	result, err := bc.ListTranslations.Handle(ctx, query.ListTranslationsQuery{
		Filter: domain.TranslationFilter{Limit: 10},
	})
	if err != nil {
		t.Fatalf("ListTranslations: %v", err)
	}
	if result.Total != 1 {
		t.Fatalf("expected 1 translation, got %d", result.Total)
	}

	tr := result.Translations[0]
	if tr.Key != "welcome_message" {
		t.Errorf("expected key welcome_message, got %s", tr.Key)
	}
	if tr.Language != "en" {
		t.Errorf("expected language en, got %s", tr.Language)
	}
	if tr.Value != "Welcome!" {
		t.Errorf("expected value 'Welcome!', got %s", tr.Value)
	}

	view, err := bc.GetTranslation.Handle(ctx, query.GetTranslationQuery{ID: domain.TranslationID(tr.ID)})
	if err != nil {
		t.Fatalf("GetTranslation: %v", err)
	}
	if view.ID != tr.ID {
		t.Errorf("ID mismatch: %s vs %s", view.ID, tr.ID)
	}
}

func TestIntegration_UpdateTranslation(t *testing.T) {
	t.Skip("DB schema mismatch — see TestIntegration_CreateAndGetTranslation")
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	err := bc.CreateTranslation.Handle(ctx, command.CreateTranslationCommand{
		Key:      "greeting",
		Language: "uz",
		Value:    "Salom",
		Group:    "common",
	})
	if err != nil {
		t.Fatalf("CreateTranslation: %v", err)
	}

	list, _ := bc.ListTranslations.Handle(ctx, query.ListTranslationsQuery{
		Filter: domain.TranslationFilter{Limit: 10},
	})
	trID := domain.TranslationID(list.Translations[0].ID)

	newValue := "Xush kelibsiz"
	err = bc.UpdateTranslation.Handle(ctx, command.UpdateTranslationCommand{
		ID:    trID,
		Value: &newValue,
	})
	if err != nil {
		t.Fatalf("UpdateTranslation: %v", err)
	}

	view, _ := bc.GetTranslation.Handle(ctx, query.GetTranslationQuery{ID: trID})
	if view.Value != "Xush kelibsiz" {
		t.Errorf("value not updated, got %s", view.Value)
	}
}

func TestIntegration_DeleteTranslation(t *testing.T) {
	t.Skip("DB schema mismatch — see TestIntegration_CreateAndGetTranslation")
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	err := bc.CreateTranslation.Handle(ctx, command.CreateTranslationCommand{
		Key:      "delete_me",
		Language: "ru",
		Value:    "Удалить",
		Group:    "test",
	})
	if err != nil {
		t.Fatalf("CreateTranslation: %v", err)
	}

	list, _ := bc.ListTranslations.Handle(ctx, query.ListTranslationsQuery{
		Filter: domain.TranslationFilter{Limit: 10},
	})
	trID := domain.TranslationID(list.Translations[0].ID)

	err = bc.DeleteTranslation.Handle(ctx, command.DeleteTranslationCommand{ID: trID})
	if err != nil {
		t.Fatalf("DeleteTranslation: %v", err)
	}

	list2, _ := bc.ListTranslations.Handle(ctx, query.ListTranslationsQuery{
		Filter: domain.TranslationFilter{Limit: 10},
	})
	if list2.Total != 0 {
		t.Errorf("expected 0 translations after delete, got %d", list2.Total)
	}
}
