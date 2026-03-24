package middleware

import (
	"net/http"

	"gct/internal/controller/restapi/response"
	apperrors "gct/internal/shared/infrastructure/errors"

	"github.com/gin-gonic/gin"
)

// MethodNotAllowedHandler returns a middleware that handles requests with unsupported HTTP methods.
// It returns a 405 Method Not Allowed status with a proper JSON error response.
func MethodNotAllowedHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		err := apperrors.NewHandlerError(
			apperrors.ErrHandlerMethodNotAllowed,
			"The HTTP method is not supported for this endpoint",
		).WithDetails("Please check the API documentation for supported methods")

		response.RespondWithError(ctx, err, http.StatusMethodNotAllowed)
	}
}
