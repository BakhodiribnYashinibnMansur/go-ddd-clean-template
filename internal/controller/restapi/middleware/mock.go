package middleware

import (
	"gct/config"
	"gct/consts"
	"gct/internal/controller/restapi/util"
	"github.com/gin-gonic/gin"
)

// MockMiddleware handles mocking requests based on query parameters
func MockMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only enable in Dev or Test environments
		if cfg.IsProd() {
			c.Next()
			return
		}

		// 1. Handle mock_delay
		util.HandleMockDelay(c)

		// 2. Handle mock_error
		if util.HandleMockError(c) {
			return
		}

		// 3. Handle mock_empty
		if util.HandleMockEmpty(c) {
			return
		}

		// 4. Handle mock flag (pass to context)
		if c.Query(consts.QueryMock) == "true" {
			c.Set(consts.CtxMockMode, true)
		}

		c.Next()
	}
}
