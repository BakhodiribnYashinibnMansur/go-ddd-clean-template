// Package role manages the RBAC role lifecycle and their linkage to permissions and users.
package role

import (
	"gct/config"
	"gct/internal/usecase"
	"gct/pkg/logger"

	"github.com/gin-gonic/gin"
)

// Controller handles HTTP requests targeting the Role domain.
type Controller struct {
	u   *usecase.UseCase
	l   logger.Log
	cfg *config.Config
}

// ControllerI defines the standard contract for managing security roles.
type ControllerI interface {
	Create(c *gin.Context)           // Register a new system role.
	Get(c *gin.Context)              // Retrieve metadata for a single role.
	Gets(c *gin.Context)             // List roles with optional filtering and pagination.
	Update(c *gin.Context)           // Modify a role's properties (e.g. description).
	Delete(c *gin.Context)           // Remove a role from the system.
	Assign(c *gin.Context)           // Link a specific role to a user account.
	AddPermission(c *gin.Context)    // Grant a new permission to an existing role.
	RemovePermission(c *gin.Context) // Revoke a permission from a role.
}

// New initializes a new Role controller.
func New(u *usecase.UseCase, cfg *config.Config, l logger.Log) ControllerI {
	return &Controller{u: u, l: l, cfg: cfg}
}
