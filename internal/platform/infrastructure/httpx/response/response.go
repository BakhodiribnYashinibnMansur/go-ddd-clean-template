// Package response defines the standard structure for all API responses,
// ensuring consistency across success and error outcomes.
package response

import (
	"gct/internal/platform/domain/consts"

	"github.com/gin-gonic/gin"
)

// ============================================================================
// Success Response Structure
// ============================================================================

// SuccessResponse represents the unified structure for all successful API outcomes.
// Parameters like Status and StatusCode provide explicit context, while Data and Meta
// store the primary payload and supporting information (like pagination).
type SuccessResponse struct {
	Status     string `example:"success"                          json:"status"` // "success" or "error"
	StatusCode int    `example:"200"                              json:"statusCode"`
	Message    string `example:"Operation completed successfully" json:"message"`
	Data       any    `json:"data,omitempty"`
	Meta       any    `json:"meta,omitempty"`
}

// Meta encapsulates pagination and other relevant environmental metadata.
// It helps clients navigate through large datasets using total count, limits, and offsets.
type Meta struct {
	Total  int64 `json:"total"`
	Limit  int64 `json:"limit"`
	Offset int64 `json:"offset"`
	Page   int64 `json:"page"`
}

// ============================================================================
// Controller Helper Function (Backward Compatibility)
// ============================================================================

// ControllerResponse is a versatile helper that serializes data into the standard JSON response format.
// It handles both success and error cases based on the provided parameters:
// 1. Error: Pass (ctx, code, "error msg", nil, false) - invokes RespondWithError internally.
// 2. Success with data: Pass (ctx, code, dataStruct, nil, true).
// 3. Success with message: Pass (ctx, code, "message string", nil, true).
// 4. Success with metadata: Pass (ctx, code, slice, metaStruct, true).
func ControllerResponse(c *gin.Context, code int, payload, meta any, success bool) {
	if !success {
		// Route internal errors through the central error handler
		var err error
		if e, ok := payload.(error); ok {
			err = e
		} else if msg, ok := payload.(string); ok {
			err = &simpleError{msg: msg}
		} else {
			err = &simpleError{msg: consts.ResponseMessageUnknownError}
		}

		RespondWithError(c, err, code)
		return
	}

	// Default success message if not overridden by custom logic
	message := consts.ResponseMessageSuccess
	data := payload

	res := SuccessResponse{
		Status:     consts.ResponseStatusSuccess,
		StatusCode: code,
		Message:    message,
		Data:       data,
		Meta:       meta,
	}
	c.JSON(code, res)
}

// simpleError provides a minimal implementation of the error interface for string-based messages.
type simpleError struct {
	msg string
}

// Error returns the underlying message string.
func (e *simpleError) Error() string {
	return e.msg
}
