package middleware

import (
	"net/http"
	"runtime/debug"

	"gct/internal/platform/infrastructure/httpx/response"
	"gct/internal/platform/infrastructure/httpx"
	"gct/internal/platform/infrastructure/logger"

	"github.com/gin-gonic/gin"
)

// Recovery returns a Gin middleware that catches any unhandled panics during request processing.
// It logs the panic error along with a full stack trace to help with debugging, and sends
// a standardized 500 Internal Server Error response to the client instead of crashing the process.
//
// See: https://github.com/gin-gonic/gin#custom-recovery-behavior
func Recovery(l logger.Log) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		// Log the critical failure with structured context.
		// Including the stack trace is vital for post-mortem analysis.
		l.Errorw("panic recovered",
			"error", recovered,
			"stack", string(debug.Stack()),
		)

		// Return a generic error to the client to avoid leaking sensitive internal state.
		response.ControllerResponse(c, http.StatusInternalServerError, httpx.ErrPanicRecovered, nil, false)
		c.Abort()
	})
}
