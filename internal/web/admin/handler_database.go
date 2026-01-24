package admin

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// DatabaseMonitoring shows database monitoring dashboard
func (h *Handler) DatabaseMonitoring(ctx *gin.Context) {
	metrics, err := h.uc.Database.GetDBMetrics(ctx.Request.Context())
	if err != nil {
		h.l.Errorw("failed to get DB metrics", "error", err)
		h.servePage(ctx, "coming_soon.html", "Database Monitoring", "database_monitoring", nil)
		return
	}

	sessions, err := h.uc.Database.GetActiveSessions(ctx.Request.Context())
	if err != nil {
		h.l.Errorw("failed to get active sessions", "error", err)
	}

	slowQueries, err := h.uc.Database.GetSlowQueries(ctx.Request.Context(), 10)
	if err != nil {
		h.l.Warnw("failed to get slow queries", "error", err)
	}

	cacheStats, err := h.uc.Database.GetCacheStats(ctx.Request.Context())
	if err != nil {
		h.l.Errorw("failed to get cache stats", "error", err)
	}

	vacuumStats, err := h.uc.Database.GetVacuumStats(ctx.Request.Context())
	if err != nil {
		h.l.Errorw("failed to get vacuum stats", "error", err)
	}

	h.servePage(ctx, "db_monitoring.html", "Database Monitoring", "database_monitoring", map[string]any{
		"Metrics":     metrics,
		"Sessions":    sessions,
		"SlowQueries": slowQueries,
		"CacheStats":  cacheStats,
		"VacuumStats": vacuumStats,
	})
}

// DatabaseTables shows all database tables with size info
func (h *Handler) DatabaseTables(ctx *gin.Context) {
	tables, err := h.uc.Database.GetTableSizes(ctx.Request.Context())
	if err != nil {
		h.l.Errorw("failed to get table sizes", "error", err)
		h.servePage(ctx, "coming_soon.html", "Database Tables", "database_tables", nil)
		return
	}

	h.servePage(ctx, "db_tables.html", "Database Tables", "database_tables", map[string]any{
		"Tables": tables,
	})
}

// GetTablesAPI returns list of database tables
func (h *Handler) GetTablesAPI(ctx *gin.Context) {
	tables, err := h.uc.Database.GetTableSizes(ctx.Request.Context())
	if err != nil {
		h.l.Errorw("failed to get tables", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tables"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"success": true, "tables": tables})
}

// TableData shows table data viewer page
func (h *Handler) TableData(ctx *gin.Context) {
	tableName := ctx.Param("name")

	schema, err := h.uc.Database.GetTableSchema(ctx.Request.Context(), tableName)
	if err != nil {
		h.l.Errorw("failed to get table schema", "error", err, "table", tableName)
		h.servePage(ctx, "coming_soon.html", "Coming Soon", "database_tables", nil)
		return
	}

	h.servePage(ctx, "db_table_data.html", "Table: "+tableName, "database_tables", map[string]any{
		"TableName": tableName,
		"Schema":    schema,
	})
}

// GetTableDataAPI returns table data as JSON
func (h *Handler) GetTableDataAPI(ctx *gin.Context) {
	tableName := ctx.Param("name")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "50"))

	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	data, err := h.uc.Database.GetTableData(ctx.Request.Context(), tableName, limit, offset)
	if err != nil {
		h.l.Errorw("failed to get table data", "error", err, "table", tableName)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
		"page":    page,
		"limit":   limit,
	})
}

// CreateRecord handles record creation
func (h *Handler) CreateRecord(ctx *gin.Context) {
	tableName := ctx.Param("name")

	var data map[string]interface{}
	if err := ctx.BindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := h.uc.Database.InsertRecord(ctx.Request.Context(), tableName, data); err != nil {
		h.l.Warnw("failed to create record", "error", err, "table", tableName)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true, "message": "Record created successfully"})
}

// UpdateRecord handles record update
func (h *Handler) UpdateRecord(ctx *gin.Context) {
	tableName := ctx.Param("name")
	pkValue := ctx.Param("id")

	var req struct {
		PrimaryKey string                 `json:"primary_key"`
		Data       map[string]interface{} `json:"data"`
	}

	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := h.uc.Database.UpdateRecord(ctx.Request.Context(), tableName, req.PrimaryKey, pkValue, req.Data); err != nil {
		h.l.Warnw("failed to update record", "error", err, "table", tableName)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true, "message": "Record updated successfully"})
}

// DeleteRecord handles record deletion
func (h *Handler) DeleteRecord(ctx *gin.Context) {
	tableName := ctx.Param("name")
	pkColumn := ctx.Query("pk_column")
	pkValue := ctx.Param("id")

	if pkColumn == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Primary key column required"})
		return
	}

	if err := h.uc.Database.DeleteRecord(ctx.Request.Context(), tableName, pkColumn, pkValue); err != nil {
		h.l.Warnw("failed to delete record", "error", err, "table", tableName)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true, "message": "Record deleted successfully"})
}

// SQLEditor shows SQL editor page
func (h *Handler) SQLEditor(ctx *gin.Context) {
	h.servePage(ctx, "sql_editor.html", "SQL Editor", "sql_editor", map[string]any{})
}

// ExecuteSQL executes SQL query from editor
func (h *Handler) ExecuteSQL(ctx *gin.Context) {
	var req struct {
		SQL string `json:"sql" binding:"required"`
	}

	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	results, err := h.uc.Database.ExecuteQuery(ctx.Request.Context(), req.SQL)
	if err != nil {
		h.l.Warnw("SQL execution failed", "error", err, "sql", req.SQL)
		ctx.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"results": results,
		"count":   len(results),
	})
}
