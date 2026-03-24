package featureflagcrud

import (
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/internal/domain"
	"gct/internal/shared/infrastructure/httpx"

	"github.com/gin-gonic/gin"
)

func (ctrl *Controller) Create(c *gin.Context) {
	var req domain.CreateFeatureFlagRequest
	if !httpx.BindJSON(c, &req) {
		return
	}
	res, err := ctrl.useCase.Create(c.Request.Context(), req)
	if err != nil {
		response.RespondWithError(c, err, http.StatusInternalServerError)
		return
	}
	response.ControllerResponse(c, http.StatusCreated, res, nil, true)
}
