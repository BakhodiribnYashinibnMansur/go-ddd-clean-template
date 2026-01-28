package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/config"
	"gct/pkg/logger"
	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRateLimiter(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Setup miniredis
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	// Setup redis client
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	l := logger.New("error")
	cfg := config.Limiter{
		Enabled: true,
		Limit:   2,
		Period:  "M",
	}

	// Create router with middleware
	r := gin.New()
	r.Use(RateLimiter(cfg, client, l))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	// First request - OK
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Second request - OK
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	// Third request - Too Many Requests
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusTooManyRequests, w3.Code)
}
