package database

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"gct/config"
	"gct/internal/domain"
	"gct/internal/repo/persistent/postgres"
	"gct/pkg/logger"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type UseCaseI interface {
	GetActiveSessions(ctx context.Context) ([]*domain.DBSession, error)
	GetSlowQueries(ctx context.Context, limit int) ([]*domain.SlowQuery, error)
	GetTableSizes(ctx context.Context) ([]*domain.TableSize, error)
	GetCacheStats(ctx context.Context) ([]*domain.CacheStats, error)
	GetVacuumStats(ctx context.Context) ([]*domain.VacuumStats, error)
	GetDBMetrics(ctx context.Context) (*domain.DBMetrics, error)
	ExecuteQuery(ctx context.Context, sqlInput string) ([]domain.QueryResult, error)
	ValidateTableName(ctx context.Context, tableName string) error
	GetTableSchema(ctx context.Context, tableName string) (*domain.TableSchema, error)
	GetTableData(ctx context.Context, tableName string, limit, offset int) (*domain.TableData, error)
	InsertRecord(ctx context.Context, tableName string, data map[string]interface{}) error
	UpdateRecord(ctx context.Context, tableName string, pkColumn string, pkValue interface{}, data map[string]interface{}) error
	DeleteRecord(ctx context.Context, tableName string, pkColumn string, pkValue interface{}) error
}

type UseCase struct {
	repo   *postgres.Repo
	logger logger.Log
	cfg    *config.Config
}

func New(repo *postgres.Repo, logger logger.Log, cfg *config.Config) UseCaseI {
	return &UseCase{
		repo:   repo,
		logger: logger,
		cfg:    cfg,
	}
}

// GetActiveSessions returns all active database sessions
func (uc *UseCase) GetActiveSessions(ctx context.Context) ([]*domain.DBSession, error) {
	query := `
		SELECT
			pid,
			usename,
			COALESCE(application_name, '') as application_name,
			client_addr::text,
			state,
			COALESCE(query, '') as query,
			COALESCE(query_start, NOW()) as query_start
		FROM pg_stat_activity
		WHERE state != 'idle'
		  AND pid != pg_backend_pid()
		ORDER BY query_start ASC
	`

	rows, err := uc.repo.DB.Pool.Query(ctx, query)
	if err != nil {
		uc.logger.Errorw("failed to get active sessions", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var sessions []*domain.DBSession
	for rows.Next() {
		var s domain.DBSession
		if err := rows.Scan(&s.PID, &s.Username, &s.ApplicationName, &s.ClientAddr, &s.State, &s.Query, &s.QueryStart); err != nil {
			uc.logger.Warnw("failed to scan session", zap.Error(err))
			continue
		}
		sessions = append(sessions, &s)
	}

	return sessions, nil
}

// GetSlowQueries returns slowest queries from pg_stat_statements
func (uc *UseCase) GetSlowQueries(ctx context.Context, limit int) ([]*domain.SlowQuery, error) {
	query := `
		SELECT
			queryid,
			query,
			calls,
			total_exec_time,
			mean_exec_time,
			rows
		FROM pg_stat_statements
		ORDER BY total_exec_time DESC
		LIMIT $1
	`

	rows, err := uc.repo.DB.Pool.Query(ctx, query, limit)
	if err != nil {
		// pg_stat_statements might not be enabled
		uc.logger.Warnw("pg_stat_statements not available", zap.Error(err))
		return nil, nil
	}
	defer rows.Close()

	var queries []*domain.SlowQuery
	for rows.Next() {
		var q domain.SlowQuery
		if err := rows.Scan(&q.QueryID, &q.Query, &q.Calls, &q.TotalTime, &q.MeanTime, &q.Rows); err != nil {
			uc.logger.Warnw("failed to scan slow query", zap.Error(err))
			continue
		}
		queries = append(queries, &q)
	}

	return queries, nil
}

// GetTableSizes returns table size information
func (uc *UseCase) GetTableSizes(ctx context.Context) ([]*domain.TableSize, error) {
	query := `
		SELECT
			t.relname,
			pg_size_pretty(pg_total_relation_size(t.relid)) AS total_size,
			pg_size_pretty(pg_relation_size(t.relid)) AS table_size,
			pg_size_pretty(pg_indexes_size(t.relid)) AS index_size,
			t.n_live_tup,
			count(a.attname) as column_count
		FROM pg_catalog.pg_stat_user_tables t
		JOIN pg_catalog.pg_attribute a ON a.attrelid = t.relid
		WHERE a.attnum > 0 AND NOT a.attisdropped
		GROUP BY t.relname, t.relid, t.n_live_tup
		ORDER BY pg_total_relation_size(t.relid) DESC
	`

	rows, err := uc.repo.DB.Pool.Query(ctx, query)
	if err != nil {
		uc.logger.Errorw("failed to get table sizes", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var tables []*domain.TableSize
	for rows.Next() {
		var t domain.TableSize
		if err := rows.Scan(&t.TableName, &t.TotalSize, &t.TableSize, &t.IndexSize, &t.RowCount, &t.ColumnCount); err != nil {
			uc.logger.Warnw("failed to scan table size", zap.Error(err))
			continue
		}
		tables = append(tables, &t)
	}

	return tables, nil
}

// GetCacheStats returns cache hit ratio statistics
func (uc *UseCase) GetCacheStats(ctx context.Context) ([]*domain.CacheStats, error) {
	query := `
		SELECT
			relname,
			heap_blks_read,
			heap_blks_hit
		FROM pg_statio_user_tables
		WHERE heap_blks_hit + heap_blks_read > 0
		ORDER BY relname
	`

	rows, err := uc.repo.DB.Pool.Query(ctx, query)
	if err != nil {
		uc.logger.Errorw("failed to get cache stats", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var stats []*domain.CacheStats
	for rows.Next() {
		var s domain.CacheStats
		if err := rows.Scan(&s.TableName, &s.HeapRead, &s.HeapHit); err != nil {
			uc.logger.Warnw("failed to scan cache stats", zap.Error(err))
			continue
		}

		// Calculate cache hit ratio
		total := s.HeapHit + s.HeapRead
		if total > 0 {
			s.CacheHitRatio = float64(s.HeapHit) * 100.0 / float64(total)
		}

		stats = append(stats, &s)
	}

	return stats, nil
}

// GetVacuumStats returns vacuum statistics
func (uc *UseCase) GetVacuumStats(ctx context.Context) ([]*domain.VacuumStats, error) {
	query := `
		SELECT
			relname,
			n_live_tup,
			n_dead_tup,
			last_vacuum,
			last_autovacuum
		FROM pg_stat_user_tables
		ORDER BY n_dead_tup DESC
	`

	rows, err := uc.repo.DB.Pool.Query(ctx, query)
	if err != nil {
		uc.logger.Errorw("failed to get vacuum stats", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var stats []*domain.VacuumStats
	for rows.Next() {
		var s domain.VacuumStats
		if err := rows.Scan(&s.TableName, &s.LiveTuples, &s.DeadTuples, &s.LastVacuum, &s.LastAutoVacuum); err != nil {
			uc.logger.Warnw("failed to scan vacuum stats", zap.Error(err))
			continue
		}
		stats = append(stats, &s)
	}

	return stats, nil
}

// GetDBMetrics returns overall database metrics
func (uc *UseCase) GetDBMetrics(ctx context.Context) (*domain.DBMetrics, error) {
	var metrics domain.DBMetrics

	// Get connection counts
	err := uc.repo.DB.Pool.QueryRow(ctx, `
		SELECT
			COUNT(CASE WHEN state = 'active' THEN 1 END) as active,
			COUNT(CASE WHEN state = 'idle' THEN 1 END) as idle,
			COUNT(*) as total
		FROM pg_stat_activity
	`).Scan(&metrics.ActiveConnections, &metrics.IdleConnections, &metrics.TotalConnections)

	if err != nil {
		uc.logger.Errorw("failed to get connection counts", zap.Error(err))
		return nil, err
	}

	// Get database size
	err = uc.repo.DB.Pool.QueryRow(ctx, `
		SELECT pg_size_pretty(pg_database_size(current_database()))
	`).Scan(&metrics.DatabaseSize)

	if err != nil {
		uc.logger.Warnw("failed to get database size", zap.Error(err))
	}

	// Get table count
	err = uc.repo.DB.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public'
	`).Scan(&metrics.TableCount)

	if err != nil {
		uc.logger.Warnw("failed to get table count", zap.Error(err))
	}

	// Get overall cache hit ratio
	err = uc.repo.DB.Pool.QueryRow(ctx, `
		SELECT
			ROUND(
				SUM(heap_blks_hit) * 100.0 / NULLIF(SUM(heap_blks_hit + heap_blks_read), 0),
				2
			)
		FROM pg_statio_user_tables
	`).Scan(&metrics.CacheHitRatio)

	if err != nil {
		uc.logger.Warnw("failed to get cache hit ratio", zap.Error(err))
	}

	// Get total dead tuples
	err = uc.repo.DB.Pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(n_dead_tup), 0) FROM pg_stat_user_tables
	`).Scan(&metrics.DeadTuples)

	if err != nil {
		uc.logger.Warnw("failed to get dead tuples count", zap.Error(err))
	}

	return &metrics, nil
}

// ExecuteQuery executes raw SQL query/queries (read-only)
func (uc *UseCase) ExecuteQuery(ctx context.Context, sqlInput string) ([]domain.QueryResult, error) {
	// 1. Split queries by semicolon
	// This is a naive split, but sufficient for simple admin usage.
	// For production-grade splitting (handling semicolons in quotes), a parser is needed.
	rawQueries := strings.Split(sqlInput, ";")
	var results []domain.QueryResult

	for _, q := range rawQueries {
		q = strings.TrimSpace(q)
		if q == "" {
			continue
		}

		startTime := time.Now()
		result := domain.QueryResult{
			SQL:       q,
			Timestamp: startTime,
		}

		// Security: Check based on Environment
		upperQ := strings.ToUpper(q)
		isSelect := strings.HasPrefix(upperQ, "SELECT") || strings.HasPrefix(upperQ, "EXPLAIN")

		if uc.cfg.IsProd() {
			// PROD: Only allow SELECT
			if !isSelect {
				result.Error = "Production Mode: Only SELECT and EXPLAIN queries are allowed"
				result.Duration = time.Since(startTime).String()
				results = append(results, result)
				continue
			}
		} else {
			// Non-PROD: Block destructive commands (DROP, TRUNCATE)
			// Note: This is a basic check. A real parser would be safer, but this covers accidental runs.
			if strings.HasPrefix(upperQ, "DROP") || strings.HasPrefix(upperQ, "TRUNCATE") {
				result.Error = fmt.Sprintf("Restricted: %s statements are not allowed in this environment", strings.Split(upperQ, " ")[0])
				result.Duration = time.Since(startTime).String()
				results = append(results, result)
				continue
			}
		}

		rows, err := uc.repo.DB.Pool.Query(ctx, q)
		if err != nil {
			result.Error = err.Error()
			result.Duration = time.Since(startTime).String()
			results = append(results, result)
			continue
		}

		// Get columns
		fieldDescriptions := rows.FieldDescriptions()
		columns := make([]string, len(fieldDescriptions))
		for i, fd := range fieldDescriptions {
			columns[i] = string(fd.Name)
		}
		result.Columns = columns

		// Scan results
		var rowData []map[string]interface{}
		for rows.Next() {
			values, err := rows.Values()
			if err != nil {
				uc.logger.Warnw("failed to scan row", zap.Error(err))
				continue
			}

			row := make(map[string]interface{})
			for i, col := range columns {
				// Handle byte arrays (UUID) conversion for display
				if v, ok := values[i].([16]uint8); ok {
					row[col] = fmt.Sprintf("%x-%x-%x-%x-%x", v[0:4], v[4:6], v[6:8], v[8:10], v[10:16])
				} else {
					row[col] = values[i]
				}
			}
			rowData = append(rowData, row)
		}
		rows.Close()

		result.Rows = rowData
		result.RowCount = len(rowData)
		result.Duration = time.Since(startTime).String()
		results = append(results, result)
	}

	return results, nil
}

// ValidateTableName validates that table exists in public schema
func (uc *UseCase) ValidateTableName(ctx context.Context, tableName string) error {
	// Regex validation for safety
	matched, _ := regexp.MatchString(`^[a-zA-Z_][a-zA-Z0-9_]*$`, tableName)
	if !matched {
		return fmt.Errorf("invalid table name format")
	}

	// Check if table exists
	var exists bool
	err := uc.repo.DB.Pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables 
			WHERE table_schema = 'public' AND table_name = $1
		)
	`, tableName).Scan(&exists)

	if err != nil {
		return fmt.Errorf("failed to validate table: %w", err)
	}

	if !exists {
		return fmt.Errorf("table '%s' does not exist", tableName)
	}

	return nil
}

// GetTableSchema returns table structure
func (uc *UseCase) GetTableSchema(ctx context.Context, tableName string) (*domain.TableSchema, error) {
	if err := uc.ValidateTableName(ctx, tableName); err != nil {
		return nil, err
	}

	schema := &domain.TableSchema{
		TableName: tableName,
	}

	// Get columns
	query := `
		SELECT 
			column_name,
			data_type,
			is_nullable,
			column_default
		FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = $1
		ORDER BY ordinal_position
	`

	rows, err := uc.repo.DB.Pool.Query(ctx, query, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var col domain.ColumnInfo
		var defaultVal *string
		var isNullable string

		if err := rows.Scan(&col.Name, &col.DataType, &isNullable, &defaultVal); err != nil {
			continue
		}

		col.IsNullable = isNullable == "YES"
		col.DefaultValue = defaultVal
		schema.Columns = append(schema.Columns, col)
	}

	// Get primary key
	err = uc.repo.DB.Pool.QueryRow(ctx, `
		SELECT a.attname
		FROM pg_index i
		JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY(i.indkey)
		WHERE i.indrelid = $1::regclass AND i.indisprimary
	`, tableName).Scan(&schema.PrimaryKey)

	if err != nil && err != pgx.ErrNoRows {
		uc.logger.Warnw("failed to get primary key", zap.Error(err))
	}

	// Mark primary key column
	for i := range schema.Columns {
		if schema.Columns[i].Name == schema.PrimaryKey {
			schema.Columns[i].IsPrimaryKey = true
		}
	}

	// Get row count
	err = uc.repo.DB.Pool.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).Scan(&schema.RowCount)
	if err != nil {
		uc.logger.Warnw("failed to get row count", zap.Error(err))
	}

	return schema, nil
}

// GetTableData returns paginated table data
func (uc *UseCase) GetTableData(ctx context.Context, tableName string, limit, offset int) (*domain.TableData, error) {
	if err := uc.ValidateTableName(ctx, tableName); err != nil {
		return nil, err
	}

	data := &domain.TableData{
		TableName: tableName,
	}

	// Get total count
	err := uc.repo.DB.Pool.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).Scan(&data.Total)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	// Get data
	query := fmt.Sprintf("SELECT * FROM %s LIMIT $1 OFFSET $2", tableName)
	rows, err := uc.repo.DB.Pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query table: %w", err)
	}
	defer rows.Close()

	// Get column names
	fieldDescriptions := rows.FieldDescriptions()
	data.Columns = make([]string, len(fieldDescriptions))
	for i, fd := range fieldDescriptions {
		data.Columns[i] = string(fd.Name)
	}

	// Scan rows
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			uc.logger.Warnw("failed to scan row", zap.Error(err))
			continue
		}

		row := make(map[string]interface{})
		for i, col := range data.Columns {
			row[col] = values[i]
		}
		data.Rows = append(data.Rows, row)
	}

	return data, nil
}

// InsertRecord inserts a new record
func (uc *UseCase) InsertRecord(ctx context.Context, tableName string, data map[string]interface{}) error {
	if err := uc.ValidateTableName(ctx, tableName); err != nil {
		return err
	}

	// Build INSERT query
	var columns []string
	var placeholders []string
	var values []interface{}
	i := 1

	for col, val := range data {
		columns = append(columns, col)
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		values = append(values, val)
		i++
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	_, err := uc.repo.DB.Pool.Exec(ctx, query, values...)
	if err != nil {
		uc.logger.Errorw("failed to insert record", zap.Error(err), zap.String("table", tableName))
		return fmt.Errorf("failed to insert record: %w", err)
	}

	return nil
}

// UpdateRecord updates existing record
func (uc *UseCase) UpdateRecord(ctx context.Context, tableName string, pkColumn string, pkValue interface{}, data map[string]interface{}) error {
	if err := uc.ValidateTableName(ctx, tableName); err != nil {
		return err
	}

	// Build UPDATE query
	var setParts []string
	var values []interface{}
	i := 1

	for col, val := range data {
		if col == pkColumn {
			continue // Don't update primary key
		}
		setParts = append(setParts, fmt.Sprintf("%s = $%d", col, i))
		values = append(values, val)
		i++
	}

	values = append(values, pkValue)

	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE %s = $%d",
		tableName,
		strings.Join(setParts, ", "),
		pkColumn,
		i,
	)

	_, err := uc.repo.DB.Pool.Exec(ctx, query, values...)
	if err != nil {
		uc.logger.Errorw("failed to update record", zap.Error(err), zap.String("table", tableName))
		return fmt.Errorf("failed to update record: %w", err)
	}

	return nil
}

// cleanQuery removes comments and whitespace for security check
func cleanQuery(q string) string {
	// Remove block comments /* ... */
	reBlock := regexp.MustCompile(`/\*.*?\*/`)
	q = reBlock.ReplaceAllString(q, "")

	// Remove line comments -- ...
	reLine := regexp.MustCompile(`--.*`)
	q = reLine.ReplaceAllString(q, "")

	// Normalize spaces
	q = strings.TrimSpace(q)

	// Remove non-alphanumeric prefix (e.g. ;)
	// This is a simple cleanup, not a full parser
	for len(q) > 0 && !isAlpha(q[0]) {
		q = q[1:]
		q = strings.TrimSpace(q)
	}

	return q
}

func isAlpha(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

// DeleteRecord deletes a record
func (uc *UseCase) DeleteRecord(ctx context.Context, tableName string, pkColumn string, pkValue interface{}) error {
	if err := uc.ValidateTableName(ctx, tableName); err != nil {
		return err
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE %s = $1", tableName, pkColumn)

	_, err := uc.repo.DB.Pool.Exec(ctx, query, pkValue)
	if err != nil {
		uc.logger.Errorw("failed to delete record", zap.Error(err), zap.String("table", tableName))
		return fmt.Errorf("failed to delete record: %w", err)
	}

	return nil
}
