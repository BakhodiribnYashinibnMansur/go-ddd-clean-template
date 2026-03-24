package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestContextWithParams creates a gin.Context with URL params set.
func newTestContextWithParams(params gin.Params) *gin.Context {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Params = params
	return c
}

func TestGetStringParam(t *testing.T) {
	tests := []struct {
		name    string
		params  gin.Params
		param   string
		want    string
		wantErr bool
	}{
		{
			name:   "present",
			params: gin.Params{{Key: "name", Value: "alice"}},
			param:  "name",
			want:   "alice",
		},
		{
			name:    "missing",
			params:  gin.Params{},
			param:   "name",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestContextWithParams(tt.params)
			got, err := GetStringParam(c, tt.param)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestGetUUIDParam(t *testing.T) {
	validUUID := "550e8400-e29b-41d4-a716-446655440000"
	tests := []struct {
		name    string
		params  gin.Params
		want    uuid.UUID
		wantErr bool
	}{
		{
			name:   "valid",
			params: gin.Params{{Key: "id", Value: validUUID}},
			want:   uuid.MustParse(validUUID),
		},
		{
			name:    "invalid",
			params:  gin.Params{{Key: "id", Value: "not-uuid"}},
			wantErr: true,
		},
		{
			name:    "missing",
			params:  gin.Params{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestContextWithParams(tt.params)
			got, err := GetUUIDParam(c, "id")
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestGetNullUUIDParam(t *testing.T) {
	validUUID := "550e8400-e29b-41d4-a716-446655440000"

	t.Run("present", func(t *testing.T) {
		c := newTestContextWithParams(gin.Params{{Key: "id", Value: validUUID}})
		got, err := GetNullUUIDParam(c, "id")
		require.NoError(t, err)
		assert.Equal(t, uuid.MustParse(validUUID), got)
	})

	t.Run("missing_returns_nil", func(t *testing.T) {
		c := newTestContextWithParams(gin.Params{})
		got, err := GetNullUUIDParam(c, "id")
		require.NoError(t, err)
		assert.Equal(t, uuid.Nil, got)
	})

	t.Run("invalid", func(t *testing.T) {
		c := newTestContextWithParams(gin.Params{{Key: "id", Value: "bad"}})
		_, err := GetNullUUIDParam(c, "id")
		assert.Error(t, err)
	})
}

func TestGetInt64Param(t *testing.T) {
	tests := []struct {
		name    string
		params  gin.Params
		want    int64
		wantErr bool
	}{
		{
			name:   "valid",
			params: gin.Params{{Key: "id", Value: "42"}},
			want:   42,
		},
		{
			name:    "invalid",
			params:  gin.Params{{Key: "id", Value: "abc"}},
			wantErr: true,
		},
		{
			name:    "missing",
			params:  gin.Params{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestContextWithParams(tt.params)
			got, err := GetInt64Param(c, "id")
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestGetNullInt64Param(t *testing.T) {
	t.Run("present", func(t *testing.T) {
		c := newTestContextWithParams(gin.Params{{Key: "id", Value: "99"}})
		got, err := GetNullInt64Param(c, "id")
		require.NoError(t, err)
		assert.Equal(t, int64(99), got)
	})

	t.Run("missing_returns_zero", func(t *testing.T) {
		c := newTestContextWithParams(gin.Params{})
		got, err := GetNullInt64Param(c, "id")
		require.NoError(t, err)
		assert.Equal(t, int64(0), got)
	})

	t.Run("invalid", func(t *testing.T) {
		c := newTestContextWithParams(gin.Params{{Key: "id", Value: "xyz"}})
		_, err := GetNullInt64Param(c, "id")
		assert.Error(t, err)
	})
}

func TestGetNullIntParam(t *testing.T) {
	t.Run("present", func(t *testing.T) {
		c := newTestContextWithParams(gin.Params{{Key: "n", Value: "7"}})
		got, err := GetNullIntParam(c, "n")
		require.NoError(t, err)
		assert.Equal(t, 7, got)
	})

	t.Run("missing", func(t *testing.T) {
		c := newTestContextWithParams(gin.Params{})
		got, err := GetNullIntParam(c, "n")
		require.NoError(t, err)
		assert.Equal(t, 0, got)
	})

	t.Run("invalid", func(t *testing.T) {
		c := newTestContextWithParams(gin.Params{{Key: "n", Value: "abc"}})
		_, err := GetNullIntParam(c, "n")
		assert.Error(t, err)
	})
}

func TestGetNullFloat64Param(t *testing.T) {
	t.Run("present", func(t *testing.T) {
		c := newTestContextWithParams(gin.Params{{Key: "lat", Value: "37.7749"}})
		got, err := GetNullFloat64Param(c, "lat")
		require.NoError(t, err)
		assert.InDelta(t, 37.7749, got, 0.0001)
	})

	t.Run("missing", func(t *testing.T) {
		c := newTestContextWithParams(gin.Params{})
		got, err := GetNullFloat64Param(c, "lat")
		require.NoError(t, err)
		assert.Equal(t, 0.0, got)
	})

	t.Run("invalid", func(t *testing.T) {
		c := newTestContextWithParams(gin.Params{{Key: "lat", Value: "abc"}})
		_, err := GetNullFloat64Param(c, "lat")
		assert.Error(t, err)
	})
}

func TestGetNullStringParam(t *testing.T) {
	t.Run("present", func(t *testing.T) {
		c := newTestContextWithParams(gin.Params{{Key: "slug", Value: "my-item"}})
		got, err := GetNullStringParam(c, "slug")
		require.NoError(t, err)
		assert.Equal(t, "my-item", got)
	})

	t.Run("missing", func(t *testing.T) {
		c := newTestContextWithParams(gin.Params{})
		got, err := GetNullStringParam(c, "slug")
		require.NoError(t, err)
		assert.Equal(t, "", got)
	})
}
