package entity_test

import (
	"gct/internal/context/iam/generic/authz/domain/entity"
	"testing"
)

func TestScope_ValueObject(t *testing.T) {
	t.Parallel()

	s := entity.Scope{Path: "/api/users", Method: "GET"}

	if s.Path != "/api/users" {
		t.Errorf("expected path '/api/users', got %q", s.Path)
	}
	if s.Method != "GET" {
		t.Errorf("expected method 'GET', got %q", s.Method)
	}
}

func TestScope_Equality(t *testing.T) {
	t.Parallel()

	a := entity.Scope{Path: "/api/users", Method: "GET"}
	b := entity.Scope{Path: "/api/users", Method: "GET"}
	c := entity.Scope{Path: "/api/users", Method: "POST"}
	d := entity.Scope{Path: "/api/roles", Method: "GET"}

	if a != b {
		t.Error("identical scopes should be equal")
	}
	if a == c {
		t.Error("scopes with different methods should not be equal")
	}
	if a == d {
		t.Error("scopes with different paths should not be equal")
	}
}

func TestScope_ZeroValue(t *testing.T) {
	t.Parallel()

	var s entity.Scope
	if s.Path != "" {
		t.Errorf("expected empty path, got %q", s.Path)
	}
	if s.Method != "" {
		t.Errorf("expected empty method, got %q", s.Method)
	}
}
