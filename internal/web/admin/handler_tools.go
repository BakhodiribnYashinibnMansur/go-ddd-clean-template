package admin

import (
	"net/http"
	"os/exec"

	"github.com/gin-gonic/gin"
)

// RunLinter executes golangci-lint and returns JSON report
func (h *Handler) RunLinter(ctx *gin.Context) {
	// Security: This command runs on the server. Ensure only admins can access.
	// Executing golangci-lint
	cmd := exec.Command("golangci-lint", "run", "--out-format", "json")
	// Use current directory (assuming app runs from root)
	// cmd.Dir = "."

	output, err := cmd.CombinedOutput()

	// Check if command failed to start (not just lint issues)
	if err != nil && len(output) == 0 {
		h.l.Errorw("failed to run linter", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to run linter. Is golangci-lint installed?"})
		return
	}

	// golangci-lint returns exit code 1 if issues found, but output is valid JSON
	// We return the raw output directly
	ctx.Data(http.StatusOK, "application/json", output)
}
