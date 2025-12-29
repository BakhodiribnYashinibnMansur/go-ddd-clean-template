package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// var allowedOrigins = map[string]bool{
// 	"http://localhost:3000": true,
// 	"http://localhost:5173": true,
// }

func CORSMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// accessURL := "*"
		// clientHost := ctx.Request.Header.Get("Origin")
		// if clientHost != "" {
		// 	accessURL = clientHost
		// }
		ctx.Header("Access-Control-Allow-Origin", "*")
		ctx.Writer.Header().Set("Content-Type", "application/json")
		ctx.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		ctx.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Accept, Origin, Cache-Control, X-Requested-With,Access-Control-Request-Method , Access-Control-Request-Headers,Language,language,Accept-Language")

		ctx.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT , DELETE ,PATCH, HEAD")
		ctx.Header("Access-Control-Max-Age", "3600")
		if ctx.Request.Method == http.MethodOptions {
			ctx.AbortWithStatus(204)
			return
		}
		// if !allowedOrigins[clientHost] {
		// 	ctx.AbortWithStatus(403)
		// 	return
		// }
		ctx.Next()
	}
}
