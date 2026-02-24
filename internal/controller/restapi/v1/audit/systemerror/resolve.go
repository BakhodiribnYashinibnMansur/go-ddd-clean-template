package systemerror

import (
	"net/http"

	"gct/internal/controller/restapi/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (ctrl *Controller) Resolve(c *gin.Context) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		response.ControllerResponse(c, http.StatusBadRequest, err, nil, false)
		return
	}

	var resolvedByStr *string
	if uid, exists := c.Get("user_id"); exists {
		if userID, ok := uid.(uuid.UUID); ok {
			s := userID.String()
			resolvedByStr = &s
		}
	}

	if err := ctrl.useCase.Resolve(c.Request.Context(), id, resolvedByStr); err != nil {
		response.ControllerResponse(c, http.StatusInternalServerError, err, nil, false)
		return
	}
	response.ControllerResponse(c, http.StatusOK, nil, nil, true)
}
