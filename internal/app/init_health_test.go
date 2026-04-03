package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestHandleHealth(t *testing.T) {
	r := gin.New()
	r.GET("/health", handleHealth)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if resp["status"] != "ok" {
		t.Fatalf("expected status ok, got %v", resp["status"])
	}
	if _, ok := resp["uptime"]; !ok {
		t.Fatal("expected uptime field")
	}
}

func TestHandleReady_NoDeps(t *testing.T) {
	deps := healthDeps{}
	r := gin.New()
	r.GET("/ready", handleReady(deps))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ready", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if resp["status"] != "ok" {
		t.Fatalf("expected status ok, got %v", resp["status"])
	}

	checks, ok := resp["checks"].(map[string]any)
	if !ok {
		t.Fatal("expected checks map")
	}

	// All deps nil => "not configured"
	for _, key := range []string{"postgres", "redis", "asynq"} {
		check, ok := checks[key].(map[string]any)
		if !ok {
			t.Fatalf("expected %s check map", key)
		}
		if check["status"] != "not configured" {
			t.Fatalf("expected %s status 'not configured', got %v", key, check["status"])
		}
	}
}

func TestHandleReady_WithAsynqAddr(t *testing.T) {
	deps := healthDeps{
		asynqAddr: "localhost:6379",
	}
	r := gin.New()
	r.GET("/ready", handleReady(deps))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ready", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	checks := resp["checks"].(map[string]any)
	asynqCheck := checks["asynq"].(map[string]any)
	if asynqCheck["status"] != "ok" {
		t.Fatalf("expected asynq status ok, got %v", asynqCheck["status"])
	}
	if asynqCheck["addr"] != "localhost:6379" {
		t.Fatalf("expected asynq addr localhost:6379, got %v", asynqCheck["addr"])
	}
}

func TestRegisterHealthRoutes(t *testing.T) {
	r := gin.New()
	deps := healthDeps{}
	registerHealthRoutes(r, deps)

	// Test /health route exists
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected /health 200, got %d", w.Code)
	}

	// Test /ready route exists
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/ready", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected /ready 200, got %d", w.Code)
	}
}
