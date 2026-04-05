package domain

import (
	"testing"
	"time"

	shared "gct/internal/kernel/domain"
)

func newTestAnnouncement() *Announcement {
	title := shared.Lang{Uz: "Sarlavha", Ru: "Заголовок", En: "Title"}
	content := shared.Lang{Uz: "Matn", Ru: "Содержание", En: "Content"}
	a, _ := NewAnnouncement(title, content, 1, nil, nil)
	return a
}

func TestInvariant_NewAnnouncement_NotPublished(t *testing.T) {
	a := newTestAnnouncement()
	if a.Published() {
		t.Fatal("expected new announcement to be unpublished")
	}
}

func TestInvariant_NewAnnouncement_HasNilPublishedAt(t *testing.T) {
	a := newTestAnnouncement()
	if a.PublishedAt() != nil {
		t.Fatalf("expected nil PublishedAt on new announcement, got %v", a.PublishedAt())
	}
}

func TestInvariant_Publish_SetsPublished(t *testing.T) {
	a := newTestAnnouncement()
	a.Publish()
	if !a.Published() {
		t.Fatal("expected Published() to be true after Publish()")
	}
}

func TestInvariant_Publish_SetsPublishedAt(t *testing.T) {
	a := newTestAnnouncement()
	before := time.Now()
	a.Publish()
	after := time.Now()

	if a.PublishedAt() == nil {
		t.Fatal("expected non-nil PublishedAt after Publish()")
	}
	pub := *a.PublishedAt()
	if pub.Before(before) || pub.After(after) {
		t.Errorf("expected PublishedAt between %v and %v, got %v", before, after, pub)
	}
}

func TestInvariant_Publish_RaisesEvent(t *testing.T) {
	a := newTestAnnouncement()
	a.Publish()

	events := a.Events()
	if len(events) != 1 {
		t.Fatalf("expected 1 event after Publish, got %d", len(events))
	}
	evt, ok := events[0].(AnnouncementPublished)
	if !ok {
		t.Fatalf("expected AnnouncementPublished event, got %T", events[0])
	}
	if evt.EventName() != "announcement.published" {
		t.Errorf("expected event name %q, got %q", "announcement.published", evt.EventName())
	}
	if evt.AggregateID() != a.ID() {
		t.Errorf("expected aggregate ID %v, got %v", a.ID(), evt.AggregateID())
	}
}

func TestInvariant_Update_PartialFields(t *testing.T) {
	a := newTestAnnouncement()

	origTitle := a.Title()
	origContent := a.Content()
	origPriority := a.Priority()

	// Update only priority; title and content should remain unchanged.
	newPriority := 5
	a.Update(nil, nil, &newPriority, nil, nil)

	if a.Title() != origTitle {
		t.Errorf("expected title unchanged %v, got %v", origTitle, a.Title())
	}
	if a.Content() != origContent {
		t.Errorf("expected content unchanged %v, got %v", origContent, a.Content())
	}
	if a.Priority() != newPriority {
		t.Errorf("expected priority %d, got %d", newPriority, a.Priority())
	}
	if a.Priority() == origPriority {
		t.Error("priority should have changed from original")
	}

	// Update only title; priority should remain at the new value.
	newTitle := shared.Lang{Uz: "Yangi", Ru: "Новый", En: "New"}
	a.Update(&newTitle, nil, nil, nil, nil)

	if a.Title() != newTitle {
		t.Errorf("expected title %v, got %v", newTitle, a.Title())
	}
	if a.Priority() != newPriority {
		t.Errorf("expected priority still %d, got %d", newPriority, a.Priority())
	}
}
