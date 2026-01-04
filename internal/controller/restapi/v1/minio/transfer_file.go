package minio

import (
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/internal/controller/restapi/util"
	"github.com/gin-gonic/gin"
)

// TransferFile handles file transfer
func (h *Controller) TransferFile(ctx *gin.Context) {
	// Handle mock mode
	if util.Mock(ctx, util.MockTypeGet, func() any { return string("file_transferred") }) {
		return
	}
	response.ControllerResponse(ctx, http.StatusNotImplemented, "not implemented", nil, false)
}
