package integration

import (
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UpdateIntegration handles PUT /integrations/:id
func (ctrl *Controller) UpdateIntegration(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.ControllerResponse(c, http.StatusBadRequest, "invalid integration id", nil, false)
		return
	}

	var req domain.UpdateIntegrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ControllerResponse(c, http.StatusBadRequest, err, nil, false)
		return
	}

	res, err := ctrl.useCase.UpdateIntegration(c.Request.Context(), id, req)
	if err != nil {
		response.ControllerResponse(c, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(c, http.StatusOK, res, nil, true)
}
