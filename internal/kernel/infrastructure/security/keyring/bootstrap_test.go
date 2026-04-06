package keyring

import (
	"context"
	"errors"
	"testing"
)

func TestBootstrap_GeneratesForMissingKeyID(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	kr, err := New(dir, testBits)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	var called []string
	updateFn := func(_ context.Context, name, pubPEM, kid string) error {
		called = append(called, name)
		if pubPEM == "" {
			t.Errorf("expected non-empty publicKeyPEM for %s", name)
		}
		if kid == "" {
			t.Errorf("expected non-empty keyID for %s", name)
		}
		return nil
	}

	integrations := []BootstrapIntegration{
		{Name: "svc-alpha", KeyID: ""},
		{Name: "svc-beta", KeyID: ""},
	}

	l := &noopLog{}
	if err := Bootstrap(context.Background(), kr, integrations, updateFn, l); err != nil {
		t.Fatalf("Bootstrap: %v", err)
	}

	if len(called) != 2 {
		t.Fatalf("expected updateFn called 2 times, got %d", len(called))
	}
	if called[0] != "svc-alpha" || called[1] != "svc-beta" {
		t.Fatalf("unexpected call order: %v", called)
	}
}

func TestBootstrap_SkipsExistingMatchingKeyID(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	kr, err := New(dir, testBits)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Pre-generate a key pair.
	kp, err := kr.EnsureAndLoad("svc-existing", "")
	if err != nil {
		t.Fatalf("EnsureAndLoad: %v", err)
	}

	var called int
	updateFn := func(_ context.Context, _, _, _ string) error {
		called++
		return nil
	}

	integrations := []BootstrapIntegration{
		{Name: "svc-existing", KeyID: kp.KeyID},
	}

	l := &noopLog{}
	if err := Bootstrap(context.Background(), kr, integrations, updateFn, l); err != nil {
		t.Fatalf("Bootstrap: %v", err)
	}

	if called != 0 {
		t.Fatalf("expected updateFn not called for existing key, got %d calls", called)
	}
}

func TestBootstrap_ContinuesOnIndividualFailure(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	kr, err := New(dir, testBits)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	callCount := 0
	failOnFirst := func(_ context.Context, name, _, _ string) error {
		callCount++
		if name == "svc-fail" {
			return errors.New("simulated DB error")
		}
		return nil
	}

	integrations := []BootstrapIntegration{
		{Name: "svc-fail", KeyID: ""},
		{Name: "svc-ok", KeyID: ""},
	}

	l := &noopLog{}
	err = Bootstrap(context.Background(), kr, integrations, failOnFirst, l)
	// Should return the first error but still process svc-ok.
	if err == nil {
		t.Fatal("expected error from Bootstrap, got nil")
	}

	if callCount != 2 {
		t.Fatalf("expected updateFn called 2 times (both attempted), got %d", callCount)
	}
}

// noopLog satisfies logger.Log for tests without pulling in the full logger.
type noopLog struct{}

func (n *noopLog) Debug(_ ...any)                                    {}
func (n *noopLog) Info(_ ...any)                                     {}
func (n *noopLog) Warn(_ ...any)                                     {}
func (n *noopLog) Error(_ ...any)                                    {}
func (n *noopLog) Fatal(_ ...any)                                    {}
func (n *noopLog) Debugf(_ string, _ ...any)                         {}
func (n *noopLog) Infof(_ string, _ ...any)                          {}
func (n *noopLog) Warnf(_ string, _ ...any)                          {}
func (n *noopLog) Errorf(_ string, _ ...any)                         {}
func (n *noopLog) Fatalf(_ string, _ ...any)                         {}
func (n *noopLog) Debugw(_ string, _ ...any)                         {}
func (n *noopLog) Infow(_ string, _ ...any)                          {}
func (n *noopLog) Warnw(_ string, _ ...any)                          {}
func (n *noopLog) Errorw(_ string, _ ...any)                         {}
func (n *noopLog) Fatalw(_ string, _ ...any)                         {}
func (n *noopLog) Debugc(_ context.Context, _ string, _ ...any)      {}
func (n *noopLog) Infoc(_ context.Context, _ string, _ ...any)       {}
func (n *noopLog) Warnc(_ context.Context, _ string, _ ...any)       {}
func (n *noopLog) Errorc(_ context.Context, _ string, _ ...any)      {}
func (n *noopLog) Fatalc(_ context.Context, _ string, _ ...any)      {}
