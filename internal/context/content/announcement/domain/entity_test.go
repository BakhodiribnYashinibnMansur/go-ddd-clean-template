package domain_test

import (
	"testing"
	"time"

	domain "gct/internal/context/content/announcement/domain"
	shared "gct/internal/kernel/domain"

	"github.com/google/uuid"
)

func TestNewAnnouncement(t *testing.T) {
	t.Parallel()

	title := shared.Lang{Uz: "Sarlavha", Ru: "Заголовок", En: "Title"}
	content := shared.Lang{Uz: "Mazmun", Ru: "Содержание", En: "Content"}
	a, _ := domain.NewAnnouncement(title, content, 1, nil, nil)

	if a.Title().Uz != "Sarlavha" {
		t.Fatalf("expected title uz Sarlavha, got %s", a.Title().Uz)
	}
	if a.Content().En != "Content" {
		t.Fatalf("expected content en Content, got %s", a.Content().En)
	}
	if a.Published() {
		t.Fatal("new announcement should not be published")
	}
	if a.PublishedAt() != nil {
		t.Fatal("publishedAt should be nil")
	}
	if a.Priority() != 1 {
		t.Fatalf("expected priority 1, got %d", a.Priority())
	}
}

func TestAnnouncement_Publish(t *testing.T) {
	t.Parallel()

	title := shared.Lang{Uz: "T", Ru: "T", En: "T"}
	content := shared.Lang{Uz: "C", Ru: "C", En: "C"}
	a, _ := domain.NewAnnouncement(title, content, 0, nil, nil)

	a.Publish()

	if !a.Published() {
		t.Fatal("should be published after Publish()")
	}
	if a.PublishedAt() == nil {
		t.Fatal("publishedAt should be set")
	}

	events := a.Events()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].EventName() != "announcement.published" {
		t.Fatalf("expected announcement.published, got %s", events[0].EventName())
	}
}

func TestReconstructAnnouncement(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	now := time.Now()
	title := shared.Lang{Uz: "T", Ru: "T", En: "T"}
	content := shared.Lang{Uz: "C", Ru: "C", En: "C"}

	a := domain.ReconstructAnnouncement(id, now, now, title, content, true, &now, 5, nil, nil)

	if a.ID() != id {
		t.Fatal("ID mismatch")
	}
	if !a.Published() {
		t.Fatal("should be published")
	}
	if a.Priority() != 5 {
		t.Fatal("priority mismatch")
	}
	if len(a.Events()) != 0 {
		t.Fatalf("expected 0 events, got %d", len(a.Events()))
	}
}
