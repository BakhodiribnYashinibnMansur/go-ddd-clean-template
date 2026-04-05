package domain_test

import (
	"testing"
	"time"

	domain "gct/internal/context/admin/supporting/sitesetting/domain"
)

func TestNewSiteSetting(t *testing.T) {
	t.Parallel()

	s := domain.NewSiteSetting("site_name", "My Site", "string", "The name of the site")

	if s.Key() != "site_name" {
		t.Fatalf("expected key site_name, got %s", s.Key())
	}
	if s.Value() != "My Site" {
		t.Fatalf("expected value My Site, got %s", s.Value())
	}
	if s.Type() != "string" {
		t.Fatalf("expected type string, got %s", s.Type())
	}
	if s.Description() != "The name of the site" {
		t.Fatalf("expected description, got %s", s.Description())
	}
}

func TestSiteSetting_Update(t *testing.T) {
	t.Parallel()

	s := domain.NewSiteSetting("site_name", "My Site", "string", "desc")

	newValue := "New Site Name"
	s.Update(nil, &newValue, nil, nil)

	if s.Value() != "New Site Name" {
		t.Fatalf("expected value New Site Name, got %s", s.Value())
	}

	events := s.Events()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].EventName() != "sitesetting.updated" {
		t.Fatalf("expected sitesetting.updated, got %s", events[0].EventName())
	}
}

func TestReconstructSiteSetting(t *testing.T) {
	t.Parallel()

	id := domain.NewSiteSettingID()
	now := time.Now()

	s := domain.ReconstructSiteSetting(id.UUID(), now, now, "key", "val", "string", "desc")

	if s.TypedID() != id {
		t.Fatal("ID mismatch")
	}
	if s.Key() != "key" {
		t.Fatal("key mismatch")
	}
	if len(s.Events()) != 0 {
		t.Fatalf("expected 0 events, got %d", len(s.Events()))
	}
}
