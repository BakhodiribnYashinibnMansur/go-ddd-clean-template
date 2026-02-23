package translation

import (
	"net/http"

	"gct/internal/controller/restapi/response"
	"gct/internal/domain"
	"gct/pkg/httpx"

	"github.com/gin-gonic/gin"
)

// Gets godoc
// @Summary     Get translations
// @Description Get all translations for an entity. Filter by ?lang=uz to get a specific language.
// @Tags        translations
// @Produce     json
// @Param       entity_type path   string  true  "Entity type"
// @Param       entity_id   path   string  true  "Entity UUID"
// @Param       lang        query  string  false "Language code (uz, ru, en)"
// @Success     200 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     401 {object} response.ErrorResponse
// @Security    BearerAuth
// @Router      /translations/{entity_type}/{entity_id} [get]
func (c *Controller) Gets(ctx *gin.Context) {
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

	result, err := c.uc.Gets(ctx.Request.Context(), filter)
	if err != nil {
		response.ControllerResponse(ctx, http.StatusInternalServerError, err, nil, false)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, result, nil, true)
}
