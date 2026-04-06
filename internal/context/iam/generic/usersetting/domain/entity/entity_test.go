package entity_test

import (
	"testing"
	"time"

	"gct/internal/context/iam/generic/usersetting/domain/entity"

	"github.com/google/uuid"
)

func TestNewUserSetting(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	us := entity.NewUserSetting(userID, "theme", "dark")

	if us.UserID() != userID {
		t.Fatal("user ID mismatch")
	}
	if us.Key() != "theme" {
		t.Fatalf("expected key theme, got %s", us.Key())
	}
	if us.Value() != "dark" {
		t.Fatalf("expected value dark, got %s", us.Value())
	}

	events := us.Events()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].EventName() != "usersetting.changed" {
		t.Fatalf("expected usersetting.changed, got %s", events[0].EventName())
	}
}

func TestUserSetting_ChangeValue(t *testing.T) {
	t.Parallel()

	us := entity.NewUserSetting(uuid.New(), "language", "en")
	us.ChangeValue("fr")

	if us.Value() != "fr" {
		t.Fatalf("expected value fr, got %s", us.Value())
	}

	events := us.Events()
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events[1].EventName() != "usersetting.changed" {
		t.Fatalf("expected usersetting.changed, got %s", events[1].EventName())
	}
}

func TestReconstructUserSetting(t *testing.T) {
	t.Parallel()

	id := entity.NewUserSettingID()
	userID := uuid.New()
	now := time.Now()

	us := entity.ReconstructUserSetting(id.UUID(), now, now, userID, "timezone", "UTC")

	if us.TypedID() != id {
		t.Fatal("ID mismatch")
	}
	if us.UserID() != userID {
		t.Fatal("user ID mismatch")
	}
	if us.Key() != "timezone" {
		t.Fatal("key mismatch")
	}
	if us.Value() != "UTC" {
		t.Fatal("value mismatch")
	}
	if len(us.Events()) != 0 {
		t.Fatalf("expected 0 events, got %d", len(us.Events()))
	}
}
