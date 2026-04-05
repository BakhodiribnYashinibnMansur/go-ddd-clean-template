package domain_test

import (
	"testing"

	"gct/internal/context/iam/generic/authz/domain"

	"github.com/google/uuid"
)

func TestRoleID(t *testing.T) {
	t.Parallel()

	t.Run("round-trip", func(t *testing.T) {
		t.Parallel()
		id := domain.NewRoleID()
		if id.IsZero() {
			t.Fatal("newly generated RoleID should not be zero")
		}
		parsed, err := domain.ParseRoleID(id.String())
		if err != nil {
			t.Fatalf("domain.ParseRoleID round-trip failed: %v", err)
		}
		if parsed != id {
			t.Fatalf("round-trip mismatch: got %s, want %s", parsed, id)
		}
		if parsed.UUID() != id.UUID() {
			t.Fatal("UUID() mismatch")
		}
	})

	t.Run("invalid", func(t *testing.T) {
		t.Parallel()
		cases := []struct{ name, in string }{
			{"empty", ""},
			{"garbage", "not-a-uuid"},
			{"truncated", "123e4567-e89b-12d3-a456"},
		}
		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				if _, err := domain.ParseRoleID(tc.in); err == nil {
					t.Fatalf("expected error for %q, got nil", tc.in)
				}
			})
		}
	})

	t.Run("is-zero", func(t *testing.T) {
		t.Parallel()
		var zero domain.RoleID
		if !zero.IsZero() {
			t.Fatal("zero-valued RoleID should report IsZero()")
		}
		nonZero := domain.RoleID(uuid.New())
		if nonZero.IsZero() {
			t.Fatal("non-zero RoleID should not report IsZero()")
		}
	})

	t.Run("distinct", func(t *testing.T) {
		t.Parallel()
		a := domain.NewRoleID()
		b := domain.NewRoleID()
		if a == b {
			t.Fatal("separately generated IDs should differ")
		}
	})
}

func TestPermissionID(t *testing.T) {
	t.Parallel()

	t.Run("round-trip", func(t *testing.T) {
		t.Parallel()
		id := domain.NewPermissionID()
		if id.IsZero() {
			t.Fatal("newly generated PermissionID should not be zero")
		}
		parsed, err := domain.ParsePermissionID(id.String())
		if err != nil {
			t.Fatalf("domain.ParsePermissionID round-trip failed: %v", err)
		}
		if parsed != id {
			t.Fatalf("round-trip mismatch: got %s, want %s", parsed, id)
		}
		if parsed.UUID() != id.UUID() {
			t.Fatal("UUID() mismatch")
		}
	})

	t.Run("invalid", func(t *testing.T) {
		t.Parallel()
		cases := []struct{ name, in string }{
			{"empty", ""},
			{"garbage", "xyz"},
			{"truncated", "123e4567"},
		}
		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				if _, err := domain.ParsePermissionID(tc.in); err == nil {
					t.Fatalf("expected error for %q, got nil", tc.in)
				}
			})
		}
	})

	t.Run("is-zero", func(t *testing.T) {
		t.Parallel()
		var zero domain.PermissionID
		if !zero.IsZero() {
			t.Fatal("zero-valued PermissionID should report IsZero()")
		}
		nonZero := domain.PermissionID(uuid.New())
		if nonZero.IsZero() {
			t.Fatal("non-zero PermissionID should not report IsZero()")
		}
	})

	t.Run("distinct", func(t *testing.T) {
		t.Parallel()
		a := domain.NewPermissionID()
		b := domain.NewPermissionID()
		if a == b {
			t.Fatal("separately generated IDs should differ")
		}
	})
}

func TestPolicyID(t *testing.T) {
	t.Parallel()

	t.Run("round-trip", func(t *testing.T) {
		t.Parallel()
		id := domain.NewPolicyID()
		if id.IsZero() {
			t.Fatal("newly generated PolicyID should not be zero")
		}
		parsed, err := domain.ParsePolicyID(id.String())
		if err != nil {
			t.Fatalf("domain.ParsePolicyID round-trip failed: %v", err)
		}
		if parsed != id {
			t.Fatalf("round-trip mismatch: got %s, want %s", parsed, id)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		t.Parallel()
		cases := []struct{ name, in string }{
			{"empty", ""},
			{"garbage", "not-a-uuid"},
		}
		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				if _, err := domain.ParsePolicyID(tc.in); err == nil {
					t.Fatalf("expected error for %q, got nil", tc.in)
				}
			})
		}
	})

	t.Run("is-zero", func(t *testing.T) {
		t.Parallel()
		var zero domain.PolicyID
		if !zero.IsZero() {
			t.Fatal("zero-valued PolicyID should report IsZero()")
		}
		nonZero := domain.PolicyID(uuid.New())
		if nonZero.IsZero() {
			t.Fatal("non-zero PolicyID should not report IsZero()")
		}
	})

	t.Run("distinct", func(t *testing.T) {
		t.Parallel()
		a := domain.NewPolicyID()
		b := domain.NewPolicyID()
		if a == b {
			t.Fatal("separately generated IDs should differ")
		}
	})
}

func TestScopeID(t *testing.T) {
	t.Parallel()

	t.Run("round-trip", func(t *testing.T) {
		t.Parallel()
		id := domain.NewScopeID()
		if id.IsZero() {
			t.Fatal("newly generated ScopeID should not be zero")
		}
		parsed, err := domain.ParseScopeID(id.String())
		if err != nil {
			t.Fatalf("domain.ParseScopeID round-trip failed: %v", err)
		}
		if parsed != id {
			t.Fatalf("round-trip mismatch: got %s, want %s", parsed, id)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		t.Parallel()
		cases := []struct{ name, in string }{
			{"empty", ""},
			{"garbage", "not-a-uuid"},
		}
		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				if _, err := domain.ParseScopeID(tc.in); err == nil {
					t.Fatalf("expected error for %q, got nil", tc.in)
				}
			})
		}
	})

	t.Run("is-zero", func(t *testing.T) {
		t.Parallel()
		var zero domain.ScopeID
		if !zero.IsZero() {
			t.Fatal("zero-valued ScopeID should report IsZero()")
		}
		nonZero := domain.ScopeID(uuid.New())
		if nonZero.IsZero() {
			t.Fatal("non-zero ScopeID should not report IsZero()")
		}
	})

	t.Run("distinct", func(t *testing.T) {
		t.Parallel()
		a := domain.NewScopeID()
		b := domain.NewScopeID()
		if a == b {
			t.Fatal("separately generated IDs should differ")
		}
	})
}

func TestAuthzIDs_DistinctTypes(t *testing.T) {
	t.Parallel()

	// Ensure the typed IDs are distinct Go types — this is enforced at compile
	// time by the Go type system, so we merely construct each via the shared
	// uuid and assert they cannot be compared without conversion.
	u := uuid.New()
	r := domain.RoleID(u)
	p := domain.PermissionID(u)
	po := domain.PolicyID(u)
	s := domain.ScopeID(u)

	if r.UUID() != p.UUID() || p.UUID() != po.UUID() || po.UUID() != s.UUID() {
		t.Fatal("underlying UUIDs should match when constructed from the same source")
	}
}
