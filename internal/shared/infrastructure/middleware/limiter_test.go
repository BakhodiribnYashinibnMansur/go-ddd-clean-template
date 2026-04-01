package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/config"

	"github.com/gin-gonic/gin"
)

func TestRateLimiter_Disabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	cfg := config.Limiter{
		Enabled: false,
		Limit:   100,
		Period:  "M",
	}

	// When disabled, should return a passthrough handler (no Redis needed)
	handler := RateLimiter(cfg, nil, &mockLog{})
	r.Use(handler)
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 when limiter disabled, got %d", w.Code)
	}
}

// RateLimiter with Enabled=true requires a real Redis client for the store.
// Testing with a live Redis client would be an integration test.
// The disabled path is the only unit-testable path.
