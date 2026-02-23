package translation

import (
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/internal/domain"
	"gct/pkg/httpx"

	"github.com/gin-gonic/gin"
)

// Delete godoc
// @Summary     Delete translations
// @Description Delete translations for an entity. Add ?lang=uz to delete only one language.
// @Tags        translations
// @Produce     json
// @Param       entity_type path   string  true  "Entity type"
// @Param       entity_id   path   string  true  "Entity UUID"
// @Param       lang        query  string  false "Language code — omit to delete all languages"
// @Success     200 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     401 {object} response.ErrorResponse
// @Security    BearerAuth
// @Router      /translations/{entity_type}/{entity_id} [delete]
func (c *Controller) Delete(ctx *gin.Context) {
	entityType := ctx.Param("entity_type")

	entityID, err := httpx.GetUUIDParam(ctx, "entity_id")
	if err != nil {
		response.ControllerResponse(ctx, http.StatusBadRequest, "invalid entity_id", nil, false)
		return
	}

	filter := domain.TranslationFilter{
		EntityType: entityType,
		EntityID:   entityID,
	}

	if lang := ctx.Query("lang"); lang != "" {
		filter.LangCode = &lang
	}

	if err := c.uc.Delete(ctx.Request.Context(), filter); err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, nil, nil, true)
}
