package domain

import "time"

// DBSession represents an active database session
type DBSession struct {
	PID             int       `db:"pid"`
	Username        string    `db:"usename"`
	ApplicationName string    `db:"application_name"`
	ClientAddr      *string   `db:"client_addr"`
	State           string    `db:"state"`
	Query           string    `db:"query"`
	QueryStart      time.Time `db:"query_start"`
	Duration        string    // Calculated
}

// SlowQuery represents a slow query from pg_stat_statements
type SlowQuery struct {
	QueryID   int64   `db:"queryid"`
	Query     string  `db:"query"`
	Calls     int64   `db:"calls"`
	TotalTime float64 `db:"total_exec_time"`
	MeanTime  float64 `db:"mean_exec_time"`
	Rows      int64   `db:"rows"`
}

// TableSize represents table size information
type TableSize struct {
	TableName   string `db:"relname"`
	Schema      string `db:"schema_name"`
	Type        string `db:"table_type"`
	TotalSize   string `db:"total_size"`
	TableSize   string `db:"table_size"`
	IndexSize   string `db:"index_size"`
	RowCount    int64  `db:"n_live_tup"`
	ColumnCount int    `db:"column_count"`
}

// CacheStats represents cache hit ratio statistics
type CacheStats struct {
	TableName     string  `db:"relname"`
	HeapRead      int64   `db:"heap_blks_read"`
	HeapHit       int64   `db:"heap_blks_hit"`
	CacheHitRatio float64 // Calculated
}

// VacuumStats represents vacuum statistics
type VacuumStats struct {
	TableName      string     `db:"relname"`
	LiveTuples     int64      `db:"n_live_tup"`
	DeadTuples     int64      `db:"n_dead_tup"`
	LastVacuum     *time.Time `db:"last_vacuum"`
	LastAutoVacuum *time.Time `db:"last_autovacuum"`
}

// DBMetrics aggregates all database metrics
type DBMetrics struct {
	ActiveConnections int
	IdleConnections   int
	TotalConnections  int
	DatabaseSize      string
	TableCount        int
	CacheHitRatio     float64
	DeadTuples        int64
}

// TableSchema represents table structure
type TableSchema struct {
	TableName  string
	Columns    []ColumnInfo
	PrimaryKey string
	RowCount   int64
}

// ColumnInfo represents column metadata
type ColumnInfo struct {
	Name         string
	DataType     string
	IsNullable   bool
	DefaultValue *string
	IsPrimaryKey bool
}

// TableData represents paginated table data
type TableData struct {
	TableName string
	Columns   []string
	Rows      []map[string]interface{}
	Total     int64
}

// QueryResult represents result of a single SQL query execution
type QueryResult struct {
	SQL       string                   `json:"sql"`
	Columns   []string                 `json:"columns,omitempty"`
	Rows      []map[string]interface{} `json:"rows,omitempty"`
	RowCount  int                      `json:"row_count"`
	Duration  string                   `json:"duration"`
	Error     string                   `json:"error,omitempty"`
	Timestamp time.Time                `json:"timestamp"`
}
