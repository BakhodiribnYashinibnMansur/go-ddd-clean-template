package announcement

import (
	"context"
	"testing"

	"gct/internal/context/content/announcement"
	"gct/internal/context/content/announcement/application/command"
	"gct/internal/context/content/announcement/application/query"
	"gct/internal/context/content/announcement/domain"
	shared "gct/internal/kernel/domain"
	"gct/internal/kernel/infrastructure/eventbus"
	"gct/internal/kernel/infrastructure/logger"
	"gct/test/integration/common/setup"
)

func newTestBC(t *testing.T) *announcement.BoundedContext {
	t.Helper()
	eb := eventbus.NewInMemoryEventBus()
	l := logger.New("error")
	return announcement.NewBoundedContext(setup.TestPG.Pool, eb, l)
}

func TestIntegration_CreateAndGetAnnouncement(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	err := bc.CreateAnnouncement.Handle(ctx, command.CreateAnnouncementCommand{
		Title:    shared.Lang{Uz: "Test", Ru: "Тест", En: "Test"},
		Content:  shared.Lang{Uz: "Mazmun", Ru: "Содержание", En: "Content"},
		Priority: 1,
	})
	if err != nil {
		t.Fatalf("CreateAnnouncement: %v", err)
	}

	result, err := bc.ListAnnouncements.Handle(ctx, query.ListAnnouncementsQuery{
		Filter: domain.AnnouncementFilter{Limit: 10},
	})
	if err != nil {
		t.Fatalf("ListAnnouncements: %v", err)
	}
	if result.Total != 1 {
		t.Fatalf("expected 1 announcement, got %d", result.Total)
	}

	a := result.Announcements[0]
	if a.Title.En != "Test" {
		t.Errorf("expected title_en Test, got %s", a.Title.En)
	}
	if a.Priority != 1 {
		t.Errorf("expected priority 1, got %d", a.Priority)
	}

	view, err := bc.GetAnnouncement.Handle(ctx, query.GetAnnouncementQuery{ID: domain.AnnouncementID(a.ID)})
	if err != nil {
		t.Fatalf("GetAnnouncement: %v", err)
	}
	if view.ID != a.ID {
		t.Errorf("ID mismatch: %s vs %s", view.ID, a.ID)
	}
}

func TestIntegration_UpdateAndPublishAnnouncement(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	err := bc.CreateAnnouncement.Handle(ctx, command.CreateAnnouncementCommand{
		Title:    shared.Lang{Uz: "Draft", Ru: "Черновик", En: "Draft"},
		Content:  shared.Lang{Uz: "Mazmun", Ru: "Содержание", En: "Content"},
		Priority: 2,
	})
	if err != nil {
		t.Fatalf("CreateAnnouncement: %v", err)
	}

	list, _ := bc.ListAnnouncements.Handle(ctx, query.ListAnnouncementsQuery{
		Filter: domain.AnnouncementFilter{Limit: 10},
	})
	aID := list.Announcements[0].ID

	if list.Announcements[0].Published {
		t.Fatal("new announcement should not be published")
	}

	newTitle := shared.Lang{Uz: "Yangilangan", Ru: "Обновлено", En: "Updated"}
	newPriority := 5
	err = bc.UpdateAnnouncement.Handle(ctx, command.UpdateAnnouncementCommand{
		ID:       domain.AnnouncementID(aID),
		Title:    &newTitle,
		Priority: &newPriority,
		Publish:  true,
	})
	if err != nil {
		t.Fatalf("UpdateAnnouncement: %v", err)
	}

	view, _ := bc.GetAnnouncement.Handle(ctx, query.GetAnnouncementQuery{ID: domain.AnnouncementID(aID)})
	if view.Title.Uz != "Yangilangan" {
		t.Errorf("title not updated, got %s", view.Title.Uz)
	}
	if view.Priority != 5 {
		t.Errorf("priority not updated, got %d", view.Priority)
	}
	if !view.Published {
		t.Error("announcement should be published")
	}
}

func TestIntegration_DeleteAnnouncement(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	err := bc.CreateAnnouncement.Handle(ctx, command.CreateAnnouncementCommand{
		Title:    shared.Lang{Uz: "O'chirish", Ru: "Удалить", En: "Delete"},
		Content:  shared.Lang{Uz: "Mazmun", Ru: "Содержание", En: "Content"},
		Priority: 1,
	})
	if err != nil {
		t.Fatalf("CreateAnnouncement: %v", err)
	}

	list, _ := bc.ListAnnouncements.Handle(ctx, query.ListAnnouncementsQuery{
		Filter: domain.AnnouncementFilter{Limit: 10},
	})
	aID := list.Announcements[0].ID

	err = bc.DeleteAnnouncement.Handle(ctx, command.DeleteAnnouncementCommand{ID: domain.AnnouncementID(aID)})
	if err != nil {
		t.Fatalf("DeleteAnnouncement: %v", err)
	}

	list2, _ := bc.ListAnnouncements.Handle(ctx, query.ListAnnouncementsQuery{
		Filter: domain.AnnouncementFilter{Limit: 10},
	})
	if list2.Total != 0 {
		t.Errorf("expected 0 announcements after delete, got %d", list2.Total)
	}
}
