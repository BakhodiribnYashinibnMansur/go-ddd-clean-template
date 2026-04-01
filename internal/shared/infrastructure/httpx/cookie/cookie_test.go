package cookie

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/config"
	"gct/internal/shared/domain/consts"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func testGinContext() (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	return c, w
}

func testCookieConfig() config.Cookie {
	return config.Cookie{
		Domain:   "localhost",
		HttpOnly: true,
		MaxAge:   3600,
		Secure:   true,
	}
}

func TestSaveCookies(t *testing.T) {
	c, w := testGinContext()
	cfg := testCookieConfig()

	data := map[string]string{
		"token":   "abc123",
		"session": "xyz789",
	}

	SaveCookies(c, data, cfg)

	cookies := w.Result().Cookies()
	if len(cookies) != 2 {
		t.Fatalf("expected 2 cookies, got %d", len(cookies))
	}

	found := make(map[string]string)
	for _, cookie := range cookies {
		found[cookie.Name] = cookie.Value
	}

	if found["token"] != "abc123" {
		t.Errorf("expected token cookie 'abc123', got %q", found["token"])
	}
	if found["session"] != "xyz789" {
		t.Errorf("expected session cookie 'xyz789', got %q", found["session"])
	}
}

func TestSaveCookies_SecurityAttributes(t *testing.T) {
	c, w := testGinContext()
	cfg := testCookieConfig()

	SaveCookies(c, map[string]string{"test": "value"}, cfg)

	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}

	cookie := cookies[0]
	if cookie.MaxAge != 3600 {
		t.Errorf("expected MaxAge 3600, got %d", cookie.MaxAge)
	}
	if cookie.Domain != "localhost" {
		t.Errorf("expected domain 'localhost', got %q", cookie.Domain)
	}
	if !cookie.Secure {
		t.Error("expected Secure to be true")
	}
	if !cookie.HttpOnly {
		t.Error("expected HttpOnly to be true")
	}
	if cookie.SameSite != http.SameSiteLaxMode {
		t.Errorf("expected SameSite Lax, got %v", cookie.SameSite)
	}
	if cookie.Path != consts.CookiePath {
		t.Errorf("expected path %q, got %q", consts.CookiePath, cookie.Path)
	}
}

func TestGetCookie_Exists(t *testing.T) {
	c, _ := testGinContext()
	c.Request.AddCookie(&http.Cookie{Name: "mykey", Value: "myvalue"})

	val := GetCookie(c, "mykey")
	if val != "myvalue" {
		t.Errorf("expected 'myvalue', got %q", val)
	}
}

func TestGetCookie_NotExists(t *testing.T) {
	c, _ := testGinContext()

	val := GetCookie(c, "nonexistent")
	if val != "" {
		t.Errorf("expected empty string for nonexistent cookie, got %q", val)
	}
}

func TestExpireCookies(t *testing.T) {
	c, w := testGinContext()
	cfg := testCookieConfig()

	ExpireCookies(c, cfg, "token", "session")

	cookies := w.Result().Cookies()
	if len(cookies) != 2 {
		t.Fatalf("expected 2 expired cookies, got %d", len(cookies))
	}

	for _, cookie := range cookies {
		if cookie.MaxAge != -1 {
			t.Errorf("expected MaxAge -1 for cookie %q, got %d", cookie.Name, cookie.MaxAge)
		}
		if cookie.Value != "" {
			t.Errorf("expected empty value for expired cookie %q, got %q", cookie.Name, cookie.Value)
		}
	}
}

func TestGetCookieConfig(t *testing.T) {
	cookie := GetCookieConfig("mykey", "myvalue")
	if cookie == nil {
		t.Fatal("expected non-nil cookie")
	}
	if cookie.Name != "mykey" {
		t.Errorf("expected name 'mykey', got %q", cookie.Name)
	}
	if cookie.Value != "myvalue" {
		t.Errorf("expected value 'myvalue', got %q", cookie.Value)
	}
	if cookie.MaxAge != consts.CookieExpiredTime {
		t.Errorf("expected MaxAge %d, got %d", consts.CookieExpiredTime, cookie.MaxAge)
	}
	if cookie.Path != consts.CookiePath {
		t.Errorf("expected path %q, got %q", consts.CookiePath, cookie.Path)
	}
	if cookie.Domain != consts.CookieDomain {
		t.Errorf("expected domain %q, got %q", consts.CookieDomain, cookie.Domain)
	}
	if !cookie.Secure {
		t.Error("expected Secure to be true")
	}
	if cookie.HttpOnly != consts.CookieHttpOnly {
		t.Errorf("expected HttpOnly %v, got %v", consts.CookieHttpOnly, cookie.HttpOnly)
	}
	if cookie.SameSite != http.SameSiteNoneMode {
		t.Errorf("expected SameSite None, got %v", cookie.SameSite)
	}
}
