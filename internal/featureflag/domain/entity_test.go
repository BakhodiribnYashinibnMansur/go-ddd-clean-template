package domain_test

import (
	"testing"

	"gct/internal/featureflag/domain"
)

func TestNewFeatureFlag(t *testing.T) {
	ff := domain.NewFeatureFlag("dark-mode", "Enable dark mode", false, 50)

	if ff.Name() != "dark-mode" {
		t.Fatalf("expected name dark-mode, got %s", ff.Name())
	}
	if ff.Description() != "Enable dark mode" {
		t.Fatalf("expected description 'Enable dark mode', got %s", ff.Description())
	}
	if ff.Enabled() {
		t.Fatal("expected enabled false")
	}
	if ff.RolloutPercentage() != 50 {
		t.Fatalf("expected rollout 50, got %d", ff.RolloutPercentage())
	}
	if ff.ID().String() == "" {
		t.Fatal("expected non-empty ID")
	}
}

func TestFeatureFlag_Toggle(t *testing.T) {
	ff := domain.NewFeatureFlag("test", "desc", false, 100)

	ff.Toggle()
	if !ff.Enabled() {
		t.Fatal("expected enabled true after toggle")
	}
	if len(ff.Events()) != 1 {
		t.Fatalf("expected 1 event, got %d", len(ff.Events()))
	}
	if ff.Events()[0].EventName() != "featureflag.toggled" {
		t.Fatalf("expected event featureflag.toggled, got %s", ff.Events()[0].EventName())
	}

	ff.Toggle()
	if ff.Enabled() {
		t.Fatal("expected enabled false after second toggle")
	}
}
