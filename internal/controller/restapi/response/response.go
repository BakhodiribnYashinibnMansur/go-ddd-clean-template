package response

import (
	"github.com/gin-gonic/gin"
)

// ============================================================================
// Success Response Structure
// ============================================================================

// SuccessResponse is the standard success response structure
type SuccessResponse struct {
	Status     string `json:"status" example:"success"` // "success" or "error"
	StatusCode int    `json:"statusCode" example:"200"`
	Message    string `json:"message" example:"Operation completed successfully"`
	Data       any    `json:"data,omitempty"`
	Pagination any    `json:"pagination,omitempty"`
}

// ============================================================================
// Controller Helper Function (Backward Compatibility)
// ============================================================================

// ControllerResponse handles the response sending.
// Updated signature to handle diverse call patterns:
// 1. Error: (ctx, code, "error msg", nil, false)
// 2. Success with data: (ctx, code, userStruct, nil, true)
// 3. Success with message: (ctx, code, "message", nil, true)
// 4. Success with pagination: (ctx, code, list, pagination, true)
func ControllerResponse(c *gin.Context, code int, payload any, pagination any, success bool) {
	if !success {
		// Error case
		// payload odatda string (mesaj) yoki error bo'ladi
		var err error
		if e, ok := payload.(error); ok {
			err = e
		} else if msg, ok := payload.(string); ok {
			err = &simpleError{msg: msg}
		} else {
			err = &simpleError{msg: "Unknown error occurred"}
		}

		RespondWithError(c, err)
		return
	}

	// Success case
	var message string = "Success"
	var data any = payload

	// Agar payload string bo'lsa va message bo'sh bo'lsa, uni message deb olamiz?
	// Lekin hozirgi logic bo'yicha payload har doim data fieldiga tushadi.
	// Agar payload string bo'lib, data bo'sh bo'lishi kerak bo'lgan holatlar bo'lsa,
	// uni keyinchalik refinement qilish mumkin.

	res := SuccessResponse{
		Status:     "success",
		StatusCode: code,
		Message:    message,
		Data:       data,
		Pagination: pagination,
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
