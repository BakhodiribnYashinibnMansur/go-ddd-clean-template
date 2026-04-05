package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/internal/kernel/consts"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func newTestContextWithHeaders(headers map[string]string) *gin.Context {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	c.Request = req
	return c
}

func TestGetLanguage(t *testing.T) {
	tests := []struct {
		name    string
		headers map[string]string
		want    string
	}{
		{"present", map[string]string{consts.HeaderLanguage: "FR"}, "fr"},
		{"default_en", map[string]string{}, "en"},
		{"lowercase", map[string]string{consts.HeaderLanguage: "de"}, "de"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestContextWithHeaders(tt.headers)
			assert.Equal(t, tt.want, GetLanguage(c))
		})
	}
}

func TestGetVersion(t *testing.T) {
	c := newTestContextWithHeaders(map[string]string{consts.HeaderAppVersion: "1.2.3"})
	assert.Equal(t, "1.2.3", GetVersion(c))

	c2 := newTestContextWithHeaders(map[string]string{})
	assert.Equal(t, "", GetVersion(c2))
}

func TestGetDeviceID(t *testing.T) {
	c := newTestContextWithHeaders(map[string]string{consts.HeaderXDeviceID: "device-abc"})
	assert.Equal(t, "device-abc", GetDeviceID(c))
}

func TestGetDeviceIDUUID(t *testing.T) {
	validUUID := uuid.New()
	t.Run("valid", func(t *testing.T) {
		c := newTestContextWithHeaders(map[string]string{consts.HeaderXDeviceID: validUUID.String()})
		assert.Equal(t, validUUID, GetDeviceIDUUID(c))
	})
	t.Run("invalid", func(t *testing.T) {
		c := newTestContextWithHeaders(map[string]string{consts.HeaderXDeviceID: "not-uuid"})
		assert.Equal(t, uuid.Nil, GetDeviceIDUUID(c))
	})
	t.Run("missing", func(t *testing.T) {
		c := newTestContextWithHeaders(map[string]string{})
		assert.Equal(t, uuid.Nil, GetDeviceIDUUID(c))
	})
}

func TestGetAPIKey(t *testing.T) {
	c := newTestContextWithHeaders(map[string]string{consts.HeaderXAPIKey: "my-api-key"})
	assert.Equal(t, "my-api-key", GetAPIKey(c))
}

func TestGetCtxRequestID(t *testing.T) {
	t.Run("present", func(t *testing.T) {
		c := newTestContextWithHeaders(map[string]string{consts.HeaderXRequestID: "req-123"})
		assert.Equal(t, "req-123", GetCtxRequestID(c))
	})
	t.Run("missing_generates_uuid", func(t *testing.T) {
		c := newTestContextWithHeaders(map[string]string{})
		result := GetCtxRequestID(c)
		_, err := uuid.Parse(result)
		assert.NoError(t, err, "should generate a valid UUID when header is missing")
	})
}

func TestGetAuthorization(t *testing.T) {
	c := newTestContextWithHeaders(map[string]string{consts.HeaderAuthorization: "Bearer token123"})
	assert.Equal(t, "Bearer token123", GetAuthorization(c))
}

func TestGetHeader(t *testing.T) {
	c := newTestContextWithHeaders(map[string]string{"X-Custom": "value"})
	assert.Equal(t, "value", GetHeader(c, "X-Custom"))
	assert.Equal(t, "", GetHeader(c, "X-Missing"))
}

func TestGetIPAddress(t *testing.T) {
	t.Run("ipv6_localhost_normalized", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		c.Request.RemoteAddr = "[::1]:12345"
		assert.Equal(t, "127.0.0.1", GetIPAddress(c))
	})
}

func TestGetClientDomain(t *testing.T) {
	tests := []struct {
		name   string
		origin string
		want   string
	}{
		{"with_protocol", "https://example.com", "example.com"},
		{"without_protocol", "example.com", "example.com"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestContextWithHeaders(map[string]string{consts.HeaderOrigin: tt.origin})
			assert.Equal(t, tt.want, GetClientDomain(c))
		})
	}
}

func TestGetApiKeyType(t *testing.T) {
	t.Run("present", func(t *testing.T) {
		c := newTestContextWithHeaders(map[string]string{consts.HeaderXApiKeyType: "admin"})
		got, err := GetApiKeyType(c)
		assert.NoError(t, err)
		assert.Equal(t, "admin", got)
	})
	t.Run("missing", func(t *testing.T) {
		c := newTestContextWithHeaders(map[string]string{})
		_, err := GetApiKeyType(c)
		assert.Error(t, err)
	})
}

func TestResponseHeaderXTotalCountWrite(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	ResponseHeaderXTotalCountWrite(c, 42)
	assert.Equal(t, "42", w.Header().Get(consts.HeaderXTotalCount))
}

func TestGenerateToken(t *testing.T) {
	token := GenerateToken()
	_, err := uuid.Parse(token)
	assert.NoError(t, err, "GenerateToken should return a valid UUID")
}

func TestGetForwardedProto(t *testing.T) {
	c := newTestContextWithHeaders(map[string]string{consts.HeaderXForwardedProto: "https"})
	assert.Equal(t, "https", GetForwardedProto(c))
}
