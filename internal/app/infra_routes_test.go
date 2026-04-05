package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/config"

	"github.com/gin-gonic/gin"
)

func TestSetupAdminRedirectPage(t *testing.T) {
	r := gin.New()
	cfg := &config.Config{}
	cfg.Admin.URL = "http://localhost:3000"
	setupAdminRedirectPage(r, cfg)

	// Test /admin redirect
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected 307, got %d", w.Code)
	}
	if loc := w.Header().Get("Location"); loc != cfg.Admin.URL {
		t.Fatalf("expected redirect to %s, got %s", cfg.Admin.URL, loc)
	}
}

func TestSetupAdminRedirectPage_SubPath(t *testing.T) {
	r := gin.New()
	cfg := &config.Config{}
	cfg.Admin.URL = "http://localhost:3000"
	setupAdminRedirectPage(r, cfg)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin/dashboard", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected 307, got %d", w.Code)
	}
}

func TestSetupRootPage(t *testing.T) {
	r := gin.New()
	cfg := &config.Config{}
	setupRootPage(r, cfg)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	body := w.Body.String()
	if len(body) == 0 {
		t.Fatal("expected non-empty HTML body")
	}
	// Check it contains expected content
	if !containsString(body, "Go Clean Template") {
		t.Fatal("expected root page to contain 'Go Clean Template'")
	}
}

func TestSetupInfraRoutes_StaticEndpoints(t *testing.T) {
	r := gin.New()
	cfg := &config.Config{}
	cfg.Middleware.HealthCheck = true

	setupInfraRoutes(r, cfg, nil, nil, nil, nil, nil)

	// Test robots.txt
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/robots.txt", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected robots.txt 200, got %d", w.Code)
	}
	if !containsString(w.Body.String(), "Disallow") {
		t.Fatal("expected robots.txt to contain Disallow directive")
	}

	// Test favicon.ico
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/favicon.ico", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected favicon 204, got %d", w.Code)
	}

	// Test /healthz
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/healthz", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected /healthz 200, got %d", w.Code)
	}

	// Test /ping
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/ping", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected /ping 200, got %d", w.Code)
	}

	// Test /health/live
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/health/live", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected /health/live 200, got %d", w.Code)
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && stringContains(s, substr))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
