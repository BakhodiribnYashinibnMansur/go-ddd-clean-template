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

func init() {
	gin.SetMode(gin.TestMode)
}

func newTestContext(queryString string) *gin.Context {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/?"+queryString, nil)
	return c
}

func TestGetStringQuery(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		param   string
		want    string
		wantErr bool
	}{
		{"present", "name=hello", "name", "hello", false},
		{"missing", "", "name", "", true},
		{"empty_value", "name=", "name", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestContext(tt.query)
			got, err := GetStringQuery(c, tt.param)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestGetNullStringQuery(t *testing.T) {
	c := newTestContext("name=++hello++")
	got := GetNullStringQuery(c, "name")
	assert.Equal(t, "hello", got)

	c2 := newTestContext("")
	got2 := GetNullStringQuery(c2, "name")
	assert.Equal(t, "", got2)
}

func TestGetStringArrayQuery(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  []string
	}{
		{"multiple", "tags=a,b,c", []string{"a", "b", "c"}},
		{"single", "tags=a", []string{"a"}},
		{"empty", "", []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestContext(tt.query)
			got := GetStringArrayQuery(c, "tags")
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetInt64Query(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		want    int64
		wantErr bool
	}{
		{"valid", "count=42", int64(42), false},
		{"missing_returns_zero", "", int64(0), false},
		{"invalid", "count=abc", int64(0), true},
		{"negative", "count=-5", int64(-5), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestContext(tt.query)
			got, err := GetInt64Query(c, "count")
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestGetNullIntQuery(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		want    int
		wantErr bool
	}{
		{"valid", "num=10", 10, false},
		{"empty_returns_zero", "", 0, false},
		{"invalid", "num=xyz", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestContext(tt.query)
			got, err := GetNullIntQuery(c, "num")
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestGetFloat64Query(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		want    float64
		wantErr bool
	}{
		{"valid", "price=19.99", 19.99, false},
		{"empty_returns_zero", "", 0, false},
		{"invalid", "price=abc", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestContext(tt.query)
			got, err := GetFloat64Query(c, "price")
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.InDelta(t, tt.want, got, 0.001)
			}
		})
	}
}

func TestGetBooleanQuery(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		want    bool
		wantErr bool
	}{
		{"true", "active=true", true, false},
		{"false", "active=false", false, false},
		{"one", "active=1", true, false},
		{"missing", "", false, true},
		{"invalid", "active=maybe", false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestContext(tt.query)
			got, err := GetBooleanQuery(c, "active")
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestGetNullBooleanQuery(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		want    bool
		wantErr bool
	}{
		{"true", "flag=true", true, false},
		{"false", "flag=false", false, false},
		{"empty_returns_true", "", true, false},
		{"invalid", "flag=maybe", false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestContext(tt.query)
			got, err := GetNullBooleanQuery(c, "flag")
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestGetUUIDQuery(t *testing.T) {
	validUUID := "550e8400-e29b-41d4-a716-446655440000"
	tests := []struct {
		name    string
		query   string
		want    uuid.UUID
		wantErr bool
	}{
		{"valid", "id=" + validUUID, uuid.MustParse(validUUID), false},
		{"missing", "", uuid.Nil, true},
		{"invalid", "id=not-a-uuid", uuid.Nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestContext(tt.query)
			got, err := GetUUIDQuery(c, "id")
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestGetNullUUIDQuery(t *testing.T) {
	validUUID := "550e8400-e29b-41d4-a716-446655440000"

	t.Run("present", func(t *testing.T) {
		c := newTestContext("id=" + validUUID)
		got, err := GetNullUUIDQuery(c, "id")
		require.NoError(t, err)
		assert.Equal(t, uuid.MustParse(validUUID), got)
	})

	t.Run("missing_returns_nil", func(t *testing.T) {
		c := newTestContext("")
		got, err := GetNullUUIDQuery(c, "id")
		require.NoError(t, err)
		assert.Equal(t, uuid.Nil, got)
	})

	t.Run("invalid", func(t *testing.T) {
		c := newTestContext("id=bad")
		_, err := GetNullUUIDQuery(c, "id")
		assert.Error(t, err)
	})
}

func TestGetNullBooleanStringQuery(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		want    string
		wantErr bool
	}{
		{"true", "flag=true", "true", false},
		{"false", "flag=false", "false", false},
		{"empty", "", "", false},
		{"invalid", "flag=maybe", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestContext(tt.query)
			got, err := GetNullBooleanStringQuery(c, "flag")
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestGetPageQuery(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		want    int64
		wantErr bool
	}{
		{"default", "", int64(1), false},
		{"page_2", "page=2", int64(2), false},
		{"invalid", "page=abc", int64(0), true},
		{"negative", "page=-1", int64(0), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestContext(tt.query)
			got, err := GetPageQuery(c)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestGetPageSizeQuery(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		want    int64
		wantErr bool
	}{
		{"default", "", int64(10), false},
		{"custom", "pageSize=25", int64(25), false},
		{"invalid", "pageSize=abc", int64(0), true},
		{"negative", "pageSize=-5", int64(0), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestContext(tt.query)
			got, err := GetPageSizeQuery(c)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestGetPagination(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		wantLimit  int64
		wantOffset int64
		wantErr    bool
	}{
		{"defaults", "", int64(20), int64(0), false},
		{"custom", "limit=50&offset=10", int64(50), int64(10), false},
		{"limit_too_high", "limit=5000", int64(0), int64(0), true},
		{"negative_offset", "offset=-1", int64(0), int64(0), true},
		{"invalid_limit", "limit=abc", int64(0), int64(0), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestContext(tt.query)
			p, err := GetPagination(c)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantLimit, p.Limit)
				assert.Equal(t, tt.wantOffset, p.Offset)
			}
		})
	}
}
