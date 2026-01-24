// Package asynqController provides HTTP endpoints to trigger background tasks
// using the Asynq task queue system.
package asynqController

import (
	"net/http"

	"gct/pkg/asynq"
	"gct/pkg/logger"

	"github.com/gin-gonic/gin"
)

// Controller facilitates the enqueuing of various asynchronous tasks via HTTP requests.
type Controller struct {
	asynqClient *asynq.Client
	log         logger.Log
}

// NewController initializes a new queue controller with an Asynq client and a logger.
func NewController(asynqClient *asynq.Client, log logger.Log) *Controller {
	return &Controller{
		asynqClient: asynqClient,
		log:         log,
	}
}

// SendTestEmail godoc
// @Summary Send test email
// @Description Enqueue a test email task into the queue system
// @Tags asynq
// @Accept json
// @Produce json
// @Param request body EmailRequest true "Email details: recipient, subject, and body"
// @Success 200 {object} map[string]interface{} "Returns task ID and queue name on success"
// @Failure 400 {object} map[string]interface{} "Malformed request body or invalid email format"
// @Failure 500 {object} map[string]interface{} "Failed to connect to the task queue"
// @Router /api/v1/asynq/email/test [post]
func (ctrl *Controller) SendTestEmail(c *gin.Context) {
	var req EmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payload := asynq.EmailPayload{
		To:      req.To,
		Subject: req.Subject,
		Body:    req.Body,
		Data:    req.Data,
	}

	// Schedule the email task using the Asynq client
	info, err := ctrl.asynqClient.EnqueueEmail(c.Request.Context(), asynq.TypeEmailWelcome, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enqueue email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Email task enqueued successfully",
		"task_id": info.ID,
		"queue":   info.Queue,
	})
}

// SendTestNotification godoc
// @Summary Send test notification
// @Description Enqueue a push notification task for a specific user
// @Tags asynq
// @Accept json
// @Produce json
// @Param request body NotificationRequest true "Notification details: userID, title, and message"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/asynq/notification/test [post]
func (ctrl *Controller) SendTestNotification(c *gin.Context) {
	var req NotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payload := asynq.NotificationPayload{
		UserID:  req.UserID,
		Title:   req.Title,
		Message: req.Message,
		Data:    req.Data,
	}

	// Schedule the notification task
	info, err := ctrl.asynqClient.EnqueueNotification(c.Request.Context(), asynq.TypePushNotification, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enqueue notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Notification task enqueued successfully",
		"task_id": info.ID,
		"queue":   info.Queue,
	})
}

// SeedDatabase godoc
// @Summary Seed database
// @Description Enqueue a task to populate the database with mock data for testing
// @Tags asynq
// @Accept json
// @Produce json
// @Param request body SeedRequest true "Seeding parameters: counts for users, roles, etc."
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/asynq/seed [post]
func (ctrl *Controller) SeedDatabase(c *gin.Context) {
	var req SeedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payload := asynq.SeedPayload{
		UsersCount:       req.UsersCount,
		RolesCount:       req.RolesCount,
		PermissionsCount: req.PermissionsCount,
		Seed:             req.Seed,
		ClearData:        req.ClearData,
	}

	// Schedule the seeding task; this can be time-consuming, so it's best handled in a queue
	info, err := ctrl.asynqClient.EnqueueSeed(c.Request.Context(), payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enqueue seed task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Seeding task enqueued successfully",
		"task_id": info.ID,
		"queue":   info.Queue,
	})
}

// ============================================================================
// Request Models
// ============================================================================

// EmailRequest represents the JSON payload for triggering an email task.
type EmailRequest struct {
	To      string            `json:"to" binding:"required,email"`
	Subject string            `json:"subject" binding:"required"`
	Body    string            `json:"body" binding:"required"`
	Data    map[string]string `json:"data,omitempty"`
}

// NotificationRequest represents the JSON payload for triggering a notification task.
type NotificationRequest struct {
	UserID  string            `json:"user_id" binding:"required"`
	Title   string            `json:"title" binding:"required"`
	Message string            `json:"message" binding:"required"`
	Data    map[string]string `json:"data,omitempty"`
}

// SeedRequest represents the configuration parameters for a database seeding task.
type SeedRequest struct {
	UsersCount       int   `json:"users_count"`
	RolesCount       int   `json:"roles_count"`
	PermissionsCount int   `json:"permissions_count"`
	Seed             int64 `json:"seed"`
	ClearData        bool  `json:"clear_data"`
}
