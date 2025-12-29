package minio

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gct/internal/controller/restapi/response"
)

// TransferFile handles file transfer
func (h *Controller) TransferFile(ctx *gin.Context) {
	response.ControllerResponse(ctx, http.StatusNotImplemented, "not implemented", nil, false)
}
