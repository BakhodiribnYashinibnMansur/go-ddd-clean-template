package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestOTelMetricsDoesNotPanicOnCreation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("OTelMetrics panicked on creation: %v", r)
		}
	}()

	_ = OTelMetrics("test-service")
}

func TestOTelMetricsReturnsCorrectStatusCode(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(OTelMetrics("test-service"))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestOTelMetricsHandlesUnmatchedRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(OTelMetrics("test-service"))
	r.GET("/known", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Request to a route that is not registered; Gin will return 404.
	req := httptest.NewRequest(http.MethodGet, "/unknown-path", nil)
	w := httptest.NewRecorder()

	// NoRoute handler so the middleware chain executes fully.
	r.NoRoute(func(c *gin.Context) {
		c.Status(http.StatusNotFound)
	})

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d for unmatched route, got %d", http.StatusNotFound, w.Code)
	}
}

func TestOTelMetricsRecordsDifferentHTTPMethods(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(OTelMetrics("test-service"))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	r.POST("/test", func(c *gin.Context) {
		c.Status(http.StatusCreated)
	})

	tests := []struct {
		name       string
		method     string
		wantStatus int
	}{
		{"GET request", http.MethodGet, http.StatusOK},
		{"POST request", http.MethodPost, http.StatusCreated},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/test", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d for %s, got %d", tt.wantStatus, tt.method, w.Code)
			}
		})
	}
}
