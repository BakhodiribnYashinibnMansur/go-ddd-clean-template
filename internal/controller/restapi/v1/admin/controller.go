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

type Controller struct {
	l logger.Log
}

func New(l logger.Log) *Controller {
	return &Controller{
		l: l,
	}
}

func (c *Controller) Register(r *gin.RouterGroup) {
	g := r.Group("/admin")
	g.POST("/linter/run", c.RunLinter)
}

type LinterResponse struct {
	TotalIssues    int               `json:"totalIssues"`
	TotalLinters   int               `json:"totalLinters"`
	TotalFiles     int               `json:"totalFiles"`
	IssuesByLinter map[string]int    `json:"issuesByLinter"`
	Report         string            `json:"report"`
	ReportPaths    map[string]string `json:"reportPaths"`
}

func (c *Controller) RunLinter(ctx *gin.Context) {
	// Get project root directory
	projectRoot, err := os.Getwd()
	if err != nil {
		c.l.Errorw("failed to get working directory", "error", err)
		response.ControllerResponse(ctx, http.StatusInternalServerError, "Failed to get working directory", nil, false)
		return
	}

	// Create report directory if it doesn't exist
	reportDir := filepath.Join(projectRoot, "docs", "report", "linter")
	if err := os.MkdirAll(reportDir, 0o755); err != nil {
		c.l.Errorw("failed to create report directory", "error", err)
		response.ControllerResponse(ctx, http.StatusInternalServerError, "Failed to create report directory", nil, false)
		return
	}

	// Define report paths
	reportTextPath := filepath.Join(reportDir, "report.txt")
	reportJSONPath := filepath.Join(reportDir, "report.json")
	reportHTMLPath := filepath.Join(reportDir, "report.html")

	// Run golangci-lint with timeout
	// The .golangci.yml config will automatically generate all formats
	cmdCtx, cancel := context.WithTimeout(ctx.Request.Context(), 3*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, "golangci-lint", "run")
	cmd.Dir = projectRoot

	// Capture both stdout and stderr
	output, err := cmd.CombinedOutput()

	// golangci-lint returns non-zero exit code when issues are found, which is expected
	// Only treat it as error if context deadline exceeded
	if err != nil && errors.Is(cmdCtx.Err(), context.DeadlineExceeded) {
		c.l.Errorw("linter command timeout", "error", err)
		response.ControllerResponse(ctx, http.StatusRequestTimeout, "Linter execution timeout", nil, false)
		return
	}

	// Read the generated text report for parsing
	textContent, err := os.ReadFile(reportTextPath)
	if err != nil {
		c.l.Warnw("failed to read text report", "error", err)
		textContent = output // fallback to command output
	}

	// Parse the report
	linterData := parseReport(string(textContent))
	linterData.ReportPaths = map[string]string{
		"text": reportTextPath,
		"json": reportJSONPath,
		"html": reportHTMLPath,
	}

	response.ControllerResponse(ctx, http.StatusOK, "Linter completed successfully", linterData, true)
}

func parseReport(report string) *LinterResponse {
	result := &LinterResponse{
		IssuesByLinter: make(map[string]int),
		Report:         report,
	}

	if report == "" {
		return result
	}

	// Parse the summary at the end of the report
	// Example: "2616 issues:"
	issuesPattern := regexp.MustCompile(`(\d+) issues?:`)
	if matches := issuesPattern.FindStringSubmatch(report); len(matches) > 1 {
		_, _ = fmt.Sscanf(matches[1], "%d", &result.TotalIssues)
	}

	// Parse linter counts from summary
	// Example: "* bodyclose: 5"
	scanner := bufio.NewScanner(strings.NewReader(report))
	inSummary := false
	filesMap := make(map[string]bool)

	for scanner.Scan() {
		line := scanner.Text()

		// Detect summary section
		if strings.Contains(line, "issues:") {
			inSummary = true
			continue
		}

		if inSummary {
			// Parse linter statistics
			// Format: "* lintername: count"
			linterPattern := regexp.MustCompile(`^\* ([a-z0-9]+): (\d+)`)
			if matches := linterPattern.FindStringSubmatch(line); len(matches) > 2 {
				linterName := matches[1]
				count := 0
				_, _ = fmt.Sscanf(matches[2], "%d", &count)
				result.IssuesByLinter[linterName] = count
			}
		} else {
			// Parse file paths from issue lines
			// Format: "path/to/file.go:123:45: message (linter)"
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
