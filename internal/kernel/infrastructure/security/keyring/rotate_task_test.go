package keyring

import (
	"context"
	"testing"
	"time"

	hibikenasynq "github.com/hibiken/asynq"
)

// stubLister implements IntegrationLister for tests.
type stubLister struct {
	views []JWTIntegrationView
	err   error
}

func (s *stubLister) ListActiveJWT(_ context.Context) ([]JWTIntegrationView, error) {
	return s.views, s.err
}

func TestRotateHandler_RotatesDueIntegrations(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	kr, err := New(dir, testBits)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Pre-generate a key so Rotate has something to rotate from.
	orig, err := kr.EnsureAndLoad("svc-due", "kid-old")
	if err != nil {
		t.Fatalf("EnsureAndLoad: %v", err)
	}

	rotatedAt := time.Now().Add(-8 * 24 * time.Hour) // 8 days ago
	lister := &stubLister{
		views: []JWTIntegrationView{
			{
				Name:            "svc-due",
				KeyID:           "kid-old",
				RotatedAt:       &rotatedAt,
				RotateEveryDays: 7,
			},
		},
	}

	var updatedName, updatedKID string
	updateFn := func(_ context.Context, name, pubPEM, kid string) error {
		updatedName = name
		updatedKID = kid
		if pubPEM == "" {
			t.Error("expected non-empty publicKeyPEM")
		}
		return nil
	}

	l := &noopLog{}
	handler := NewRotateKeysHandler(kr, lister, updateFn, l)

	task, err := NewRotateKeysTask()
	if err != nil {
		t.Fatalf("NewRotateKeysTask: %v", err)
	}
	if err := handler.Handle(context.Background(), task); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	if updatedName != "svc-due" {
		t.Fatalf("expected update for svc-due, got %q", updatedName)
	}
	if updatedKID == "" || updatedKID == orig.KeyID {
		t.Fatalf("expected new keyID, got %q (original was %q)", updatedKID, orig.KeyID)
	}
}

func TestRotateHandler_SkipsRecentRotation(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	kr, err := New(dir, testBits)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, err := kr.EnsureAndLoad("svc-recent", "kid-recent"); err != nil {
		t.Fatalf("EnsureAndLoad: %v", err)
	}

	rotatedAt := time.Now().Add(-1 * 24 * time.Hour) // 1 day ago
	lister := &stubLister{
		views: []JWTIntegrationView{
			{
				Name:            "svc-recent",
				KeyID:           "kid-recent",
				RotatedAt:       &rotatedAt,
				RotateEveryDays: 7,
			},
		},
	}

	called := false
	updateFn := func(_ context.Context, _, _, _ string) error {
		called = true
		return nil
	}

	l := &noopLog{}
	handler := NewRotateKeysHandler(kr, lister, updateFn, l)

	task := hibikenasynq.NewTask(TaskTypeRotateKeys, nil)
	if err := handler.Handle(context.Background(), task); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	if called {
		t.Fatal("expected updateFn NOT called for recent rotation")
	}
}

func TestRotateHandler_SkipsNilRotatedAt(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	kr, err := New(dir, testBits)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, err := kr.EnsureAndLoad("svc-new", "kid-new"); err != nil {
		t.Fatalf("EnsureAndLoad: %v", err)
	}

	lister := &stubLister{
		views: []JWTIntegrationView{
			{
				Name:            "svc-new",
				KeyID:           "kid-new",
				RotatedAt:       nil, // just generated
				RotateEveryDays: 7,
			},
		},
	}

	called := false
	updateFn := func(_ context.Context, _, _, _ string) error {
		called = true
		return nil
	}

	l := &noopLog{}
	handler := NewRotateKeysHandler(kr, lister, updateFn, l)

	task := hibikenasynq.NewTask(TaskTypeRotateKeys, nil)
	if err := handler.Handle(context.Background(), task); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	if called {
		t.Fatal("expected updateFn NOT called for nil RotatedAt")
	}
}

func TestRotateHandler_SkipsZeroRotateEveryDays(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	kr, err := New(dir, testBits)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, err := kr.EnsureAndLoad("svc-disabled", "kid-disabled"); err != nil {
		t.Fatalf("EnsureAndLoad: %v", err)
	}

	rotatedAt := time.Now().Add(-30 * 24 * time.Hour)
	lister := &stubLister{
		views: []JWTIntegrationView{
			{
				Name:            "svc-disabled",
				KeyID:           "kid-disabled",
				RotatedAt:       &rotatedAt,
				RotateEveryDays: 0, // rotation disabled
			},
		},
	}

	called := false
	updateFn := func(_ context.Context, _, _, _ string) error {
		called = true
		return nil
	}

	l := &noopLog{}
	handler := NewRotateKeysHandler(kr, lister, updateFn, l)

	task := hibikenasynq.NewTask(TaskTypeRotateKeys, nil)
	if err := handler.Handle(context.Background(), task); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	if called {
		t.Fatal("expected updateFn NOT called when RotateEveryDays is 0")
	}
}
