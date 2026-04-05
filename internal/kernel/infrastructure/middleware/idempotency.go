package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"gct/internal/kernel/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	HeaderIdempotencyKey = "Idempotency-Key"
	IdempotencyKeyPrefix = "idempotency:"
	IdempotencyTTL       = 24 * time.Hour
	LockTTL              = 30 * time.Second // Time to hold lock while processing
)

type CachedResponse struct {
	Status  int                 `json:"status"`
	Headers map[string][]string `json:"headers"`
	Body    []byte              `json:"body"`
}

type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	n, err := w.ResponseWriter.Write(b)
	if err != nil {
		return n, fmt.Errorf("middleware.idempotency.Write: %w", err)
	}
	return n, nil
}

func (w *responseWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	n, err := w.ResponseWriter.WriteString(s)
	if err != nil {
		return n, fmt.Errorf("middleware.idempotency.WriteString: %w", err)
	}
	return n, nil
}

func Idempotency(client *redis.Client, l logger.Log) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Check for Idempotency-Key header
		key := c.GetHeader(HeaderIdempotencyKey)
		if key == "" {
			c.Next()
			return
		}

		ctx := c.Request.Context()
		redisKey := IdempotencyKeyPrefix + key

		// 2. Check existance in Redis
		val, err := client.Get(ctx, redisKey).Result()
		if err == nil {
			// Key exists
			if val == "PROCESSING" {
				// Concurrent request with same key
				c.AbortWithStatusJSON(http.StatusConflict, gin.H{
					"status":  "error",
					"message": "Request with this Idempotency-Key is currently being processed",
				})
				return
			}

			// Return cached response
			var resp CachedResponse
			if err := json.Unmarshal([]byte(val), &resp); err != nil {
				l.Errorw("failed to unmarshal cached response", zap.Error(err))
				c.Next() // Fallback to processing
				return
			}

			// Replay headers
			for k, v := range resp.Headers {
				for _, h := range v {
					c.Writer.Header().Add(k, h)
				}
			}
			c.Writer.Header().Set("X-Idempotency-Hit", "true")
			c.Data(resp.Status, c.Writer.Header().Get("Content-Type"), resp.Body)
			c.Abort()
			return
		} else if err != redis.Nil {
			l.Errorw("redis error checking idempotency key", zap.Error(err))
			// Fail open or fail closed? Fail open for now.
			c.Next()
			return
		}

		// 3. Key does not exist, mark as PROCESSING
		// Use SETNX to acquire lock
		// Value "PROCESSING", Expiration LockTTL (to prevent deadlock if crash)
		ok, err := client.SetNX(ctx, redisKey, "PROCESSING", LockTTL).Result()
		if err != nil {
			l.Errorw("redis error setting idempotency lock", zap.Error(err))
			c.Next()
			return
		}
		if !ok {
			// Race condition: another request beat us to it
			c.AbortWithStatusJSON(http.StatusConflict, gin.H{
				"status":  "error",
				"message": "Request with this Idempotency-Key is currently being processed",
			})
			return
		}

		// 4. Wrap writer to capture response
		w := &responseWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = w

		// 5. Process Request
		c.Next()

		// 6. Cache Response (only if 2xx/3xx/4xx, maybe skip 500s?)
		// For true idempotency, we should cache the result regardless, but usually we don't cache 500s to allow retries.
		// Let's cache everything for strictness, or maybe filter.
		// A common pattern is to NOT cache 500s so client can retry.
		if w.Status() >= 500 {
			// Release lock so it can be retried immediately
			client.Del(ctx, redisKey)
			return
		}

		cached := CachedResponse{
			Status:  w.Status(),
			Headers: w.Header(),
			Body:    w.body.Bytes(),
		}

		data, err := json.Marshal(cached)
		if err != nil {
			l.Errorw("failed to marshal cached response", zap.Error(err))
			// Don't fail the request, just don't cache
			client.Del(ctx, redisKey) // clear processing lock
			return
		}

		// Store final response with full TTL
		// Overwrite "PROCESSING"
		if err := client.Set(ctx, redisKey, data, IdempotencyTTL).Err(); err != nil {
			l.Errorw("failed to save idempotency response", zap.Error(err))
		}
	}
}
