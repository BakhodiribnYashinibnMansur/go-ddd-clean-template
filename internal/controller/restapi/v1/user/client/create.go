package client

import (
	"net/http"

	"github.com/evrone/go-clean-template/internal/controller/restapi/response"
	"github.com/evrone/go-clean-template/internal/controller/restapi/util"
	"github.com/evrone/go-clean-template/internal/domain"
	uc_client "github.com/evrone/go-clean-template/internal/usecase/user/client"
	"github.com/gin-gonic/gin"
)

// Create godoc
// @Summary     Create a new user
// @Description Register a new user with username, phone and password
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       request body domain.User true "User creation query"
// @Success     201 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Router      /users [post]
func (c *Controller) Create(ctx *gin.Context) {
	var user domain.User
	if err := ctx.ShouldBindJSON(&user); err != nil {
		util.LogError(c.l, err, "http - v1 - client - create - bind")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid request body", nil, false)
		return
	}

	err := c.u.User.Client.Create(ctx.Request.Context(), uc_client.CreateInput{User: &user})
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusCreated, nil, nil, true)
}
