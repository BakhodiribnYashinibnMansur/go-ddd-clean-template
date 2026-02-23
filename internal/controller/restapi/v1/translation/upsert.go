package translation

import (
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/internal/domain"
	"gct/pkg/httpx"

	"github.com/gin-gonic/gin"
)

// Upsert godoc
// @Summary     Upsert translations
// @Description Create or update translations for an entity. Merges new fields into existing data.
// @Tags        translations
// @Accept      json
// @Produce     json
// @Param       entity_type path   string                             true  "Entity type (role, permission, relation, etc.)"
// @Param       entity_id   path   string                             true  "Entity UUID"
// @Param       request     body   domain.UpsertTranslationsRequest   true  "Translations body"
// @Success     200 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     401 {object} response.ErrorResponse
// @Security    BearerAuth
// @Router      /translations/{entity_type}/{entity_id} [put]
func (c *Controller) Upsert(ctx *gin.Context) {
	entityType := ctx.Param("entity_type")

	entityID, err := httpx.GetUUIDParam(ctx, "entity_id")
	if err != nil {
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid entity_id", nil, false)
		return
	}

	var req domain.UpsertTranslationsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		httpx.LogError(c.logger, err, "translation - upsert - bind")
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}

	if err := c.uc.Upsert(ctx.Request.Context(), entityType, entityID, req); err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, nil, nil, true)
}
