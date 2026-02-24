package announcement

import (
	"net/http"

	"gct/internal/controller/restapi/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (ctrl *Controller) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.ControllerResponse(c, http.StatusBadRequest, err, nil, false)
		return
	}
	res, err := ctrl.useCase.GetByID(c.Request.Context(), id)
	if err != nil {
		response.ControllerResponse(c, http.StatusNotFound, err, nil, false)
		return
	}
	response.ControllerResponse(c, http.StatusOK, res, nil, true)
}
