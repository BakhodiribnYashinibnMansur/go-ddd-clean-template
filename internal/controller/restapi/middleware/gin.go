package middleware

import (
	"github.com/gin-gonic/gin"
)

// FiberMiddleware provide Fiber's built-in middlewares.
// See: https://docs.gofiber.io/api/middleware
func GinMiddleware(route *gin.Engine) {
	route.Use(
		gin.Logger(),
		gin.Recovery(),
		gin.ErrorLogger(),
		CORSMiddleware(),
	)
}
