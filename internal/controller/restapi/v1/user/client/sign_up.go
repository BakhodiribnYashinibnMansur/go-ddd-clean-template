package client

import (
	"net/http"

	"github.com/evrone/go-clean-template/internal/controller/restapi/response"
	"github.com/evrone/go-clean-template/internal/controller/restapi/util"
	"github.com/evrone/go-clean-template/internal/domain"
	uc_client "github.com/evrone/go-clean-template/internal/usecase/user/client"
	"github.com/gin-gonic/gin"
)

// SignUp godoc
// @Summary     Sign Up
// @Description Register a new user
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       request body domain.User true "User info"
// @Success     200 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Router      /auth/sign-up [post]
func (c *Controller) SignUp(ctx *gin.Context) {
	var user domain.User
	if err := ctx.ShouldBindJSON(&user); err != nil {
		util.LogError(c.l, err, "http - v1 - auth - signup - bind")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid request body", nil, false)
		return
	}

	username := ""
	if user.Username != nil {
		username = *user.Username
	}

	err := c.u.User.Client.SignUp(ctx.Request.Context(), uc_client.SignUpInput{
		Username: username,
		Phone:    user.Phone,
		Password: user.Password,
	})
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusCreated, "User created successfully", nil, true)
}
