package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"gct/pkg/logger"
)

// Recovery -.
func Recovery(l logger.Log) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		l.Errorw("panic recovered",
			"error", recovered,
			"stack", string(debug.Stack()),
		)
		c.AbortWithStatus(http.StatusInternalServerError)
	})
}
