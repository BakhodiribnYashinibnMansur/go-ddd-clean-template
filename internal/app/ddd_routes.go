package app

import "github.com/gin-gonic/gin"

// RegisterDDDRoutes registers HTTP routes for all DDD bounded contexts.
// This will replace the current route setup in Plan 6 Part 2.
func RegisterDDDRoutes(router *gin.Engine, bcs *DDDBoundedContexts) {
	// TODO: Register routes for each BC
	// For now this is a placeholder.
	// Each BC will have its own interfaces/http/ layer with handlers.
	_ = router
	_ = bcs
}
