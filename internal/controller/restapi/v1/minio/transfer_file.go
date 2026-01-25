package minio

import (
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/pkg/httpx"

	"github.com/gin-gonic/gin"
)

// TransferFile godoc
// @Summary     Transfer file
// @Description Transfer media files
// @Tags        files
// @Accept      json
// @Produce     json
// @Success     501 {object} response.ErrorResponse
// @Security    BearerAuth
// @Router      /files/transfer [post]
func (h *Controller) TransferFile(ctx *gin.Context) {
	// Handle mock mode
	if httpx.Mock(ctx, httpx.MockTypeGet, func() any { return string("file_transferred") }) {
		return
	}
	response.ControllerResponse(ctx, http.StatusNotImplemented, "not implemented", nil, false)
}
