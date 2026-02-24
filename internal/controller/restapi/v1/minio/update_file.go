package minio

import (
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/internal/domain"

	"github.com/gin-gonic/gin"
)

// UpdateFile godoc
// @Summary     Update file metadata
// @Description Updates the original name of a file metadata record
// @Tags        files
// @Accept      json
// @Produce     json
// @Param       id   path string                        true "File metadata UUID"
// @Param       body body domain.UpdateFileMetadataRequest true "Update payload"
// @Success     200 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Security    BearerAuth
// @Router      /files/{id} [put]
func (h *Controller) UpdateFile(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		response.ControllerResponse(ctx, http.StatusBadRequest, ErrInvalidFileFormat, nil, false)
		return
	}

	var req domain.UpdateFileMetadataRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ControllerResponse(ctx, http.StatusBadRequest, err, nil, false)
		return
	}

	result, err := h.useCase.File.UpdateFile(ctx.Request.Context(), id, req)
	if err != nil {
		h.logger.Error(err)
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, result, nil, true)
}
