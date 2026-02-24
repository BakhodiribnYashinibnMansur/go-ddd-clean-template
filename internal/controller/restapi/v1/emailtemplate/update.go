package emailtemplate

import (
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/internal/domain"

	"github.com/gin-gonic/gin"
)

func (ctrl *Controller) Update(c *gin.Context) {
	id := c.Param("id")
	var req domain.UpdateEmailTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ControllerResponse(c, http.StatusBadRequest, err, nil, false)
		return
	}
	res, err := ctrl.useCase.Update(c.Request.Context(), id, req)
	if err != nil {
		response.ControllerResponse(c, http.StatusInternalServerError, err, nil, false)
		return
	}
	response.ControllerResponse(c, http.StatusOK, res, nil, true)
}
