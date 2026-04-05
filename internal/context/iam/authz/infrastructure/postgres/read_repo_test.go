package postgres

import (
	"testing"
)

func TestMatchScope_ExactMatch(t *testing.T) {
	tests := []struct {
		name          string
		scopePath     string
		scopeMethod   string
		requestPath   string
		requestMethod string
		want          bool
	}{
		{
			name:          "exact path and method",
			scopePath:     "/api/v1/users",
			scopeMethod:   "GET",
			requestPath:   "/api/v1/users",
			requestMethod: "GET",
			want:          true,
		},
		{
			name:          "exact path different method",
			scopePath:     "/api/v1/users",
			scopeMethod:   "GET",
			requestPath:   "/api/v1/users",
			requestMethod: "POST",
			want:          false,
		},
		{
			name:          "different path same method",
			scopePath:     "/api/v1/users",
			scopeMethod:   "GET",
			requestPath:   "/api/v1/roles",
			requestMethod: "GET",
			want:          false,
		},
		{
			name:          "completely different",
			scopePath:     "/api/v1/users",
			scopeMethod:   "GET",
			requestPath:   "/api/v1/roles",
			requestMethod: "POST",
			want:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchScope(tt.scopePath, tt.scopeMethod, tt.requestPath, tt.requestMethod)
			if got != tt.want {
				t.Errorf("matchScope(%q, %q, %q, %q) = %v, want %v",
					tt.scopePath, tt.scopeMethod, tt.requestPath, tt.requestMethod, got, tt.want)
			}
		})
	}
}

func TestMatchScope_WildcardMethod(t *testing.T) {
	tests := []struct {
		name          string
		scopePath     string
		scopeMethod   string
		requestPath   string
		requestMethod string
		want          bool
	}{
		{
			name:          "wildcard method matches GET",
			scopePath:     "/api/v1/users",
			scopeMethod:   "*",
			requestPath:   "/api/v1/users",
			requestMethod: "GET",
			want:          true,
		},
		{
			name:          "wildcard method matches POST",
			scopePath:     "/api/v1/users",
			scopeMethod:   "*",
			requestPath:   "/api/v1/users",
			requestMethod: "POST",
			want:          true,
		},
		{
			name:          "wildcard method matches DELETE",
			scopePath:     "/api/v1/users",
			scopeMethod:   "*",
			requestPath:   "/api/v1/users",
			requestMethod: "DELETE",
			want:          true,
		},
		{
			name:          "wildcard method but different path",
			scopePath:     "/api/v1/users",
			scopeMethod:   "*",
			requestPath:   "/api/v1/roles",
			requestMethod: "GET",
			want:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchScope(tt.scopePath, tt.scopeMethod, tt.requestPath, tt.requestMethod)
			if got != tt.want {
				t.Errorf("matchScope(%q, %q, %q, %q) = %v, want %v",
					tt.scopePath, tt.scopeMethod, tt.requestPath, tt.requestMethod, got, tt.want)
			}
		})
	}
}

func TestMatchScope_WildcardPath(t *testing.T) {
	tests := []struct {
		name          string
		scopePath     string
		scopeMethod   string
		requestPath   string
		requestMethod string
		want          bool
	}{
		{
			name:          "global wildcard path matches anything",
			scopePath:     "*",
			scopeMethod:   "GET",
			requestPath:   "/api/v1/anything",
			requestMethod: "GET",
			want:          true,
		},
		{
			name:          "global wildcard path wrong method",
			scopePath:     "*",
			scopeMethod:   "GET",
			requestPath:   "/api/v1/anything",
			requestMethod: "POST",
			want:          false,
		},
		{
			name:          "global wildcard path and method",
			scopePath:     "*",
			scopeMethod:   "*",
			requestPath:   "/literally/anything",
			requestMethod: "PATCH",
			want:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchScope(tt.scopePath, tt.scopeMethod, tt.requestPath, tt.requestMethod)
			if got != tt.want {
				t.Errorf("matchScope(%q, %q, %q, %q) = %v, want %v",
					tt.scopePath, tt.scopeMethod, tt.requestPath, tt.requestMethod, got, tt.want)
			}
		})
	}
}

func TestMatchScope_PrefixMatch(t *testing.T) {
	tests := []struct {
		name          string
		scopePath     string
		scopeMethod   string
		requestPath   string
		requestMethod string
		want          bool
	}{
		{
			name:          "prefix matches sub-resource",
			scopePath:     "/api/v1/users*",
			scopeMethod:   "GET",
			requestPath:   "/api/v1/users/123",
			requestMethod: "GET",
			want:          true,
		},
		{
			name:          "prefix matches exact base",
			scopePath:     "/api/v1/users*",
			scopeMethod:   "GET",
			requestPath:   "/api/v1/users",
			requestMethod: "GET",
			want:          true,
		},
		{
			name:          "prefix matches deep nested",
			scopePath:     "/api/v1/users*",
			scopeMethod:   "GET",
			requestPath:   "/api/v1/users/123/sessions/456",
			requestMethod: "GET",
			want:          true,
		},
		{
			name:          "prefix does not match different base",
			scopePath:     "/api/v1/users*",
			scopeMethod:   "GET",
			requestPath:   "/api/v1/roles/123",
			requestMethod: "GET",
			want:          false,
		},
		{
			name:          "prefix with wildcard method",
			scopePath:     "/api/v1/users*",
			scopeMethod:   "*",
			requestPath:   "/api/v1/users/123",
			requestMethod: "DELETE",
			want:          true,
		},
		{
			name:          "prefix wrong method",
			scopePath:     "/api/v1/users*",
			scopeMethod:   "GET",
			requestPath:   "/api/v1/users/123",
			requestMethod: "POST",
			want:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchScope(tt.scopePath, tt.scopeMethod, tt.requestPath, tt.requestMethod)
			if got != tt.want {
				t.Errorf("matchScope(%q, %q, %q, %q) = %v, want %v",
					tt.scopePath, tt.scopeMethod, tt.requestPath, tt.requestMethod, got, tt.want)
			}
		})
	}
}

func TestMatchScope_CaseInsensitiveMethod(t *testing.T) {
	tests := []struct {
		name          string
		scopePath     string
		scopeMethod   string
		requestPath   string
		requestMethod string
		want          bool
	}{
		{
			name:          "lowercase scope method matches uppercase request",
			scopePath:     "/api/v1/users",
			scopeMethod:   "get",
			requestPath:   "/api/v1/users",
			requestMethod: "GET",
			want:          true,
		},
		{
			name:          "uppercase scope method matches lowercase request",
			scopePath:     "/api/v1/users",
			scopeMethod:   "POST",
			requestPath:   "/api/v1/users",
			requestMethod: "post",
			want:          true,
		},
		{
			name:          "mixed case both sides",
			scopePath:     "/api/v1/users",
			scopeMethod:   "Patch",
			requestPath:   "/api/v1/users",
			requestMethod: "PATCH",
			want:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchScope(tt.scopePath, tt.scopeMethod, tt.requestPath, tt.requestMethod)
			if got != tt.want {
				t.Errorf("matchScope(%q, %q, %q, %q) = %v, want %v",
					tt.scopePath, tt.scopeMethod, tt.requestPath, tt.requestMethod, got, tt.want)
			}
		})
	}
}

func TestMatchScope_EdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		scopePath     string
		scopeMethod   string
		requestPath   string
		requestMethod string
		want          bool
	}{
		{
			name:          "empty scope path does not match",
			scopePath:     "",
			scopeMethod:   "GET",
			requestPath:   "/api/v1/users",
			requestMethod: "GET",
			want:          false,
		},
		{
			name:          "empty request path matches empty scope",
			scopePath:     "",
			scopeMethod:   "GET",
			requestPath:   "",
			requestMethod: "GET",
			want:          true,
		},
		{
			name:          "trailing slash matters",
			scopePath:     "/api/v1/users/",
			scopeMethod:   "GET",
			requestPath:   "/api/v1/users",
			requestMethod: "GET",
			want:          false,
		},
		{
			name:          "path with query-like suffix no match",
			scopePath:     "/api/v1/users",
			scopeMethod:   "GET",
			requestPath:   "/api/v1/users?page=1",
			requestMethod: "GET",
			want:          false,
		},
		{
			name:          "prefix wildcard at root",
			scopePath:     "/*",
			scopeMethod:   "*",
			requestPath:   "/anything/at/all",
			requestMethod: "PUT",
			want:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchScope(tt.scopePath, tt.scopeMethod, tt.requestPath, tt.requestMethod)
			if got != tt.want {
				t.Errorf("matchScope(%q, %q, %q, %q) = %v, want %v",
					tt.scopePath, tt.scopeMethod, tt.requestPath, tt.requestMethod, got, tt.want)
			}
		})
	}
}
