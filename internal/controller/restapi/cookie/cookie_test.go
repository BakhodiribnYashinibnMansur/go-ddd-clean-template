package cookie

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/config"
	"gct/internal/shared/domain/consts"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	return c, w
}

func testCookieCfg(secure, httpOnly bool) config.Cookie {
	return config.Cookie{
		Domain:   "example.com",
		Path:     "/",
		HttpOnly: httpOnly,
		MaxAge:   3600,
		Secure:   secure,
	}
}

// --------------- SaveCookies ---------------

func TestSaveCookies(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]string
		cfg      config.Cookie
		wantKeys []string
	}{
		{
			name:     "single cookie",
			data:     map[string]string{"token": "abc123"},
			cfg:      testCookieCfg(true, true),
			wantKeys: []string{"token"},
		},
		{
			name:     "multiple cookies",
			data:     map[string]string{"a": "1", "b": "2", "c": "3"},
			cfg:      testCookieCfg(false, false),
			wantKeys: []string{"a", "b", "c"},
		},
		{
			name:     "empty map sets no cookies",
			data:     map[string]string{},
			cfg:      testCookieCfg(true, true),
			wantKeys: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, w := newTestContext()
			SaveCookies(c, tt.data, tt.cfg)

			resp := w.Result()
			defer resp.Body.Close()

			cookies := resp.Cookies()

			// Build a lookup of returned cookies.
			got := make(map[string]*http.Cookie, len(cookies))
			for _, ck := range cookies {
				got[ck.Name] = ck
			}

			assert.Len(t, cookies, len(tt.wantKeys))
			for _, key := range tt.wantKeys {
				ck, ok := got[key]
				require.True(t, ok, "cookie %q not found in response", key)
				assert.Equal(t, tt.data[key], ck.Value)
				assert.Equal(t, tt.cfg.MaxAge, ck.MaxAge)
				assert.Equal(t, consts.CookiePath, ck.Path)
				assert.Equal(t, tt.cfg.Domain, ck.Domain)
				assert.Equal(t, tt.cfg.IsSecure(), ck.Secure)
				assert.Equal(t, tt.cfg.IsHttpOnly(), ck.HttpOnly)
				assert.Equal(t, http.SameSiteLaxMode, ck.SameSite)
			}
		})
	}
}

// --------------- GetCookie ---------------

func TestGetCookie(t *testing.T) {
	tests := []struct {
		name    string
		cookies []*http.Cookie
		key     string
		want    string
	}{
		{
			name:    "existing cookie",
			cookies: []*http.Cookie{{Name: "token", Value: "abc"}},
			key:     "token",
			want:    "abc",
		},
		{
			name:    "missing cookie returns empty",
			cookies: nil,
			key:     "token",
			want:    "",
		},
		{
			name:    "wrong key returns empty",
			cookies: []*http.Cookie{{Name: "other", Value: "xyz"}},
			key:     "token",
			want:    "",
		},
		{
			name:    "empty value cookie",
			cookies: []*http.Cookie{{Name: "token", Value: ""}},
			key:     "token",
			want:    "",
		},
		{
			name: "multiple cookies picks correct one",
			cookies: []*http.Cookie{
				{Name: "a", Value: "1"},
				{Name: "b", Value: "2"},
			},
			key:  "b",
			want: "2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := newTestContext()
			for _, ck := range tt.cookies {
				c.Request.AddCookie(ck)
			}
			got := GetCookie(c, tt.key)
			assert.Equal(t, tt.want, got)
		})
	}
}

// --------------- ExpireCookies ---------------

func TestExpireCookies(t *testing.T) {
	tests := []struct {
		name string
		keys []string
		cfg  config.Cookie
	}{
		{
			name: "expire single cookie",
			keys: []string{"token"},
			cfg:  testCookieCfg(true, true),
		},
		{
			name: "expire multiple cookies",
			keys: []string{"a", "b", "c"},
			cfg:  testCookieCfg(false, false),
		},
		{
			name: "no keys does nothing",
			keys: nil,
			cfg:  testCookieCfg(true, true),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, w := newTestContext()
			ExpireCookies(c, tt.cfg, tt.keys...)

			resp := w.Result()
			defer resp.Body.Close()
			cookies := resp.Cookies()

			assert.Len(t, cookies, len(tt.keys))
			for _, ck := range cookies {
				assert.Equal(t, "", ck.Value, "expired cookie %q should have empty value", ck.Name)
				assert.Equal(t, -1, ck.MaxAge, "expired cookie %q should have MaxAge -1", ck.Name)
				assert.Equal(t, consts.CookiePath, ck.Path)
				assert.Equal(t, tt.cfg.Domain, ck.Domain)
				assert.Equal(t, tt.cfg.IsSecure(), ck.Secure)
				assert.Equal(t, tt.cfg.IsHttpOnly(), ck.HttpOnly)
				assert.Equal(t, http.SameSiteLaxMode, ck.SameSite)
			}
		})
	}
}

// --------------- GetCookieConfig ---------------

func TestGetCookieConfig(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
	}{
		{
			name:  "standard cookie",
			key:   "session",
			value: "abc123",
		},
		{
			name:  "empty value",
			key:   "empty",
			value: "",
		},
		{
			name:  "special characters",
			key:   "data",
			value: "key=val&foo=bar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ck := GetCookieConfig(tt.key, tt.value)

			require.NotNil(t, ck)
			assert.Equal(t, tt.key, ck.Name)
			assert.Equal(t, tt.value, ck.Value)
			assert.Equal(t, consts.CookieExpiredTime, ck.MaxAge)
			assert.Equal(t, consts.CookiePath, ck.Path)
			assert.Equal(t, consts.CookieDomain, ck.Domain)
			assert.True(t, ck.Secure)
			assert.Equal(t, consts.CookieHttpOnly, ck.HttpOnly)
			assert.Equal(t, http.SameSiteNoneMode, ck.SameSite)
		})
	}
}
