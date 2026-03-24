package relation

import (
	"net/http"

	"gct/internal/shared/domain/consts"
	"gct/internal/controller/restapi/response"
	"gct/internal/domain"
	"gct/internal/domain/mock"
	"gct/internal/shared/infrastructure/httpx"

	"github.com/gin-gonic/gin"
)

// Gets godoc
// @Summary     List relations
// @Description Get list of relations with filtering
// @Tags        authz-relations
// @Accept      json
// @Produce     json
// @Param       limit query int false "Limit"
// @Param       offset query int false "Offset"
// @Param       type query string false "Type"
// @Param       name query string false "Name"
// @Success     200 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     401 {object} response.ErrorResponse
// @Failure     403 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Security    BearerAuth
// @Router      /authz/relations [get]
func (c *Controller) Gets(ctx *gin.Context) {
	pagination, err := httpx.GetPagination(ctx)
	if err != nil {
		httpx.LogError(c.l, err, "http - v1 - authz - relation - gets - pagination")
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid pagination", nil, false)
		return
	}

	filter := domain.RelationsFilter{
		Pagination: &pagination,
		RelationFilter: domain.RelationFilter{
			Type: func() *string {
				s := httpx.GetNullStringQuery(ctx, "type")
				if s == "" {
					return nil
				}
				return &s
			}(),
			Name: func() *string {
				s := httpx.GetNullStringQuery(ctx, consts.QueryName)
				if s == "" {
					return nil
				}
				return &s
			}(),
		},
	}

	// Handle mock mode
	if httpx.Mock(ctx, httpx.MockTypeGets, func(count int) any { return mock.Relations(count) }) {
		return
	}

	relations, count, err := c.u.Authz.Relation().Gets(ctx.Request.Context(), &filter)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, relations, count, true)
}
