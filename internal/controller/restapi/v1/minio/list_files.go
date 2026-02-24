package minio

import (
	"net/http"
	"strconv"

	"gct/internal/controller/restapi/response"
	"gct/internal/domain"

	"github.com/gin-gonic/gin"
)

// ListFiles godoc
// @Summary     List uploaded files
// @Description Returns a paginated list of file metadata records
// @Tags        files
// @Produce     json
// @Param       search   query string false "Search by original name"
// @Param       mime_type query string false "Filter by MIME type"
// @Param       limit    query int    false "Page size" default(20)
// @Param       offset   query int    false "Page offset" default(0)
// @Success     200 {object} response.SuccessResponse
// @Failure     500 {object} response.ErrorResponse
// @Security    BearerAuth
// @Router      /files/list [get]
func (h *Controller) ListFiles(ctx *gin.Context) {
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(ctx.DefaultQuery("offset", "0"))

	filter := domain.FileMetadataFilter{
		Search:   ctx.Query("search"),
		MimeType: ctx.Query("mime_type"),
		Limit:    limit,
		Offset:   offset,
	}

	items, total, err := h.useCase.File.ListFiles(ctx.Request.Context(), filter)
	if err != nil {
		h.logger.Error(err)
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, items, response.Meta{
		Total:  total,
		Limit:  int64(limit),
		Offset: int64(offset),
	}, true)
}
