package response

import (
	"github.com/gin-gonic/gin"
)

// ============================================================================
// Success Response Structure
// ============================================================================

// SuccessResponse is the standard success response structure
type SuccessResponse struct {
	Status     string `example:"success"                          json:"status"` // "success" or "error"
	StatusCode int    `example:"200"                              json:"statusCode"`
	Message    string `example:"Operation completed successfully" json:"message"`
	Data       any    `json:"data,omitempty"`
	Meta       any    `json:"meta,omitempty"`
}

// Meta holds pagination metadata
type Meta struct {
	Total  int64 `json:"total"`
	Limit  int64 `json:"limit"`
	Offset int64 `json:"offset"`
	Page   int64 `json:"page"`
}

// ============================================================================
// Controller Helper Function (Backward Compatibility)
// ============================================================================

// ControllerResponse handles the response sending.
// Updated signature to handle diverse call patterns:
// 1. Error: (ctx, code, "error msg", nil, false)
// 2. Success with data: (ctx, code, userStruct, nil, true)
// 3. Success with message: (ctx, code, "message", nil, true)
// 4. Success with meta: (ctx, code, list, meta, true)
func ControllerResponse(c *gin.Context, code int, payload, meta any, success bool) {
	if !success {
		// Error case
		// payload is typically a string (message) or error
		var err error
		if e, ok := payload.(error); ok {
			err = e
		} else if msg, ok := payload.(string); ok {
			err = &simpleError{msg: msg}
		} else {
			err = &simpleError{msg: "Unknown error occurred"}
		}

		RespondWithError(c, err, code)
		return
	}

	// Success case
	message := "Success"
	data := payload

	res := SuccessResponse{
		Status:     "SUCCESS",
		StatusCode: code,
		Message:    message,
		Data:       data,
		Meta:       meta,
	}
	c.JSON(code, res)
}

// simpleError implements error interface for string messages
type simpleError struct {
	msg string
}

func (e *simpleError) Error() string {
	return e.msg
}
