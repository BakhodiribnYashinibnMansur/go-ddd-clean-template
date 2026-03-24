package minio

import (
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/internal/domain/mock"
	"gct/internal/shared/infrastructure/httpx"

	"github.com/gin-gonic/gin"
)

// UploadVideo godoc
// @Summary     Upload video
// @Description Upload a video file
// @Tags        files
// @Accept      multipart/form-data
// @Produce     json
// @Param       file formData file true "Video file (mp4, avi, mov)"
// @Success     200 {object} response.SuccessResponse
// @Failure     401 {object} response.ErrorResponse
// @Failure     403 {object} response.ErrorResponse
// @Failure     400 {object} response.ErrorResponse
// @Security    BearerAuth
// @Router      /files/upload/video [post]
func (h *Controller) UploadVideo(ctx *gin.Context) {
	// Handle mock mode
	if httpx.Mock(ctx, httpx.MockTypeGet, func() any { return mock.FileInfoVideo().FileName }) {
		return
	}
	file, err := ctx.FormFile("file")
	if err != nil {
		response.ControllerResponse(ctx, http.StatusBadRequest, err, nil, false)
		return
	}
	fileContentType := file.Header[headerContentType][0]
	if !videoContentTypes[fileContentType] {
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid file format", nil, false)
		return
	}

	fileMultipart, err := file.Open()
	if err != nil {
		response.ControllerResponse(ctx, http.StatusBadRequest, err, nil, false)
		return
	}
	defer fileMultipart.Close()

	videoFileName, err := h.useCase.Minio.UploadVideo(ctx.Request.Context(), fileMultipart, file.Size, fileContentType)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusBadRequest, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, videoFileName, nil, true)
}
