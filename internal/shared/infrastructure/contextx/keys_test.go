package contextx

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequestID(t *testing.T) {
	tests := []struct {
		name string
		val  string
		want string
	}{
		{"set and get", "req-123", "req-123"},
		{"empty value", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := WithRequestID(context.Background(), tt.val)
			assert.Equal(t, tt.want, GetRequestID(ctx))
		})
	}
	t.Run("missing returns empty", func(t *testing.T) {
		assert.Equal(t, "", GetRequestID(context.Background()))
	})
}

func TestSessionID(t *testing.T) {
	tests := []struct {
		name string
		val  string
		want string
	}{
		{"set and get", "sess-456", "sess-456"},
		{"empty value", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := WithSessionID(context.Background(), tt.val)
			assert.Equal(t, tt.want, GetSessionID(ctx))
		})
	}
	t.Run("missing returns empty", func(t *testing.T) {
		assert.Equal(t, "", GetSessionID(context.Background()))
	})
}

func TestUserID(t *testing.T) {
	tests := []struct {
		name string
		val  any
		want any
	}{
		{"string id", "user-1", "user-1"},
		{"int id", 42, 42},
		{"nil id", nil, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := WithUserID(context.Background(), tt.val)
			assert.Equal(t, tt.want, GetUserID(ctx))
		})
	}
	t.Run("missing returns nil", func(t *testing.T) {
		assert.Nil(t, GetUserID(context.Background()))
	})
}

func TestUserRole(t *testing.T) {
	tests := []struct {
		name string
		val  string
		want string
	}{
		{"admin role", "admin", "admin"},
		{"empty role", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := WithUserRole(context.Background(), tt.val)
			assert.Equal(t, tt.want, GetUserRole(ctx))
		})
	}
	t.Run("missing returns empty", func(t *testing.T) {
		assert.Equal(t, "", GetUserRole(context.Background()))
	})
}

func TestIPAddress(t *testing.T) {
	tests := []struct {
		name string
		val  string
		want string
	}{
		{"ipv4", "192.168.1.1", "192.168.1.1"},
		{"ipv6", "::1", "::1"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := WithIPAddress(context.Background(), tt.val)
			assert.Equal(t, tt.want, GetIPAddress(ctx))
		})
	}
	t.Run("missing returns empty", func(t *testing.T) {
		assert.Equal(t, "", GetIPAddress(context.Background()))
	})
}

func TestUserAgentCtx(t *testing.T) {
	ctx := WithUserAgent(context.Background(), "Mozilla/5.0")
	assert.Equal(t, "Mozilla/5.0", GetUserAgent(ctx))

	t.Run("missing returns empty", func(t *testing.T) {
		assert.Equal(t, "", GetUserAgent(context.Background()))
	})
}

func TestAPIVersion(t *testing.T) {
	ctx := WithAPIVersion(context.Background(), "v2")
	assert.Equal(t, "v2", GetAPIVersion(ctx))

	t.Run("missing returns empty", func(t *testing.T) {
		assert.Equal(t, "", GetAPIVersion(context.Background()))
	})
}

func TestTraceID(t *testing.T) {
	ctx := WithTraceID(context.Background(), "trace-789")
	assert.Equal(t, "trace-789", GetTraceID(ctx))

	t.Run("missing returns empty", func(t *testing.T) {
		assert.Equal(t, "", GetTraceID(context.Background()))
	})
}

func TestContextKeysDoNotCollide(t *testing.T) {
	ctx := context.Background()
	ctx = WithRequestID(ctx, "req")
	ctx = WithSessionID(ctx, "sess")
	ctx = WithUserID(ctx, "uid")
	ctx = WithUserRole(ctx, "role")
	ctx = WithIPAddress(ctx, "127.0.0.1")
	ctx = WithUserAgent(ctx, "ua")
	ctx = WithAPIVersion(ctx, "v1")
	ctx = WithTraceID(ctx, "trace")

	assert.Equal(t, "req", GetRequestID(ctx))
	assert.Equal(t, "sess", GetSessionID(ctx))
	assert.Equal(t, "uid", GetUserID(ctx))
	assert.Equal(t, "role", GetUserRole(ctx))
	assert.Equal(t, "127.0.0.1", GetIPAddress(ctx))
	assert.Equal(t, "ua", GetUserAgent(ctx))
	assert.Equal(t, "v1", GetAPIVersion(ctx))
	assert.Equal(t, "trace", GetTraceID(ctx))
}
