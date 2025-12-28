package system

import (
	"net/http"

	"github.com/evrone/go-clean-template/internal/controller/restapi/response"
	apperrors "github.com/evrone/go-clean-template/pkg/errors"
	"github.com/evrone/go-clean-template/pkg/logger"
	"github.com/gin-gonic/gin"
)

type Controller struct {
	l logger.Log
}

func NewController(l logger.Log) *Controller {
	return &Controller{l: l}
}

type ErrorFilter struct {
	Layer    string `form:"type" example:"Repository"`     // Maps to 'layer'
	Category string `form:"category" example:"Validation"` // Maps to 'category'
	Code     string `form:"code" example:"NOT_FOUND"`      // Maps to 'code'
}

// GetErrors godoc
// @Summary     Get system errors
// @Description Returns a list of system errors, optionally filtered by type (layer), category, or code.
// @Tags        system
// @Produce     json
// @Param       type     query    string  false  "Error Layer (Repository, Service, Handler)"
// @Param       category query    string  false  "Error Category (Data, Validation, Security, System, Business)"
// @Param       code     query    string  false  "Error Code (partial match)"
// @Success     200 {object} response.SuccessResponse{data=[]apperrors.ErrorDefinition}
// @Router      /system/errors [get]
func (c *Controller) GetErrors(ctx *gin.Context) {
	var filter ErrorFilter
	if err := ctx.ShouldBindQuery(&filter); err != nil {
		// Even if binding fails, we can proceed with empty filters or return bad request
		// For query params, usually ignored if invalid, but let's be safe
	}

	// Get filtered errors
	// Note: 'filter.Layer' comes from 'type' query param
	errorsList := apperrors.GetErrorsByFilter(filter.Layer, filter.Category, filter.Code)

	response.ControllerResponse(ctx, http.StatusOK, errorsList, nil, true)
}
