// Package admin provides administrative endpoints, such as system health checks and linter report generation.
package admin

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gct/internal/controller/restapi/response"
	"gct/pkg/logger"

	"github.com/gin-gonic/gin"
)

// Controller handles administrative tasks for the application.
type Controller struct {
	l logger.Log
}

// New instantiates a new admin controller with the provided logger.
func New(l logger.Log) *Controller {
	return &Controller{
		l: l,
	}
}

// Register adds admin-specific routes to the provided router group.
func (c *Controller) Register(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	g := r.Group("admin")
	g.Use(authMiddleware)
	{
		g.POST("/linter/run", c.RunLinter)
	}
}

// LinterResponse defines the structure of the linter execution results,
// including summarized statistics and paths to detailed report files.
type LinterResponse struct {
	TotalIssues    int               `json:"totalIssues"`
	TotalLinters   int               `json:"totalLinters"`
	TotalFiles     int               `json:"totalFiles"`
	IssuesByLinter map[string]int    `json:"issuesByLinter"`
	Report         string            `json:"report"`
	ReportPaths    map[string]string `json:"reportPaths"`
}

// RunLinter godoc
// @Summary     Run linter
// @Description Triggers asynchronous execution of golangci-lint
// @Tags        admin
// @Accept      json
// @Produce     json
// @Success     200 {object} response.SuccessResponse
// @Failure     400 {object} response.ErrorResponse
// @Failure     401 {object} response.ErrorResponse
// @Failure     403 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Security    BearerAuth
// @Router      /admin/linter/run [post]
func (c *Controller) RunLinter(ctx *gin.Context) {
	// Identify project root to determine where to store reports
	projectRoot, err := os.Getwd()
	if err != nil {
		c.l.Errorw("failed to get working directory", "error", err)
		response.ControllerResponse(ctx, http.StatusInternalServerError, "Failed to get working directory", nil, false)
		return
	}

	// Ensure the report directory exists locally
	reportDir := filepath.Join(projectRoot, "docs", "report", "linter")
	if err := os.MkdirAll(reportDir, 0o755); err != nil {
		c.l.Errorw("failed to create report directory", "error", err)
		response.ControllerResponse(ctx, http.StatusInternalServerError, "Failed to create report directory", nil, false)
		return
	}

	// Paths for multi-format report output
	reportTextPath := filepath.Join(reportDir, "report.txt")
	reportJSONPath := filepath.Join(reportDir, "report.json")
	reportHTMLPath := filepath.Join(reportDir, "report.html")

	// Execute linter with a strict timeout to prevent blocking resources
	cmdCtx, cancel := context.WithTimeout(ctx.Request.Context(), 3*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, "golangci-lint", "run")
	cmd.Dir = projectRoot

	// Capture all output for parsing and debugging
	output, err := cmd.CombinedOutput()

	// Handle execution timeouts explicitly
	if err != nil && errors.Is(cmdCtx.Err(), context.DeadlineExceeded) {
		c.l.Errorw("linter command timeout", "error", err)
		response.ControllerResponse(ctx, http.StatusRequestTimeout, "Linter execution timeout", nil, false)
		return
	}

	// Read generated textual report to extract summary stats
	textContent, err := os.ReadFile(reportTextPath)
	if err != nil {
		c.l.Warnw("failed to read text report", "error", err)
		textContent = output // Fallback to raw process output if file read fails
	}

	// Parse textual output into structured response data
	linterData := parseReport(string(textContent))
	linterData.ReportPaths = map[string]string{
		"text": reportTextPath,
		"json": reportJSONPath,
		"html": reportHTMLPath,
	}

	response.ControllerResponse(ctx, http.StatusOK, "Linter completed successfully", linterData, true)
}

// parseReport uses regex to extract issue counts, file counts, and linter-specific stats
// from the raw textual output of golangci-lint.
func parseReport(report string) *LinterResponse {
	result := &LinterResponse{
		IssuesByLinter: make(map[string]int),
		Report:         report,
	}

	if report == "" {
		return result
	}

	// Extract total issue count from the summary line
	issuesPattern := regexp.MustCompile(`(\d+) issues?:`)
	if matches := issuesPattern.FindStringSubmatch(report); len(matches) > 1 {
		_, _ = fmt.Sscanf(matches[1], "%d", &result.TotalIssues)
	}

	// Iterate through lines to map issues back to specific linters and files
	scanner := bufio.NewScanner(strings.NewReader(report))
	inSummary := false
	filesMap := make(map[string]bool)

	for scanner.Scan() {
		line := scanner.Text()

		// Summary section usually follows the detailed findings
		if strings.Contains(line, "issues:") {
			inSummary = true
			continue
		}

		if inSummary {
			// Extract counts formatted as "* linter: count"
			linterPattern := regexp.MustCompile(`^\* ([a-z0-9]+): (\d+)`)
			if matches := linterPattern.FindStringSubmatch(line); len(matches) > 2 {
				linterName := matches[1]
				count := 0
				_, _ = fmt.Sscanf(matches[2], "%d", &count)
				result.IssuesByLinter[linterName] = count
			}
		} else {
			// Extract file paths from issue lines to determine unique file count
			filePattern := regexp.MustCompile(`^([^:]+\.go):\d+:\d+:`)
			if matches := filePattern.FindStringSubmatch(line); len(matches) > 1 {
				filesMap[matches[1]] = true
			}
		}
	}

	result.TotalLinters = len(result.IssuesByLinter)
	result.TotalFiles = len(filesMap)

	return result
}
