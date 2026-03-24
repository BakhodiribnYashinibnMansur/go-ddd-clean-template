// Package session manages the lifecycle of active user logins across different devices and browsers.
package session

import (
	"gct/internal/usecase"
	"gct/internal/shared/infrastructure/logger"

	"github.com/gin-gonic/gin"
)

// ControllerI defines the administrative and user-level operations for session control.
type ControllerI interface {
	Create(ctx *gin.Context)         // Manually create a new session (Internal/Admin).
	Session(ctx *gin.Context)        // Retrieve metadata for a specific session.
	Sessions(ctx *gin.Context)       // List all active sessions for the current user.
	UpdateActivity(ctx *gin.Context) // Heartbeat to extend session validity.
	Delete(ctx *gin.Context)         // Invalidate a specific session by ID.
	RevokeCurrent(ctx *gin.Context)  // Force logout current device session.
	RevokeAll(ctx *gin.Context)      // Force logout from all devices.
	RevokeByDevice(ctx *gin.Context) // Invalidate sessions tied to a particular device fingerprint.
}

// Controller implements session management logic by bridging HTTP requests to usecases.
type Controller struct {
	s *usecase.UseCase
	l logger.Log
}

// New initializes a session controller with access to business logic layer and logging.
func New(s *usecase.UseCase, l logger.Log) ControllerI {
	return &Controller{s: s, l: l}
}
