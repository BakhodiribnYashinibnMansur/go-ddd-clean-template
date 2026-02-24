package setting

import (
	"net/http"

	"gct/consts"
	"gct/internal/controller/restapi/response"
	apperrors "gct/pkg/errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type setRequest struct {
	Value string `json:"value" binding:"required"`
}

func (c *Controller) Set(ctx *gin.Context) {
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

	key := ctx.Param("key")
	if key == "" {
		response.ControllerResponse(ctx, http.StatusBadRequest, "key is required", nil, false)
		return
	}

	var req setRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid request body", nil, false)
		return
	}

	if key == "passcode" {
		if err := c.uc.SetPasscode(ctx.Request.Context(), userID, req.Value); err != nil {
			response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
			return
		}
	} else {
		if err := c.uc.Set(ctx.Request.Context(), userID, key, req.Value); err != nil {
			response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
			return
		}
	}

	response.ControllerResponse(ctx, http.StatusOK, "setting saved", nil, true)
}
