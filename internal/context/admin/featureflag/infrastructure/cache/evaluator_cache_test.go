package cache

import (
	"context"
	"errors"
	"testing"
	"time"

	"gct/internal/context/admin/featureflag/domain"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Mock Logger
// ---------------------------------------------------------------------------

type mockLog struct {
	infowCalls  []string
	errorwCalls []string
	debugwCalls []string
}

func (m *mockLog) Debug(_ ...any)                                {}
func (m *mockLog) Debugf(_ string, _ ...any)                     {}
func (m *mockLog) Debugw(msg string, _ ...any)                   { m.debugwCalls = append(m.debugwCalls, msg) }
func (m *mockLog) Info(_ ...any)                                 {}
func (m *mockLog) Infof(_ string, _ ...any)                      {}
func (m *mockLog) Infow(msg string, _ ...any)                    { m.infowCalls = append(m.infowCalls, msg) }
func (m *mockLog) Warn(_ ...any)                                 {}
func (m *mockLog) Warnf(_ string, _ ...any)                      {}
func (m *mockLog) Warnw(_ string, _ ...any)                      {}
func (m *mockLog) Error(_ ...any)                                {}
func (m *mockLog) Errorf(_ string, _ ...any)                     {}
func (m *mockLog) Errorw(msg string, _ ...any)                   { m.errorwCalls = append(m.errorwCalls, msg) }
func (m *mockLog) Fatal(_ ...any)                                {}
func (m *mockLog) Fatalf(_ string, _ ...any)                     {}
func (m *mockLog) Fatalw(_ string, _ ...any)                     {}
func (m *mockLog) Debugc(_ context.Context, _ string, _ ...any)  {}
func (m *mockLog) Infoc(_ context.Context, _ string, _ ...any)   {}
func (m *mockLog) Warnc(_ context.Context, _ string, _ ...any)   {}
func (m *mockLog) Errorc(_ context.Context, _ string, _ ...any)  {}
func (m *mockLog) Fatalc(_ context.Context, _ string, _ ...any)  {}

// ---------------------------------------------------------------------------
// Mock Repository
// ---------------------------------------------------------------------------

type mockRepo struct {
	flags      []*domain.FeatureFlag
	findAllErr error
	byKey      map[string]*domain.FeatureFlag
	findKeyErr error
}

func (r *mockRepo) Save(_ context.Context, _ *domain.FeatureFlag) error   { return nil }
func (r *mockRepo) Update(_ context.Context, _ *domain.FeatureFlag) error { return nil }
func (r *mockRepo) Delete(_ context.Context, _ uuid.UUID) error           { return nil }

func (r *mockRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.FeatureFlag, error) {
	for _, f := range r.flags {
		if f.ID() == id {
			return f, nil
		}
	}
	return nil, errors.New("not found")
}

func (r *mockRepo) FindByKey(_ context.Context, key string) (*domain.FeatureFlag, error) {
	if r.findKeyErr != nil {
		return nil, r.findKeyErr
	}
	if ff, ok := r.byKey[key]; ok {
		return ff, nil
	}
	return nil, errors.New("not found")
}

func (r *mockRepo) FindAll(_ context.Context) ([]*domain.FeatureFlag, error) {
	if r.findAllErr != nil {
		return nil, r.findAllErr
	}
	return r.flags, nil
}

// ---------------------------------------------------------------------------
// Helper: create a simple bool feature flag
// ---------------------------------------------------------------------------

func newTestFlag(key string, active bool, defaultVal string) *domain.FeatureFlag {
	ff := domain.ReconstructFeatureFlag(
		uuid.New(),
		time.Now(), time.Now(), nil,
		"Test Flag", key, "desc", "bool", defaultVal,
		0, active, nil,
	)
	return ff
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestNewCachedEvaluator_Success(t *testing.T) {
	repo := &mockRepo{
		flags: []*domain.FeatureFlag{
			newTestFlag("feat-a", true, "false"),
		},
	}
	log := &mockLog{}

	ce, err := NewCachedEvaluator(context.Background(), repo, log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ce == nil {
		t.Fatal("expected non-nil CachedEvaluator")
	}
	if len(log.infowCalls) == 0 {
		t.Error("expected at least one infow call for cache load")
	}
}

func TestNewCachedEvaluator_RepoError(t *testing.T) {
	repo := &mockRepo{findAllErr: errors.New("db down")}
	log := &mockLog{}

	ce, err := NewCachedEvaluator(context.Background(), repo, log)
	if err == nil {
		t.Fatal("expected error")
	}
	if ce != nil {
		t.Error("expected nil CachedEvaluator on error")
	}
}

func TestIsEnabled_ActiveFlag(t *testing.T) {
	ff := newTestFlag("dark-mode", true, "true")
	repo := &mockRepo{flags: []*domain.FeatureFlag{ff}}
	log := &mockLog{}

	ce, err := NewCachedEvaluator(context.Background(), repo, log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Active flag with default "true" should return true
	got := ce.IsEnabled(context.Background(), "dark-mode", nil)
	if !got {
		t.Error("expected IsEnabled to return true for active flag with default 'true'")
	}
}

func TestIsEnabled_InactiveFlag(t *testing.T) {
	ff := newTestFlag("dark-mode", false, "false")
	repo := &mockRepo{flags: []*domain.FeatureFlag{ff}}
	log := &mockLog{}

	ce, err := NewCachedEvaluator(context.Background(), repo, log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := ce.IsEnabled(context.Background(), "dark-mode", nil)
	if got {
		t.Error("expected IsEnabled to return false for inactive flag")
	}
}

func TestIsEnabled_MissingFlag(t *testing.T) {
	repo := &mockRepo{
		flags:      nil,
		findKeyErr: errors.New("not found"),
	}
	log := &mockLog{}

	ce, err := NewCachedEvaluator(context.Background(), repo, log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := ce.IsEnabled(context.Background(), "nonexistent", nil)
	if got {
		t.Error("expected false for missing flag")
	}
}

func TestGetString_CachedFlag(t *testing.T) {
	ff := newTestFlag("color", true, "blue")
	repo := &mockRepo{flags: []*domain.FeatureFlag{ff}}
	log := &mockLog{}

	ce, err := NewCachedEvaluator(context.Background(), repo, log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := ce.GetString(context.Background(), "color", nil)
	if got != "blue" {
		t.Errorf("expected 'blue', got %q", got)
	}
}

func TestGetString_MissingFlag(t *testing.T) {
	repo := &mockRepo{findKeyErr: errors.New("not found")}
	log := &mockLog{}

	ce, err := NewCachedEvaluator(context.Background(), repo, log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := ce.GetString(context.Background(), "missing", nil)
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestGetInt_ValidInt(t *testing.T) {
	ff := newTestFlag("max-items", true, "42")
	repo := &mockRepo{flags: []*domain.FeatureFlag{ff}}
	log := &mockLog{}

	ce, err := NewCachedEvaluator(context.Background(), repo, log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := ce.GetInt(context.Background(), "max-items", nil)
	if got != 42 {
		t.Errorf("expected 42, got %d", got)
	}
}

func TestGetInt_InvalidInt(t *testing.T) {
	ff := newTestFlag("max-items", true, "not-a-number")
	repo := &mockRepo{flags: []*domain.FeatureFlag{ff}}
	log := &mockLog{}

	ce, err := NewCachedEvaluator(context.Background(), repo, log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := ce.GetInt(context.Background(), "max-items", nil)
	if got != 0 {
		t.Errorf("expected 0 for invalid int, got %d", got)
	}
}

func TestGetInt_MissingFlag(t *testing.T) {
	repo := &mockRepo{findKeyErr: errors.New("not found")}
	log := &mockLog{}

	ce, err := NewCachedEvaluator(context.Background(), repo, log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := ce.GetInt(context.Background(), "missing", nil)
	if got != 0 {
		t.Errorf("expected 0 for missing flag, got %d", got)
	}
}

func TestGetFloat_ValidFloat(t *testing.T) {
	ff := newTestFlag("threshold", true, "3.14")
	repo := &mockRepo{flags: []*domain.FeatureFlag{ff}}
	log := &mockLog{}

	ce, err := NewCachedEvaluator(context.Background(), repo, log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := ce.GetFloat(context.Background(), "threshold", nil)
	if got != 3.14 {
		t.Errorf("expected 3.14, got %f", got)
	}
}

func TestGetFloat_InvalidFloat(t *testing.T) {
	ff := newTestFlag("threshold", true, "abc")
	repo := &mockRepo{flags: []*domain.FeatureFlag{ff}}
	log := &mockLog{}

	ce, err := NewCachedEvaluator(context.Background(), repo, log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := ce.GetFloat(context.Background(), "threshold", nil)
	if got != 0 {
		t.Errorf("expected 0 for invalid float, got %f", got)
	}
}

func TestGetFloat_MissingFlag(t *testing.T) {
	repo := &mockRepo{findKeyErr: errors.New("not found")}
	log := &mockLog{}

	ce, err := NewCachedEvaluator(context.Background(), repo, log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := ce.GetFloat(context.Background(), "missing", nil)
	if got != 0 {
		t.Errorf("expected 0 for missing flag, got %f", got)
	}
}

func TestGetFlag_FallsBackToRepo(t *testing.T) {
	// Flag not in initial FindAll, but available via FindByKey
	ff := newTestFlag("lazy-flag", true, "hello")
	repo := &mockRepo{
		flags: nil, // empty initial load
		byKey: map[string]*domain.FeatureFlag{"lazy-flag": ff},
	}
	log := &mockLog{}

	ce, err := NewCachedEvaluator(context.Background(), repo, log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := ce.GetString(context.Background(), "lazy-flag", nil)
	if got != "hello" {
		t.Errorf("expected 'hello', got %q", got)
	}

	// Second call should use the cache (not repo again)
	got2 := ce.GetString(context.Background(), "lazy-flag", nil)
	if got2 != "hello" {
		t.Errorf("expected 'hello' on second call, got %q", got2)
	}
}

func TestInvalidate_ReloadsCache(t *testing.T) {
	ff1 := newTestFlag("feat-a", true, "v1")
	repo := &mockRepo{flags: []*domain.FeatureFlag{ff1}}
	log := &mockLog{}

	ce, err := NewCachedEvaluator(context.Background(), repo, log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify initial value
	got := ce.GetString(context.Background(), "feat-a", nil)
	if got != "v1" {
		t.Errorf("expected 'v1', got %q", got)
	}

	// Update repo and invalidate
	ff2 := newTestFlag("feat-a", true, "v2")
	repo.flags = []*domain.FeatureFlag{ff2}
	ce.Invalidate(context.Background())

	got = ce.GetString(context.Background(), "feat-a", nil)
	if got != "v2" {
		t.Errorf("expected 'v2' after invalidation, got %q", got)
	}
}

func TestInvalidate_RepoError(t *testing.T) {
	ff := newTestFlag("feat-a", true, "v1")
	repo := &mockRepo{flags: []*domain.FeatureFlag{ff}}
	log := &mockLog{}

	ce, err := NewCachedEvaluator(context.Background(), repo, log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Make repo fail on next FindAll
	repo.findAllErr = errors.New("db gone")
	ce.Invalidate(context.Background())

	if len(log.errorwCalls) == 0 {
		t.Error("expected error log when invalidation fails")
	}
}

func TestLoadAll_ClearsPreviousCache(t *testing.T) {
	ff1 := newTestFlag("old-flag", true, "old")
	repo := &mockRepo{flags: []*domain.FeatureFlag{ff1}}
	log := &mockLog{}

	ce, err := NewCachedEvaluator(context.Background(), repo, log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Replace with new flag set (no old-flag)
	ff2 := newTestFlag("new-flag", true, "new")
	repo.flags = []*domain.FeatureFlag{ff2}
	repo.findKeyErr = errors.New("not found") // prevent fallback

	if err := ce.LoadAll(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// old-flag should no longer be cached
	got := ce.GetString(context.Background(), "old-flag", nil)
	if got != "" {
		t.Errorf("expected empty string for cleared flag, got %q", got)
	}

	// new-flag should be present
	got = ce.GetString(context.Background(), "new-flag", nil)
	if got != "new" {
		t.Errorf("expected 'new', got %q", got)
	}
}
