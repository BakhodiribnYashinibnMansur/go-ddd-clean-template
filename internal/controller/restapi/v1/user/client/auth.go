package client

import (
	"net/http"

	"github.com/evrone/go-clean-template/internal/controller/restapi/v1/response"
	"github.com/gin-gonic/gin"
)

func (c *Controller) SignIn(ctx *gin.Context) {
	// TODO: Implement SignIn
	response.ControllerResponse(ctx, http.StatusNotImplemented, "not implemented", nil, false)
}

func (c *Controller) SignUp(ctx *gin.Context) {
	// TODO: Implement SignUp
	response.ControllerResponse(ctx, http.StatusNotImplemented, "not implemented", nil, false)
}

func (c *Controller) SignOut(ctx *gin.Context) {
	// TODO: Implement SignOut
	response.ControllerResponse(ctx, http.StatusNotImplemented, "not implemented", nil, false)
}
