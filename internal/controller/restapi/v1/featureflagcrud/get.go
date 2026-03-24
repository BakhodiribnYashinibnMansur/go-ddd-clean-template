package featureflagcrud

import (
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/internal/shared/infrastructure/httpx"

	"github.com/gin-gonic/gin"
)

func (ctrl *Controller) Get(c *gin.Context) {
	id, err := httpx.GetUUIDParam(c, "id")
	if err != nil {
		response.RespondWithError(c, err, http.StatusBadRequest)
		return
	}
	res, err := ctrl.useCase.GetByID(c.Request.Context(), id)
	if err != nil {
		response.RespondWithError(c, err, http.StatusNotFound)
		return
	}
	response.ControllerResponse(c, http.StatusOK, res, nil, true)
}
