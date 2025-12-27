package response

import (
	"github.com/gin-gonic/gin"
)

type ResponseModel struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

func ControllerResponse(c *gin.Context, code int, data interface{}, meta interface{}, success bool) {
	resp := ResponseModel{
		Success: success,
		Meta:    meta,
	}

	if success {
		resp.Data = data
	} else {
		resp.Error = data
	}

	c.JSON(code, resp)
}
