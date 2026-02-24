package webhook

import (
	"net/http"

	"gct/internal/controller/restapi/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (ctrl *Controller) Test(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.ControllerResponse(c, http.StatusBadRequest, err, nil, false)
		return
	}
	if err := ctrl.useCase.Test(c.Request.Context(), id); err != nil {
		response.ControllerResponse(c, http.StatusInternalServerError, err, nil, false)
		return
	}
	response.ControllerResponse(c, http.StatusOK, map[string]string{"status": "sent"}, nil, true)
}
