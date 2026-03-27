package middleware

import (
	"net/http"
	"strings"

	"gct/internal/shared/infrastructure/httpx/response"
	apperrors "gct/internal/shared/infrastructure/errors"

	"github.com/gin-gonic/gin"
)

// BindingErrorMiddleware converts Gin's plain text binding errors to JSON.
// Must be registered early in the middleware chain.
func BindingErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			// Check if this is a binding error and response hasn't been written yet
			for _, e := range c.Errors {
				if e.Type == gin.ErrorTypeBind {
					// Only respond if nothing was written yet
					if !c.Writer.Written() {
						appErr := apperrors.NewHandlerError(
							apperrors.ErrHandlerBadRequest,
							"Request validation failed",
						).WithDetails(e.Error())

						response.RespondWithError(c, appErr, http.StatusBadRequest)
						c.Abort()
						return
					}
				}
			}
		}

		// Also check if a 400 was written with plain text (catch Gin's default)
		if c.Writer.Status() == http.StatusBadRequest &&
			!c.Writer.Written() &&
			strings.Contains(c.Writer.Header().Get("Content-Type"), "text/plain") {

			appErr := apperrors.NewHandlerError(
				apperrors.ErrHandlerBadRequest,
				"Bad request",
			)
			response.RespondWithError(c, appErr, http.StatusBadRequest)
		}
	}
}
