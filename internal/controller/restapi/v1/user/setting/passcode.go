package setting

import (
	"net/http"

	"gct/consts"
	"gct/internal/controller/restapi/response"
	apperrors "gct/pkg/errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type verifyRequest struct {
	Passcode string `json:"passcode" binding:"required"`
}

func (c *Controller) VerifyPasscode(ctx *gin.Context) {
	userIDStr, ok := ctx.Get(consts.CtxUserID)
	if !ok {
		response.ControllerResponse(ctx, http.StatusUnauthorized, apperrors.ErrUnauthorized, nil, false)
		return
	}
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid user id", nil, false)
		return
	}

	var req verifyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid request body", nil, false)
		return
	}

	ok2, err := c.uc.VerifyPasscode(ctx.Request.Context(), userID, req.Passcode)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}
	if !ok2 {
		response.ControllerResponse(ctx, http.StatusUnauthorized, "incorrect passcode", nil, false)
		return
	}
	response.ControllerResponse(ctx, http.StatusOK, "passcode verified", nil, true)
}

func (c *Controller) RemovePasscode(ctx *gin.Context) {
	userIDStr, ok := ctx.Get(consts.CtxUserID)
	if !ok {
		response.ControllerResponse(ctx, http.StatusUnauthorized, apperrors.ErrUnauthorized, nil, false)
		return
	}
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid user id", nil, false)
		return
	}

	if err := c.uc.RemovePasscode(ctx.Request.Context(), userID); err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}
	response.ControllerResponse(ctx, http.StatusOK, "passcode removed", nil, true)
}
