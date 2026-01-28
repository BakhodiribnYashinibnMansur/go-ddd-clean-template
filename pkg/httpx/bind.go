package httpx

import (
	"github.com/gin-gonic/gin"
)

// BindAndValidate is a generic helper function that binds and validates request data.
// It works in conjunction with BindingErrorMiddleware which handles error responses.
// Returns true if binding succeeds, false if it fails.
//
// Note: Error handling is done automatically by BindingErrorMiddleware.
// This function just performs the binding and returns success/failure status.
//
// Usage:
//
//	var req domain.SignUpIn
//	if !httpx.BindAndValidate(c, &req) {
//	    return // Error response will be sent by middleware
//	}
func BindAndValidate[T any](c *gin.Context, req *T) bool {
	err := c.ShouldBind(req)
	// Middleware will automatically handle the error and send JSON response
	return err == nil
}

// BindJSON is a convenience wrapper for JSON-specific binding.
// It uses ShouldBindJSON instead of ShouldBind.
// Error handling is done automatically by BindingErrorMiddleware.
func BindJSON[T any](c *gin.Context, req *T) bool {
	err := c.ShouldBindJSON(req)
	// Middleware will automatically handle the error and send JSON response
	return err == nil
}

// BindQuery binds query parameters to the given struct.
// Error handling is done automatically by BindingErrorMiddleware.
func BindQuery[T any](c *gin.Context, req *T) bool {
	err := c.ShouldBindQuery(req)
	// Middleware will automatically handle the error and send JSON response
	return err == nil
}

// BindURI binds URI parameters to the given struct.
// Error handling is done automatically by BindingErrorMiddleware.
func BindURI[T any](c *gin.Context, req *T) bool {
	err := c.ShouldBindUri(req)
	// Middleware will automatically handle the error and send JSON response
	return err == nil
}
