package minio

import (
	"net/http"

	"gct/internal/controller/restapi/response"

	"github.com/gin-gonic/gin"
)

// DeleteFile godoc
// @Summary     Delete file metadata
// @Description Deletes a file metadata record from the database by ID.
// @Description Note: the actual object in MinIO/S3 storage is NOT removed by this endpoint.
// @Tags        files
// @Produce     json
// @Param       id path string true "File metadata UUID"
// @Success     200 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Security    BearerAuth
// @Router      /files/{id} [delete]
func (h *Controller) DeleteFile(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		response.ControllerResponse(ctx, http.StatusBadRequest, ErrInvalidFileFormat, nil, false)
		return
	}

	if err := h.useCase.File.DeleteFile(ctx.Request.Context(), id); err != nil {
		h.logger.Error(err)
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, nil, nil, true)
}
