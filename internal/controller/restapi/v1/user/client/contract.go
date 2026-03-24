// Package client manages user profile operations.
package client

import (
	"gct/config"
	"gct/internal/usecase"
	"gct/internal/shared/infrastructure/logger"

	"github.com/gin-gonic/gin"
)

// Controller implements the high-level business logic for user management.
type Controller struct {
	u   *usecase.UseCase
	l   logger.Log
	cfg *config.Config
}

// ControllerI defines the public contract for the client-side user operations.
type ControllerI interface {
	Create(c *gin.Context)     // Admin-only user creation.
	User(c *gin.Context)       // Fetch profile for the current user.
	Users(c *gin.Context)      // List and filter users (Admin).
	Update(c *gin.Context)     // Modify profile details.
	Delete(c *gin.Context)     // Account deactivation/deletion.
	BulkAction(c *gin.Context) // Bulk deactivate or delete users.
	Approve(c *gin.Context)    // Approve a pending user.
	ChangeRole(c *gin.Context) // Change a user's role.
}

// New initializes a User Client controller.
func New(u *usecase.UseCase, cfg *config.Config, l logger.Log) ControllerI {
	return &Controller{
		u:   u,
		l:   l,
		cfg: cfg,
	}
}
