package emailtemplate

import (
	"net/http"

	"gct/internal/controller/restapi/response"

	"github.com/gin-gonic/gin"
)

func (ctrl *Controller) Get(c *gin.Context) {
	id := c.Param("id")
	res, err := ctrl.useCase.GetByID(c.Request.Context(), id)
	if err != nil {
		response.ControllerResponse(c, http.StatusNotFound, err, nil, false)
		return
	}
	response.ControllerResponse(c, http.StatusOK, res, nil, true)
}
