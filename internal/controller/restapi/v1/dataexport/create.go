package dataexport

import (
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/internal/domain"

	"github.com/gin-gonic/gin"
)

func (ctrl *Controller) Create(c *gin.Context) {
	var req domain.CreateDataExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ControllerResponse(c, http.StatusBadRequest, err, nil, false)
		return
	}
	// Get user ID from context (set by auth middleware)
	var userID string
	if uid, exists := c.Get("user_id"); exists {
		switch v := uid.(type) {
		case string:
			userID = v
		}
	}
	res, err := ctrl.useCase.Create(c.Request.Context(), req, userID)
	if err != nil {
		response.ControllerResponse(c, http.StatusInternalServerError, err, nil, false)
		return
	}
	response.ControllerResponse(c, http.StatusCreated, res, nil, true)
}
