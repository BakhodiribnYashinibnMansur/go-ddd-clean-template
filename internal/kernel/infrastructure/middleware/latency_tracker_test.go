package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gct/internal/kernel/infrastructure/metrics/latency"

	"github.com/gin-gonic/gin"
)

func TestLatencyTrackerMiddleware_RecordsLatency(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tr := latency.NewTracker(60)

	r := gin.New()
	r.Use(LatencyTracker(tr))
	r.GET("/test", func(c *gin.Context) {
		time.Sleep(1 * time.Millisecond)
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	stats := tr.Stats()
	if stats.Count != 1 {
		t.Errorf("expected Count=1, got %d", stats.Count)
	}
	if stats.P50 < 1*time.Millisecond {
		t.Errorf("expected P50 >= 1ms, got %v", stats.P50)
	}
}

func TestLatencyTrackerMiddleware_MultipleRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tr := latency.NewTracker(60)

	r := gin.New()
	r.Use(LatencyTracker(tr))
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	for i := 0; i < 10; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		r.ServeHTTP(w, req)
	}

	stats := tr.Stats()
	if stats.Count != 10 {
		t.Errorf("expected Count=10, got %d", stats.Count)
	}
}
