package integration

import (
	"fmt"
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/internal/domain"
	"gct/internal/shared/infrastructure/httpx"

	"github.com/gin-gonic/gin"
)

// UpdateIntegration handles PUT /integrations/:id
func (ctrl *Controller) UpdateIntegration(c *gin.Context) {
	id, err := httpx.GetUUIDParam(c, "id")
	if err != nil {
		response.RespondWithError(c, fmt.Errorf("invalid integration id"), http.StatusBadRequest)
		return
	}

	var req domain.UpdateIntegrationRequest
	if !httpx.BindJSON(c, &req) {
		return
	}

	res, err := ctrl.useCase.UpdateIntegration(c.Request.Context(), id, req)
	if err != nil {
		response.RespondWithError(c, err, http.StatusInternalServerError)
		return
	}

	response.ControllerResponse(c, http.StatusOK, res, nil, true)
}
