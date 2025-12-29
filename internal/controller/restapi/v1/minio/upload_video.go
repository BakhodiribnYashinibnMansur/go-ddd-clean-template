package minio

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gct/internal/controller/restapi/response"
)

// UploadVideo handles video upload
func (h *Controller) UploadVideo(ctx *gin.Context) {
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

	videoFileName, err := h.useCase.Minio.UploadVideo(fileMultipart, file.Size, fileContentType)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusBadRequest, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, videoFileName, nil, true)
}
