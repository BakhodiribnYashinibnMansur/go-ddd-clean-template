package sse

import (
	"fmt"
	"io"
	"time"

	"github.com/gin-gonic/gin"
)

// Handler provides Gin HTTP handlers for SSE streaming.
type Handler struct {
	hub               *Hub
	heartbeatInterval time.Duration
}

// NewHandler creates a new SSE handler.
func NewHandler(hub *Hub, heartbeatInterval time.Duration) *Handler {
	return &Handler{
		hub:               hub,
		heartbeatInterval: heartbeatInterval,
	}
}

// StreamNotifications streams real-time notifications for the authenticated user.
// Channel: notifications:{user_id}
func (h *Handler) StreamNotifications(c *gin.Context) {
	userID, _ := c.Get("user_id")
	channel := fmt.Sprintf("notifications:%s", userID)
	h.stream(c, channel)
}

// StreamAudit streams real-time audit logs (admin only).
// Channel: audit
func (h *Handler) StreamAudit(c *gin.Context) {
	h.stream(c, "audit")
}

// StreamMonitoring streams system errors and metrics (admin only).
// Channel: monitoring
func (h *Handler) StreamMonitoring(c *gin.Context) {
	h.stream(c, "monitoring")
}

// StreamJobProgress streams progress updates for a specific job.
// Channel: jobs:{job_id}
func (h *Handler) StreamJobProgress(c *gin.Context) {
	jobID := c.Param("id")
	channel := fmt.Sprintf("jobs:%s", jobID)
	h.stream(c, channel)
}

// stream is the core SSE loop shared by all endpoints.
func (h *Handler) stream(c *gin.Context, channel string) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	ch := h.hub.Register(channel)
	defer h.hub.Unregister(channel, ch)

	ticker := time.NewTicker(h.heartbeatInterval)
	defer ticker.Stop()

	clientGone := c.Request.Context().Done()

	c.Stream(func(w io.Writer) bool {
		select {
		case <-clientGone:
			return false
		case msg, ok := <-ch:
			if !ok {
				return false
			}
			c.SSEvent(msg.Event, string(msg.Data))
			if msg.ID != "" {
				fmt.Fprintf(w, "id: %s\n", msg.ID)
			}
			return true
		case <-ticker.C:
			fmt.Fprintf(w, ": heartbeat\n\n")
			return true
		}
	})
}
