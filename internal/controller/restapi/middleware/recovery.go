package middleware

import (
	"net/http"
	"runtime/debug"

	"gct/pkg/logger"

	"github.com/gin-gonic/gin"
)

// Recovery -.
func Recovery(l logger.Log) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		l.WithContext(c.Request.Context()).Errorw("panic recovered",
			"error", recovered,
			"stack", string(debug.Stack()),
		)
		c.AbortWithStatus(http.StatusInternalServerError)
	})
}
