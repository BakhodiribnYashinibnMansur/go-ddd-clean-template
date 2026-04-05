package domain_test

import (
	"testing"
	"time"

	domain "gct/internal/context/content/generic/translation/domain"

	"github.com/google/uuid"
)

func TestNewTranslation(t *testing.T) {
	t.Parallel()

	tr := domain.NewTranslation("greeting", "en", "Hello", "common")

	if tr.Key() != "greeting" {
		t.Fatalf("expected key greeting, got %s", tr.Key())
	}
	if tr.Language() != "en" {
		t.Fatalf("expected language en, got %s", tr.Language())
	}
	if tr.Value() != "Hello" {
		t.Fatalf("expected value Hello, got %s", tr.Value())
	}
	if tr.Group() != "common" {
		t.Fatalf("expected group common, got %s", tr.Group())
	}
}

func TestTranslation_Update(t *testing.T) {
	t.Parallel()

	tr := domain.NewTranslation("greeting", "en", "Hello", "common")

	newValue := "Hi there"
	tr.Update(nil, nil, &newValue, nil)

	if tr.Value() != "Hi there" {
		t.Fatalf("expected value Hi there, got %s", tr.Value())
	}

	events := tr.Events()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].EventName() != "translation.updated" {
		t.Fatalf("expected translation.updated, got %s", events[0].EventName())
	}
}

func TestReconstructTranslation(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	now := time.Now()

	tr := domain.ReconstructTranslation(id, now, now, "key1", "uz", "Salom", "greetings")

	if tr.ID() != id {
		t.Fatal("ID mismatch")
	}
	if tr.Key() != "key1" {
		t.Fatal("key mismatch")
	}
	if len(tr.Events()) != 0 {
		t.Fatalf("expected 0 events, got %d", len(tr.Events()))
	}
}
