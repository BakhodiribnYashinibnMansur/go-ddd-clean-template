package http

import "github.com/gin-gonic/gin"

// RegisterRoutes registers all User HTTP routes on the given router group.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	users := rg.Group("/users")
	{
		users.POST("", h.Create)
		users.GET("", h.List)
		users.GET("/:id", h.Get)
		users.PATCH("/:id", h.Update)
		users.DELETE("/:id", h.Delete)
		users.POST("/:id/approve", h.Approve)
		users.POST("/:id/role", h.ChangeRole)
		users.POST("/bulk-action", h.BulkAction)
	}

	auth := rg.Group("/auth")
	{
		auth.POST("/sign-in", h.SignIn)
		auth.POST("/sign-up", h.SignUp)
		auth.POST("/sign-out", h.SignOut)
	}
}
