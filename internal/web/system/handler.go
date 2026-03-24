package system

import (
	"html/template"
	"net/http"
	"sort"
	"strings"

	"gct/internal/controller/restapi/response"
	apperrors "gct/internal/shared/infrastructure/errors"

	"github.com/gin-gonic/gin"
)

func (c *Controller) GetErrors(ctx *gin.Context) {
	var filter ErrorFilter
	_ = ctx.ShouldBindQuery(&filter)

	errorsList := apperrors.GetErrorsByFilter(filter.Layer, filter.Category, filter.Code)

	// Check if the client prefers HTML
	accept := ctx.GetHeader("Accept")
	if strings.Contains(accept, "text/html") {
		c.renderHTML(ctx, filter, errorsList)
		return
	}

	response.ControllerResponse(ctx, http.StatusOK, errorsList, nil, true)
}

func (c *Controller) renderHTML(ctx *gin.Context, filter ErrorFilter, errors []apperrors.ErrorDefinition) {
	// Group errors by Category
	grouped := make(map[string][]apperrors.ErrorDefinition)
	for _, err := range errors {
		grouped[err.Category] = append(grouped[err.Category], err)
	}

	// Define category order
	orderedCategories := []string{"Data", "Validation", "Security", "Business", "System"}
	var categories []CategoryData

	// Process ordered categories
	processed := make(map[string]bool)
	for _, catName := range orderedCategories {
		if errs, ok := grouped[catName]; ok {
			// Sort errors by Code within category
			sort.Slice(errs, func(i, j int) bool {
				return errs[i].Code < errs[j].Code
			})
			categories = append(categories, CategoryData{Name: catName, Errors: errs})
			processed[catName] = true
		}
	}

	// Process any remaining categories not in the ordered list
	for catName, errs := range grouped {
		if !processed[catName] {
			sort.Slice(errs, func(i, j int) bool {
				return errs[i].Code < errs[j].Code
			})
			categories = append(categories, CategoryData{Name: catName, Errors: errs})
		}
	}

	data := PageData{
		Filter:     filter,
		Categories: categories,
	}

	tmpl, err := template.New("error_page").Parse(htmlTemplate)
	if err != nil {
		c.l.Errorw("failed to parse error page template", "error", err)
		ctx.String(http.StatusInternalServerError, "Internal Server Error")
		return
	}

	ctx.Header("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(ctx.Writer, data); err != nil {
		c.l.Errorw("failed to execute error page template", "error", err)
	}
}
