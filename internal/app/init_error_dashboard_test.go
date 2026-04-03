package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegisterErrorDashboardRoutes(t *testing.T) {
	r := gin.New()
	g := r.Group("/api/v1")
	registerErrorDashboardRoutes(g)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/errors/stats", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
}

func TestHandleErrorStats(t *testing.T) {
	r := gin.New()
	r.GET("/errors/stats", handleErrorStats)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/errors/stats", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Response should be valid JSON
	var resp any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("expected valid JSON, got error: %v", err)
	}
}
