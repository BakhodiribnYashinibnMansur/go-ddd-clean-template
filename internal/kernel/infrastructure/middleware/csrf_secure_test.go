package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gct/internal/kernel/consts"
	"gct/internal/kernel/infrastructure/security/csrf"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// mockCSRFStore implements csrf.Store for testing.
type mockCSRFStore struct {
	tokens map[string]*mockStoredToken
}

type mockStoredToken struct {
	hash      string
	expiresAt time.Time
}

func newMockCSRFStore() *mockCSRFStore {
	return &mockCSRFStore{tokens: make(map[string]*mockStoredToken)}
}

func (s *mockCSRFStore) Set(_ context.Context, sessionID, tokenHash string, expiration time.Duration) error {
	s.tokens[sessionID] = &mockStoredToken{hash: tokenHash, expiresAt: time.Now().Add(expiration)}
	return nil
}

func (s *mockCSRFStore) Get(_ context.Context, sessionID string) (string, time.Time, error) {
	t, ok := s.tokens[sessionID]
	if !ok {
		return "", time.Time{}, csrf.ErrCSRFTokenNotFound
	}
	return t.hash, t.expiresAt, nil
}

func (s *mockCSRFStore) Delete(_ context.Context, sessionID string) error {
	delete(s.tokens, sessionID)
	return nil
}

func (s *mockCSRFStore) Rotate(_ context.Context, sessionID, newHash string, expiration time.Duration) error {
	return s.Set(context.Background(), sessionID, newHash, expiration)
}

func TestCSRFSecure_SafeMethodBypass(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := &mockLog{}
	store := newMockCSRFStore()
	// Generator requires a secret; we can pass nil since safe methods should bypass entirely
	m := NewCSRFMiddleware(nil, store, l)

	for _, method := range []string{"GET", "HEAD", "OPTIONS"} {
		w := httptest.NewRecorder()
		_, r := gin.CreateTestContext(w)
		r.Use(m.Protect())
		r.Handle(method, "/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req, _ := http.NewRequest(method, "/test", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("[%s] expected 200, got %d", method, w.Code)
		}
	}
}

func TestCSRFSecure_MissingSessionID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := &mockLog{}
	store := newMockCSRFStore()
	m := NewCSRFMiddleware(nil, store, l)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.Use(m.Protect())
	r.POST("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("POST", "/test", nil)
	r.ServeHTTP(w, req)

	// No session ID in context -> 401
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for missing session ID, got %d", w.Code)
	}
}

func TestCSRFSecure_MissingCookieToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := &mockLog{}
	store := newMockCSRFStore()
	m := NewCSRFMiddleware(nil, store, l)

	sessionID := uuid.New()

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.Use(func(c *gin.Context) {
		c.Set(consts.CtxSessionID, sessionID.String())
		c.Next()
	})
	r.Use(m.Protect())
	r.POST("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("POST", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for missing cookie token, got %d", w.Code)
	}
}

func TestCSRFSecure_MissingHeaderToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := &mockLog{}
	store := newMockCSRFStore()
	m := NewCSRFMiddleware(nil, store, l)

	sessionID := uuid.New()

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.Use(func(c *gin.Context) {
		c.Set(consts.CtxSessionID, sessionID.String())
		c.Next()
	})
	r.Use(m.Protect())
	r.POST("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("POST", "/test", nil)
	req.AddCookie(&http.Cookie{Name: consts.CookieCsrfToken, Value: "some-token"})
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for missing header token, got %d", w.Code)
	}
}

func TestCSRFSecure_TokenMismatch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := &mockLog{}
	store := newMockCSRFStore()
	m := NewCSRFMiddleware(nil, store, l)

	sessionID := uuid.New()

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.Use(func(c *gin.Context) {
		c.Set(consts.CtxSessionID, sessionID.String())
		c.Next()
	})
	r.Use(m.Protect())
	r.POST("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("POST", "/test", nil)
	req.AddCookie(&http.Cookie{Name: consts.CookieCsrfToken, Value: "token-a"})
	req.Header.Set(consts.HeaderXCSRFToken, "token-b")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for token mismatch, got %d", w.Code)
	}
}

func TestCSRFSecure_TokenNotInStore(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := &mockLog{}
	store := newMockCSRFStore()
	m := NewCSRFMiddleware(nil, store, l)

	sessionID := uuid.New()

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.Use(func(c *gin.Context) {
		c.Set(consts.CtxSessionID, sessionID.String())
		c.Next()
	})
	r.Use(m.Protect())
	r.POST("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("POST", "/test", nil)
	req.AddCookie(&http.Cookie{Name: consts.CookieCsrfToken, Value: "matching-token"})
	req.Header.Set(consts.HeaderXCSRFToken, "matching-token")
	r.ServeHTTP(w, req)

	// Store has no token for this session -> 403
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for token not in store, got %d", w.Code)
	}
}

func TestCSRFSecure_HybridProtect_BearerBypass(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := &mockLog{}
	store := newMockCSRFStore()
	m := NewCSRFMiddleware(nil, store, l)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.Use(m.HybridProtect())
	r.POST("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("POST", "/test", nil)
	req.Header.Set("Authorization", "Bearer some-jwt-token")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for bearer bypass, got %d", w.Code)
	}
}

func TestCSRFSecure_HybridProtect_NoBearerFallsToCSRF(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := &mockLog{}
	store := newMockCSRFStore()
	m := NewCSRFMiddleware(nil, store, l)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.Use(m.HybridProtect())
	r.POST("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("POST", "/test", nil)
	// No Authorization header -> falls to CSRF check -> no session -> 401
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 when no bearer and no session, got %d", w.Code)
	}
}

func TestIsStateChangingMethod(t *testing.T) {
	tests := []struct {
		method string
		want   bool
	}{
		{"POST", true},
		{"PUT", true},
		{"PATCH", true},
		{"DELETE", true},
		{"GET", false},
		{"HEAD", false},
		{"OPTIONS", false},
	}

	for _, tt := range tests {
		got := isStateChangingMethod(tt.method)
		if got != tt.want {
			t.Errorf("isStateChangingMethod(%q) = %v, want %v", tt.method, got, tt.want)
		}
	}
}
