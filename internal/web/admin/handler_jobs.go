package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ScheduledJobs (GET) - List scheduled jobs
func (h *Handler) ScheduledJobs(ctx *gin.Context) {
	h.servePage(ctx, "jobs/list.html", "Scheduled Jobs", "jobs", map[string]any{
		"Jobs":         []any{},
		"TotalCount":   0,
		"ActiveCount":  0,
		"RunningCount": 0,
		"FailedToday":  0,
	})
}

// JobDetail (GET)
func (h *Handler) JobDetail(ctx *gin.Context) {
	h.servePage(ctx, "jobs/detail.html", "Job Detail", "jobs", map[string]any{
		"Job":        nil,
		"Executions": []any{},
	})
}

// CreateJobPost (POST)
func (h *Handler) CreateJobPost(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"success": false, "message": "Not implemented — jobs backend not connected"})
}

// UpdateJobPost (PUT)
func (h *Handler) UpdateJobPost(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"success": false, "message": "Not implemented — jobs backend not connected"})
}

// DeleteJobPost (DELETE)
func (h *Handler) DeleteJobPost(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"success": false, "message": "Not implemented — jobs backend not connected"})
}

// TriggerJobPost (POST) - Manually trigger a job
func (h *Handler) TriggerJobPost(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"success": false, "message": "Not implemented — jobs backend not connected"})
}
