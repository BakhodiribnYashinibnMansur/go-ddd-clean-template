package setting

import (
	"net/http"

	"gct/consts"
	"gct/internal/controller/restapi/response"
	apperrors "gct/pkg/errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (c *Controller) Gets(ctx *gin.Context) {
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

	settings, err := c.uc.Gets(ctx.Request.Context(), userID)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}
	response.ControllerResponse(ctx, http.StatusOK, settings, nil, true)
}
