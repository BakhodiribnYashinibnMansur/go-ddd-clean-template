package httpx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractBearerToken(t *testing.T) {
	tests := []struct {
		name       string
		authHeader string
		expected   string
	}{
		{
			name:       "valid_bearer_token",
			authHeader: "Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9",
			expected:   "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9",
		},
		{
			name:       "empty_header",
			authHeader: "",
			expected:   "",
		},
		{
			name:       "missing_bearer_prefix",
			authHeader: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9",
			expected:   "",
		},
		{
			name:       "wrong_prefix",
			authHeader: "Basic dXNlcm5hbWU6cGFzc3dvcmQ=",
			expected:   "",
		},
		{
			name:       "malformed_no_space",
			authHeader: "BearereyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9",
			expected:   "",
		},
		{
			name:       "extra_spaces",
			authHeader: "Bearer  token with spaces",
			expected:   "",
		},
		{
			name:       "lowercase_bearer",
			authHeader: "bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9",
			expected:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractBearerToken(tt.authHeader)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractBasicToken(t *testing.T) {
	tests := []struct {
		name       string
		authHeader string
		expected   string
	}{
		{
			name:       "valid_basic_token",
			authHeader: "Basic dXNlcm5hbWU6cGFzc3dvcmQ=",
			expected:   "dXNlcm5hbWU6cGFzc3dvcmQ=",
		},
		{
			name:       "empty_header",
			authHeader: "",
			expected:   "",
		},
		{
			name:       "missing_basic_prefix",
			authHeader: "dXNlcm5hbWU6cGFzc3dvcmQ=",
			expected:   "",
		},
		{
			name:       "wrong_prefix",
			authHeader: "Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9",
			expected:   "",
		},
		{
			name:       "malformed_no_space",
			authHeader: "BasicdXNlcm5hbWU6cGFzc3dvcmQ=",
			expected:   "",
		},
		{
			name:       "lowercase_basic",
			authHeader: "basic dXNlcm5hbWU6cGFzc3dvcmQ=",
			expected:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractBasicToken(tt.authHeader)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseAuthorizationType(t *testing.T) {
	tests := []struct {
		name       string
		authHeader string
		expected   string
	}{
		{
			name:       "bearer_token",
			authHeader: "Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9",
			expected:   "Bearer",
		},
		{
			name:       "basic_token",
			authHeader: "Basic dXNlcm5hbWU6cGFzc3dvcmQ=",
			expected:   "Basic",
		},
		{
			name:       "empty_header",
			authHeader: "",
			expected:   "",
		},
		{
			name:       "unknown_type",
			authHeader: "Digest username=\"user\"",
			expected:   "",
		},
		{
			name:       "bearer_uppercase",
			authHeader: "BEARER token123",
			expected:   "Bearer",
		},
		{
			name:       "basic_mixed_case",
			authHeader: "BaSiC token123",
			expected:   "Basic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseAuthorizationType(tt.authHeader)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Benchmark tests
func BenchmarkExtractBearerToken(b *testing.B) {
	authHeader := "Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.signature"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ExtractBearerToken(authHeader)
	}
}

func BenchmarkExtractBasicToken(b *testing.B) {
	authHeader := "Basic dXNlcm5hbWU6cGFzc3dvcmQ="
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ExtractBasicToken(authHeader)
	}
}

func BenchmarkParseAuthorizationType(b *testing.B) {
	authHeader := "Bearer token123"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseAuthorizationType(authHeader)
	}
}
