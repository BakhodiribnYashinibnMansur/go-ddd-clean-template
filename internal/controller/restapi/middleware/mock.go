package middleware

import (
	"gct/config"
	"gct/internal/shared/domain/consts"
	"gct/internal/shared/infrastructure/httpx"

	"github.com/gin-gonic/gin"
)

// MockMiddleware facilitates frontend and integration testing by simulating server behaviors.
// It intercepts requests based on query parameters to inject delays, errors, or empty responses.
// This allows UI developers to test edge cases without needing backend code changes.
func MockMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Safety Guard: Disable mocking capabilities in Production to prevent abuse.
		if cfg.IsProd() {
			c.Next()
			return
		}

		// 1. Latency Simulation: "mock_delay"
		// Useful for testing loading spinners and timeout handling in clients.
		httpx.HandleMockDelay(c)

		// 2. Error Injection: "mock_error"
		// Forces the endpoint to return a specific error code.
		if httpx.HandleMockError(c) {
			return
		}

		// 3. Empty Response: "mock_empty"
		// Returns a 200 OK with null/empty body.
		if httpx.HandleMockEmpty(c) {
			return
		}

		// 4. Mock Data Mode: "mock=true"
		// Signals the controller logic to return fake/static data instead of querying the database.
		if c.Query(consts.QueryMock) == "true" {
			c.Set(consts.CtxMockMode, true)
		}

		c.Next()
	}
}
