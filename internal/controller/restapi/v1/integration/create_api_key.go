package integration

import (
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateAPIKey handles POST /integrations/:id/keys
func (ctrl *Controller) CreateAPIKey(c *gin.Context) {
	integrationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.ControllerResponse(c, http.StatusBadRequest, "invalid integration id", nil, false)
		return
	}

	var req domain.CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ControllerResponse(c, http.StatusBadRequest, err, nil, false)
		return
	}
	req.IntegrationID = integrationID

	res, rawKey, err := ctrl.useCase.CreateAPIKey(c.Request.Context(), req)
	if err != nil {
		response.ControllerResponse(c, http.StatusInternalServerError, err, nil, false)
		return
	}

	// For security, raw key is only returned once during creation
	data := map[string]any{
		"api_key": res,
		"raw_key": rawKey,
	}

	response.ControllerResponse(c, http.StatusCreated, data, nil, true)
}
