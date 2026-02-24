package dashboard

import (
	"net/http"

	"gct/internal/controller/restapi/response"

	"github.com/gin-gonic/gin"
)

func (ctrl *Controller) Get(c *gin.Context) {
	stats, err := ctrl.useCase.Get(c.Request.Context())
	if err != nil {
		response.ControllerResponse(c, http.StatusInternalServerError, err, nil, false)
		return
	}
	response.ControllerResponse(c, http.StatusOK, stats, nil, true)
}
