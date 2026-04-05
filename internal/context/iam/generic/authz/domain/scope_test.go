package domain_test

import (
	"gct/internal/context/iam/generic/authz/domain"
	"testing"
)

func TestScope_ValueObject(t *testing.T) {
	t.Parallel()

	s := domain.Scope{Path: "/api/users", Method: "GET"}

	if s.Path != "/api/users" {
		t.Errorf("expected path '/api/users', got %q", s.Path)
	}
	if s.Method != "GET" {
		t.Errorf("expected method 'GET', got %q", s.Method)
	}
}

func TestScope_Equality(t *testing.T) {
	t.Parallel()

	a := domain.Scope{Path: "/api/users", Method: "GET"}
	b := domain.Scope{Path: "/api/users", Method: "GET"}
	c := domain.Scope{Path: "/api/users", Method: "POST"}
	d := domain.Scope{Path: "/api/roles", Method: "GET"}

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

	var s domain.Scope
	if s.Path != "" {
		t.Errorf("expected empty path, got %q", s.Path)
	}
	if s.Method != "" {
		t.Errorf("expected empty method, got %q", s.Method)
	}
}
