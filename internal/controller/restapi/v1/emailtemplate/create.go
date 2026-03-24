package emailtemplate

import (
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/internal/domain"
	"gct/internal/shared/infrastructure/httpx"

	"github.com/gin-gonic/gin"
)

func (ctrl *Controller) Create(c *gin.Context) {
	var req domain.CreateEmailTemplateRequest
	if !httpx.BindJSON(c, &req) {
		return
	}
	res, err := ctrl.useCase.Create(c.Request.Context(), req)
	if err != nil {
		response.ControllerResponse(c, http.StatusInternalServerError, err, nil, false)
		return
	}
	response.ControllerResponse(c, http.StatusCreated, res, nil, true)
}
