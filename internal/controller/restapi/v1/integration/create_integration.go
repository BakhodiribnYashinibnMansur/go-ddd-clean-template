package integration

import (
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/internal/domain"

	"github.com/gin-gonic/gin"
)

// CreateIntegration handles POST /integrations
func (ctrl *Controller) CreateIntegration(c *gin.Context) {
	var req domain.CreateIntegrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ControllerResponse(c, http.StatusBadRequest, err, nil, false)
		return
	}

	res, err := ctrl.useCase.CreateIntegration(c.Request.Context(), req)
	if err != nil {
		response.ControllerResponse(c, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(c, http.StatusCreated, res, nil, true)
}
