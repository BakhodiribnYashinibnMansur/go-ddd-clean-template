package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"gct/internal/featureflag/domain"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Mock repository
// ---------------------------------------------------------------------------

type mockRepo struct {
	flags map[string]*domain.FeatureFlag
	err   error
}

func (m *mockRepo) FindByKey(_ context.Context, key string) (*domain.FeatureFlag, error) {
	if m.err != nil {
		return nil, m.err
	}
	ff, ok := m.flags[key]
	if !ok {
		return nil, errors.New("not found")
	}
	return ff, nil
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
func (m *mockRepo) FindAll(context.Context) ([]*domain.FeatureFlag, error) {
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

func repoWith(flags ...*domain.FeatureFlag) *mockRepo {
	m := &mockRepo{flags: make(map[string]*domain.FeatureFlag)}
	for _, ff := range flags {
		m.flags[ff.Key()] = ff
	}
	return m
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestPgEvaluator_IsEnabled(t *testing.T) {
	tests := []struct {
		name string
		repo *mockRepo
		key  string
		want bool
	}{
		{
			name: "active bool flag with true default returns true",
			repo: repoWith(newFlag("feat.on", "bool", "true", true)),
			key:  "feat.on",
			want: true,
		},
		{
			name: "inactive flag returns false",
			repo: repoWith(newFlag("feat.off", "bool", "false", false)),
			key:  "feat.off",
			want: false,
		},
		{
			name: "flag not found returns false",
			repo: repoWith(),
			key:  "missing",
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ev := NewPgEvaluator(tc.repo)
			got := ev.IsEnabled(context.Background(), tc.key, nil)
			if got != tc.want {
				t.Errorf("IsEnabled(%q) = %v, want %v", tc.key, got, tc.want)
			}
		})
	}
}

func TestPgEvaluator_GetString(t *testing.T) {
	tests := []struct {
		name string
		repo *mockRepo
		key  string
		want string
	}{
		{
			name: "returns evaluated string value",
			repo: repoWith(newFlag("color", "string", "blue", true)),
			key:  "color",
			want: "blue",
		},
		{
			name: "returns empty string when not found",
			repo: repoWith(),
			key:  "missing",
			want: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ev := NewPgEvaluator(tc.repo)
			got := ev.GetString(context.Background(), tc.key, nil)
			if got != tc.want {
				t.Errorf("GetString(%q) = %q, want %q", tc.key, got, tc.want)
			}
		})
	}
}

func TestPgEvaluator_GetInt(t *testing.T) {
	tests := []struct {
		name string
		repo *mockRepo
		key  string
		want int
	}{
		{
			name: "returns integer value",
			repo: repoWith(newFlag("limit", "int", "42", true)),
			key:  "limit",
			want: 42,
		},
		{
			name: "returns 0 for non-integer value",
			repo: repoWith(newFlag("bad-int", "int", "abc", true)),
			key:  "bad-int",
			want: 0,
		},
		{
			name: "returns 0 when not found",
			repo: repoWith(),
			key:  "missing",
			want: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ev := NewPgEvaluator(tc.repo)
			got := ev.GetInt(context.Background(), tc.key, nil)
			if got != tc.want {
				t.Errorf("GetInt(%q) = %d, want %d", tc.key, got, tc.want)
			}
		})
	}
}

func TestPgEvaluator_GetFloat(t *testing.T) {
	tests := []struct {
		name string
		repo *mockRepo
		key  string
		want float64
	}{
		{
			name: "returns float value",
			repo: repoWith(newFlag("rate", "float", "3.14", true)),
			key:  "rate",
			want: 3.14,
		},
		{
			name: "returns 0 for non-float value",
			repo: repoWith(newFlag("bad-float", "float", "xyz", true)),
			key:  "bad-float",
			want: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ev := NewPgEvaluator(tc.repo)
			got := ev.GetFloat(context.Background(), tc.key, nil)
			if got != tc.want {
				t.Errorf("GetFloat(%q) = %f, want %f", tc.key, got, tc.want)
			}
		})
	}
}
