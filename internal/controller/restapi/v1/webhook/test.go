package webhook

import (
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/internal/shared/infrastructure/httpx"

	"github.com/gin-gonic/gin"
)

func (ctrl *Controller) Test(c *gin.Context) {
	id, err := httpx.GetUUIDParam(c, "id")
	if err != nil {
		response.RespondWithError(c, err, http.StatusBadRequest)
		return
	}
	if err := ctrl.useCase.Test(c.Request.Context(), id); err != nil {
		response.RespondWithError(c, err, http.StatusInternalServerError)
		return
	}
	response.ControllerResponse(c, http.StatusOK, map[string]string{"status": "sent"}, nil, true)
}
