package integration

import (
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/internal/domain"
	"gct/internal/shared/infrastructure/httpx"

	"github.com/gin-gonic/gin"
)

// CreateIntegration handles POST /integrations
func (ctrl *Controller) CreateIntegration(c *gin.Context) {
	var req domain.CreateIntegrationRequest
	if !httpx.BindJSON(c, &req) {
		return
	}

	res, err := ctrl.useCase.CreateIntegration(c.Request.Context(), req)
	if err != nil {
		response.RespondWithError(c, err, http.StatusInternalServerError)
		return
	}

	response.ControllerResponse(c, http.StatusCreated, res, nil, true)
}
