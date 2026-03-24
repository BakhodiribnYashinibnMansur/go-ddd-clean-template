package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/internal/shared/infrastructure/logger"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func setupMiniredis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	return mr, client
}

func TestIdempotency_NoKeyHeaderPassesThrough(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := logger.New("debug")
	_, client := setupMiniredis(t)

	r := gin.New()
	r.Use(Idempotency(client, l))
	r.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{"id": "123"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	// No Idempotency-Key header
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestIdempotency_FirstRequestProcessed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := logger.New("debug")
	_, client := setupMiniredis(t)

	r := gin.New()
	r.Use(Idempotency(client, l))
	r.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{"id": "456"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.Header.Set(HeaderIdempotencyKey, "unique-key-001")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var body map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.Equal(t, "456", body["id"])
}

func TestIdempotency_DuplicateRequestReturnsCachedResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := logger.New("debug")
	_, client := setupMiniredis(t)

	callCount := 0

	r := gin.New()
	r.Use(Idempotency(client, l))
	r.POST("/test", func(c *gin.Context) {
		callCount++
		c.JSON(http.StatusCreated, gin.H{"id": "789", "call": callCount})
	})

	idempotencyKey := "unique-key-002"

	// First request
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodPost, "/test", nil)
	req1.Header.Set(HeaderIdempotencyKey, idempotencyKey)
	r.ServeHTTP(w1, req1)

	assert.Equal(t, http.StatusCreated, w1.Code)
	assert.Equal(t, 1, callCount)

	// Second request with same key should return cached response
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/test", nil)
	req2.Header.Set(HeaderIdempotencyKey, idempotencyKey)
	r.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusCreated, w2.Code)
	// Handler should NOT have been called again
	assert.Equal(t, 1, callCount)

	// Verify the idempotency hit header
	assert.Equal(t, "true", w2.Header().Get("X-Idempotency-Hit"))
}

func TestIdempotency_DifferentKeysProcessedSeparately(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := logger.New("debug")
	_, client := setupMiniredis(t)

	callCount := 0

	r := gin.New()
	r.Use(Idempotency(client, l))
	r.POST("/test", func(c *gin.Context) {
		callCount++
		c.JSON(http.StatusCreated, gin.H{"call": callCount})
	})

	// First request
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodPost, "/test", nil)
	req1.Header.Set(HeaderIdempotencyKey, "key-A")
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusCreated, w1.Code)

	// Second request with different key
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/test", nil)
	req2.Header.Set(HeaderIdempotencyKey, "key-B")
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusCreated, w2.Code)

	// Both should have been processed
	assert.Equal(t, 2, callCount)
}

func TestIdempotency_ServerErrorNotCached(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := logger.New("debug")
	_, client := setupMiniredis(t)

	callCount := 0

	r := gin.New()
	r.Use(Idempotency(client, l))
	r.POST("/test", func(c *gin.Context) {
		callCount++
		if callCount == 1 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db down"})
		} else {
			c.JSON(http.StatusCreated, gin.H{"id": "ok"})
		}
	})

	key := "key-500-retry"

	// First request returns 500
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodPost, "/test", nil)
	req1.Header.Set(HeaderIdempotencyKey, key)
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusInternalServerError, w1.Code)

	// Retry with same key should process again (500s not cached)
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/test", nil)
	req2.Header.Set(HeaderIdempotencyKey, key)
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusCreated, w2.Code)
	assert.Equal(t, 2, callCount)
}

func TestIdempotency_GetRequestWithKeyStillWorks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := logger.New("debug")
	_, client := setupMiniredis(t)

	r := gin.New()
	r.Use(Idempotency(client, l))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"data": "list"})
	})

	// Even GET with idempotency key should work
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(HeaderIdempotencyKey, "get-key-001")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
