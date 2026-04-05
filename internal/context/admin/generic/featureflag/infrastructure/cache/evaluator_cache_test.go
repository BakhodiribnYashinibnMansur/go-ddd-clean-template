package cache

import (
	"context"
	"errors"
	"testing"
	"time"

	"gct/internal/context/admin/generic/featureflag/domain"
	"gct/internal/kernel/infrastructure/logger"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Mock repository
// ---------------------------------------------------------------------------

type mockRepo struct {
	flags    map[string]*domain.FeatureFlag
	allFlags []*domain.FeatureFlag
	findErr  error
	allErr   error
}

func (m *mockRepo) FindByKey(_ context.Context, key string) (*domain.FeatureFlag, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	ff, ok := m.flags[key]
	if !ok {
		return nil, errors.New("not found")
	}
	return ff, nil
}

func (m *mockRepo) FindAll(context.Context) ([]*domain.FeatureFlag, error) {
	if m.allErr != nil {
		return nil, m.allErr
	}
	return m.allFlags, nil
}

func (m *mockRepo) Save(context.Context, *domain.FeatureFlag) error {
	panic("not implemented")
}
func (m *mockRepo) FindByID(context.Context, uuid.UUID) (*domain.FeatureFlag, error) {
	panic("not implemented")
}
func (m *mockRepo) Update(context.Context, *domain.FeatureFlag) error {
	panic("not implemented")
}
func (m *mockRepo) Delete(context.Context, uuid.UUID) error {
	panic("not implemented")
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func newFlag(key, flagType, defaultValue string, active bool) *domain.FeatureFlag {
	now := time.Now()
	return domain.ReconstructFeatureFlag(
		uuid.New(), now, now, nil,
		"Test Flag", key, "test description", flagType, defaultValue,
		0, active, nil,
	)
}

func buildRepo(flags ...*domain.FeatureFlag) *mockRepo {
	m := &mockRepo{
		flags:    make(map[string]*domain.FeatureFlag),
		allFlags: flags,
	}
	for _, ff := range flags {
		m.flags[ff.Key()] = ff
	}
	return m
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestNewCachedEvaluator_LoadsAllFlags(t *testing.T) {
	repo := buildRepo(
		newFlag("a", "bool", "true", true),
		newFlag("b", "string", "hello", true),
	)
	ce, err := NewCachedEvaluator(context.Background(), repo, logger.Noop())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, key := range []string{"a", "b"} {
		if _, ok := ce.cache.Load(key); !ok {
			t.Errorf("flag %q not found in cache after creation", key)
		}
	}
}

func TestNewCachedEvaluator_ErrorOnFindAllFailure(t *testing.T) {
	repo := &mockRepo{allErr: errors.New("db down")}
	ce, err := NewCachedEvaluator(context.Background(), repo, logger.Noop())
	if err == nil {
		t.Fatal("expected error when FindAll fails, got nil")
	}
	if ce != nil {
		t.Error("expected nil CachedEvaluator on error")
	}
}

func TestCachedEvaluator_IsEnabled(t *testing.T) {
	tests := []struct {
		name string
		repo *mockRepo
		key  string
		want bool
	}{
		{
			name: "from cache - active bool flag",
			repo: buildRepo(newFlag("feat.on", "bool", "true", true)),
			key:  "feat.on",
			want: true,
		},
		{
			name: "returns false for non-existent flag",
			repo: buildRepo(),
			key:  "missing",
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ce, err := NewCachedEvaluator(context.Background(), tc.repo, logger.Noop())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := ce.IsEnabled(context.Background(), tc.key, nil)
			if got != tc.want {
				t.Errorf("IsEnabled(%q) = %v, want %v", tc.key, got, tc.want)
			}
		})
	}
}

func TestCachedEvaluator_IsEnabled_FallbackToRepo(t *testing.T) {
	extra := newFlag("extra", "bool", "true", true)
	repo := &mockRepo{
		flags:    map[string]*domain.FeatureFlag{"extra": extra},
		allFlags: nil,
	}

	ce, err := NewCachedEvaluator(context.Background(), repo, logger.Noop())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := ce.IsEnabled(context.Background(), "extra", nil)
	if !got {
		t.Error("IsEnabled should return true via repo fallback")
	}

	if _, ok := ce.cache.Load("extra"); !ok {
		t.Error("flag should be cached after repo fallback")
	}
}

func TestCachedEvaluator_GetString(t *testing.T) {
	repo := buildRepo(newFlag("color", "string", "blue", true))
	ce, err := NewCachedEvaluator(context.Background(), repo, logger.Noop())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := ce.GetString(context.Background(), "color", nil)
	if got != "blue" {
		t.Errorf("GetString = %q, want %q", got, "blue")
	}
}

func TestCachedEvaluator_GetInt(t *testing.T) {
	repo := buildRepo(newFlag("limit", "int", "42", true))
	ce, err := NewCachedEvaluator(context.Background(), repo, logger.Noop())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := ce.GetInt(context.Background(), "limit", nil)
	if got != 42 {
		t.Errorf("GetInt = %d, want %d", got, 42)
	}
}

func TestCachedEvaluator_GetFloat(t *testing.T) {
	repo := buildRepo(newFlag("rate", "float", "3.14", true))
	ce, err := NewCachedEvaluator(context.Background(), repo, logger.Noop())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := ce.GetFloat(context.Background(), "rate", nil)
	if got != 3.14 {
		t.Errorf("GetFloat = %f, want %f", got, 3.14)
	}
}

func TestCachedEvaluator_EvaluateFull(t *testing.T) {
	tests := []struct {
		name     string
		repo     *mockRepo
		key      string
		wantNil  bool
		wantVal  string
		wantType string
	}{
		{
			name:     "returns result with flag type",
			repo:     buildRepo(newFlag("color", "string", "red", true)),
			key:      "color",
			wantNil:  false,
			wantVal:  "red",
			wantType: "string",
		},
		{
			name:    "returns nil for non-existent flag",
			repo:    buildRepo(),
			key:     "missing",
			wantNil: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ce, err := NewCachedEvaluator(context.Background(), tc.repo, logger.Noop())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := ce.EvaluateFull(context.Background(), tc.key, nil)
			if tc.wantNil {
				if got != nil {
					t.Errorf("EvaluateFull(%q) = %+v, want nil", tc.key, got)
				}
				return
			}
			if got == nil {
				t.Fatalf("EvaluateFull(%q) = nil, want non-nil", tc.key)
			}
			if got.Value != tc.wantVal {
				t.Errorf("Value = %q, want %q", got.Value, tc.wantVal)
			}
			if got.FlagType != tc.wantType {
				t.Errorf("FlagType = %q, want %q", got.FlagType, tc.wantType)
			}
		})
	}
}

func TestCachedEvaluator_Invalidate(t *testing.T) {
	flag1 := newFlag("a", "string", "v1", true)
	repo := buildRepo(flag1)

	ce, err := NewCachedEvaluator(context.Background(), repo, logger.Noop())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	flag2 := newFlag("a", "string", "v2", true)
	repo.allFlags = []*domain.FeatureFlag{flag2}
	repo.flags["a"] = flag2

	ce.Invalidate(context.Background())

	got := ce.GetString(context.Background(), "a", nil)
	if got != "v2" {
		t.Errorf("after Invalidate, GetString = %q, want %q", got, "v2")
	}
}

func TestCachedEvaluator_LoadAll_ClearsOldEntries(t *testing.T) {
	flagA := newFlag("a", "bool", "true", true)
	flagB := newFlag("b", "bool", "true", true)
	repo := buildRepo(flagA, flagB)

	ce, err := NewCachedEvaluator(context.Background(), repo, logger.Noop())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := ce.cache.Load("a"); !ok {
		t.Fatal("flag 'a' should be in cache")
	}
	if _, ok := ce.cache.Load("b"); !ok {
		t.Fatal("flag 'b' should be in cache")
	}

	repo.allFlags = []*domain.FeatureFlag{flagB}
	if err := ce.LoadAll(context.Background()); err != nil {
		t.Fatalf("LoadAll error: %v", err)
	}

	if _, ok := ce.cache.Load("a"); ok {
		t.Error("flag 'a' should have been removed from cache after LoadAll")
	}
	if _, ok := ce.cache.Load("b"); !ok {
		t.Error("flag 'b' should still be in cache after LoadAll")
	}
}
