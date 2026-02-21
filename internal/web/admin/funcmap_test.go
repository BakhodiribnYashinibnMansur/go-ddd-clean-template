package admin

import (
	"html/template"
	"testing"
)

func getFunc(name string) any {
	return templateFuncs[name]
}

func TestFormatBytes(t *testing.T) {
	fn := getFunc("formatBytes").(func(any) string)
	tests := []struct {
		input    any
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
		{int64(1099511627776), "1.0 TB"},
	}
	for _, tt := range tests {
		result := fn(tt.input)
		if result != tt.expected {
			t.Errorf("formatBytes(%v) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestFormatPercent(t *testing.T) {
	fn := getFunc("formatPercent").(func(any) string)
	if got := fn(99.5); got != "99.50%" {
		t.Errorf("formatPercent(99.5) = %q, want %q", got, "99.50%")
	}
	if got := fn(0); got != "0.00%" {
		t.Errorf("formatPercent(0) = %q, want %q", got, "0.00%")
	}
}

func TestFormatNumber(t *testing.T) {
	fn := getFunc("formatNumber").(func(any) string)
	tests := []struct {
		input    any
		expected string
	}{
		{0, "0"},
		{999, "999"},
		{1000, "1,000"},
		{1234567, "1,234,567"},
		{int64(10000000), "10,000,000"},
	}
	for _, tt := range tests {
		result := fn(tt.input)
		if result != tt.expected {
			t.Errorf("formatNumber(%v) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestFormatDurationMs(t *testing.T) {
	fn := getFunc("formatDurationMs").(func(any) string)
	tests := []struct {
		input    any
		expected string
	}{
		{0.5, "500.00 µs"},
		{1.0, "1.0 ms"},
		{500.0, "500.0 ms"},
		{1500.0, "1.50 s"},
		{90000.0, "1.5 min"},
		{7200000.0, "2.0 h"},
	}
	for _, tt := range tests {
		result := fn(tt.input)
		if result != tt.expected {
			t.Errorf("formatDurationMs(%v) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestPercentage(t *testing.T) {
	fn := getFunc("percentage").(func(any, any) float64)
	if got := fn(50, 100); got != 50.0 {
		t.Errorf("percentage(50, 100) = %f, want 50.0", got)
	}
	if got := fn(0, 0); got != 0.0 {
		t.Errorf("percentage(0, 0) = %f, want 0.0", got)
	}
	if got := fn(1, 3); got < 33.33 || got > 33.34 {
		t.Errorf("percentage(1, 3) = %f, want ~33.33", got)
	}
}

func TestSizeBarWidth(t *testing.T) {
	fn := getFunc("sizeBarWidth").(func(any, any) int64)
	if got := fn(50, 100); got != 50 {
		t.Errorf("sizeBarWidth(50, 100) = %d, want 50", got)
	}
	if got := fn(0, 100); got != 0 {
		t.Errorf("sizeBarWidth(0, 100) = %d, want 0", got)
	}
	if got := fn(1, 0); got != 0 {
		t.Errorf("sizeBarWidth(1, 0) = %d, want 0", got)
	}
}

func TestColumnTypeIcon(t *testing.T) {
	fn := getFunc("columnTypeIcon").(func(string) string)
	tests := map[string]string{
		"uuid":        "key",
		"VARCHAR":     "text_fields",
		"text":        "text_fields",
		"TIMESTAMP":   "schedule",
		"TIMESTAMPTZ": "schedule",
		"jsonb":       "data_object",
		"INET":        "language",
		"boolean":     "toggle_on",
		"INTEGER":     "tag",
		"BIGINT":      "tag",
		"SMALLINT":    "tag",
		"NUMERIC":     "calculate",
		"BYTEA":       "memory",
		"unknown_type": "help_outline",
	}
	for input, expected := range tests {
		if got := fn(input); got != expected {
			t.Errorf("columnTypeIcon(%q) = %q, want %q", input, got, expected)
		}
	}
}

func TestColumnTypeBadge(t *testing.T) {
	fn := getFunc("columnTypeBadge").(func(string) string)
	tests := map[string]string{
		"uuid":      "badge-uuid",
		"VARCHAR":   "badge-text",
		"TIMESTAMP": "badge-timestamp",
		"JSONB":     "badge-jsonb",
		"INET":      "badge-inet",
		"BOOLEAN":   "badge-boolean",
		"INTEGER":   "badge-integer",
		"xyz":       "badge-neutral",
	}
	for input, expected := range tests {
		if got := fn(input); got != expected {
			t.Errorf("columnTypeBadge(%q) = %q, want %q", input, got, expected)
		}
	}
}

func TestConstraintBadge(t *testing.T) {
	fn := getFunc("constraintBadge").(func(string) string)
	tests := map[string]string{
		"PK":          "badge-pk",
		"PRIMARY KEY": "badge-pk",
		"FK":          "badge-fk",
		"UNIQUE":      "badge-unique",
		"CHECK":       "badge-check",
		"NOT NULL":    "badge-notnull",
		"other":       "badge-neutral",
	}
	for input, expected := range tests {
		if got := fn(input); got != expected {
			t.Errorf("constraintBadge(%q) = %q, want %q", input, got, expected)
		}
	}
}

func TestIndexTypeBadge(t *testing.T) {
	fn := getFunc("indexTypeBadge").(func(string) string)
	tests := map[string]string{
		"BTREE": "badge-btree",
		"GIN":   "badge-gin",
		"GIST":  "badge-gist",
		"HASH":  "badge-hash",
		"BRIN":  "badge-brin",
		"other": "badge-neutral",
	}
	for input, expected := range tests {
		if got := fn(input); got != expected {
			t.Errorf("indexTypeBadge(%q) = %q, want %q", input, got, expected)
		}
	}
}

func TestQueryStateClass(t *testing.T) {
	fn := getFunc("queryStateClass").(func(string) string)
	tests := map[string]string{
		"active":                          "badge-success",
		"idle":                            "badge-neutral",
		"idle in transaction":             "badge-warning",
		"idle in transaction (aborted)":   "badge-danger",
		"fastpath function call":          "badge-info",
	}
	for input, expected := range tests {
		if got := fn(input); got != expected {
			t.Errorf("queryStateClass(%q) = %q, want %q", input, got, expected)
		}
	}
}

func TestCacheHitClass(t *testing.T) {
	fn := getFunc("cacheHitClass").(func(any) string)
	tests := []struct {
		input    float64
		expected string
	}{
		{99.5, "cache-excellent"},
		{99.0, "cache-excellent"},
		{97.0, "cache-good"},
		{95.0, "cache-good"},
		{92.0, "cache-warning"},
		{80.0, "cache-critical"},
	}
	for _, tt := range tests {
		if got := fn(tt.input); got != tt.expected {
			t.Errorf("cacheHitClass(%f) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestCacheHitColor(t *testing.T) {
	fn := getFunc("cacheHitColor").(func(any) string)
	if got := fn(99.5); got != "var(--emerald)" {
		t.Errorf("cacheHitColor(99.5) = %q, want var(--emerald)", got)
	}
	if got := fn(96.0); got != "var(--amber)" {
		t.Errorf("cacheHitColor(96.0) = %q, want var(--amber)", got)
	}
	if got := fn(80.0); got != "var(--rose)" {
		t.Errorf("cacheHitColor(80.0) = %q, want var(--rose)", got)
	}
}

func TestDeadTupleWarning(t *testing.T) {
	fn := getFunc("deadTupleWarning").(func(any, any) bool)
	if got := fn(25, 100); got != true {
		t.Error("deadTupleWarning(25, 100) should be true (25%)")
	}
	if got := fn(10, 100); got != false {
		t.Error("deadTupleWarning(10, 100) should be false (10%)")
	}
	if got := fn(5, 0); got != true {
		t.Error("deadTupleWarning(5, 0) should be true (dead > 0 with 0 live)")
	}
	if got := fn(0, 0); got != false {
		t.Error("deadTupleWarning(0, 0) should be false")
	}
}

func TestConnectionUtilClass(t *testing.T) {
	fn := getFunc("connectionUtilClass").(func(any, any) string)
	if got := fn(95, 100); got != "util-critical" {
		t.Errorf("connectionUtilClass(95, 100) = %q, want util-critical", got)
	}
	if got := fn(85, 100); got != "util-warning" {
		t.Errorf("connectionUtilClass(85, 100) = %q, want util-warning", got)
	}
	if got := fn(65, 100); got != "util-elevated" {
		t.Errorf("connectionUtilClass(65, 100) = %q, want util-elevated", got)
	}
	if got := fn(30, 100); got != "util-normal" {
		t.Errorf("connectionUtilClass(30, 100) = %q, want util-normal", got)
	}
}

func TestTruncateSQL(t *testing.T) {
	fn := getFunc("truncateSQL").(func(string, int) string)
	if got := fn("SELECT * FROM users", 50); got != "SELECT * FROM users" {
		t.Errorf("truncateSQL short = %q", got)
	}
	long := "SELECT id, username, email, phone, role_id, active, created_at, updated_at FROM users WHERE active = true"
	if got := fn(long, 30); len(got) != 33 { // 30 + "..."
		t.Errorf("truncateSQL long len = %d, want 33, got %q", len(got), got)
	}
	// Test newline collapsing
	multiline := "SELECT *\nFROM users\nWHERE id = 1"
	if got := fn(multiline, 100); got != "SELECT * FROM users WHERE id = 1" {
		t.Errorf("truncateSQL multiline = %q", got)
	}
}

func TestFkActionBadge(t *testing.T) {
	fn := getFunc("fkActionBadge").(func(string) string)
	tests := map[string]string{
		"CASCADE":    "badge-danger",
		"SET NULL":   "badge-warning",
		"RESTRICT":   "badge-info",
		"NO ACTION":  "badge-neutral",
	}
	for input, expected := range tests {
		if got := fn(input); got != expected {
			t.Errorf("fkActionBadge(%q) = %q, want %q", input, got, expected)
		}
	}
}

func TestSeverityBadge(t *testing.T) {
	fn := getFunc("severityBadge").(func(string) string)
	tests := map[string]string{
		"CRITICAL": "badge-danger",
		"HIGH":     "badge-warning",
		"MEDIUM":   "badge-info",
		"LOW":      "badge-success",
	}
	for input, expected := range tests {
		if got := fn(input); got != expected {
			t.Errorf("severityBadge(%q) = %q, want %q", input, got, expected)
		}
	}
}

func TestBoolIcon(t *testing.T) {
	fn := getFunc("boolIcon").(func(any) template.HTML)
	trueResult := fn(true)
	if trueResult == "" {
		t.Error("boolIcon(true) returned empty")
	}
	falseResult := fn(false)
	if falseResult == "" {
		t.Error("boolIcon(false) returned empty")
	}
	if trueResult == falseResult {
		t.Error("boolIcon(true) and boolIcon(false) should differ")
	}
}

func TestTableTypeIcon(t *testing.T) {
	fn := getFunc("tableTypeIcon").(func(string) string)
	if got := fn("TABLE"); got != "table_chart" {
		t.Errorf("tableTypeIcon(TABLE) = %q", got)
	}
	if got := fn("VIEW"); got != "visibility" {
		t.Errorf("tableTypeIcon(VIEW) = %q", got)
	}
	if got := fn("MATERIALIZED VIEW"); got != "layers" {
		t.Errorf("tableTypeIcon(MATERIALIZED VIEW) = %q", got)
	}
}

func TestDict(t *testing.T) {
	fn := getFunc("dict").(func(...any) map[string]any)
	result := fn("key1", "value1", "key2", 42)
	if result == nil {
		t.Fatal("dict returned nil")
	}
	if result["key1"] != "value1" {
		t.Errorf("dict key1 = %v", result["key1"])
	}
	if result["key2"] != 42 {
		t.Errorf("dict key2 = %v", result["key2"])
	}
	// Odd number of args
	if fn("key1") != nil {
		t.Error("dict with odd args should return nil")
	}
}

func TestTernary(t *testing.T) {
	fn := getFunc("ternary").(func(bool, any, any) any)
	if got := fn(true, "yes", "no"); got != "yes" {
		t.Errorf("ternary(true) = %v", got)
	}
	if got := fn(false, "yes", "no"); got != "no" {
		t.Errorf("ternary(false) = %v", got)
	}
}

func TestMod(t *testing.T) {
	fn := getFunc("mod").(func(any, any) int64)
	if got := fn(10, 3); got != 1 {
		t.Errorf("mod(10, 3) = %d", got)
	}
	if got := fn(10, 0); got != 0 {
		t.Errorf("mod(10, 0) = %d, want 0", got)
	}
}

func TestSafeDiv(t *testing.T) {
	fn := getFunc("safeDiv").(func(any, any) float64)
	if got := fn(10, 2); got != 5.0 {
		t.Errorf("safeDiv(10, 2) = %f", got)
	}
	if got := fn(10, 0); got != 0.0 {
		t.Errorf("safeDiv(10, 0) = %f, want 0", got)
	}
}

func TestFloatFmt(t *testing.T) {
	fn := getFunc("floatFmt").(func(any, int) string)
	if got := fn(3.14159, 2); got != "3.14" {
		t.Errorf("floatFmt(3.14159, 2) = %q", got)
	}
	if got := fn(99.999, 1); got != "100.0" {
		t.Errorf("floatFmt(99.999, 1) = %q", got)
	}
}

func TestToFloat64(t *testing.T) {
	tests := []struct {
		input    any
		expected float64
	}{
		{float64(3.14), 3.14},
		{float32(2.5), 2.5},
		{int(42), 42.0},
		{int64(100), 100.0},
		{int32(50), 50.0},
		{uint(10), 10.0},
		{uint64(200), 200.0},
		{"string", 0.0},
		{nil, 0.0},
	}
	for _, tt := range tests {
		got := toFloat64(tt.input)
		if got != tt.expected {
			t.Errorf("toFloat64(%v) = %f, want %f", tt.input, got, tt.expected)
		}
	}
}

func TestToInt64Extended(t *testing.T) {
	tests := []struct {
		input    any
		expected int64
	}{
		{int(42), 42},
		{int64(100), 100},
		{float64(3.7), 3},
		{int32(50), 50},
		{int16(25), 25},
		{int8(10), 10},
		{uint(5), 5},
		{uint64(200), 200},
		{uint32(150), 150},
		{"string", 0},
	}
	for _, tt := range tests {
		got := toInt64(tt.input)
		if got != tt.expected {
			t.Errorf("toInt64(%v) = %d, want %d", tt.input, got, tt.expected)
		}
	}
}

func TestEqStrNeStr(t *testing.T) {
	eqFn := getFunc("eqStr").(func(string, string) bool)
	neFn := getFunc("neStr").(func(string, string) bool)

	if !eqFn("abc", "abc") {
		t.Error("eqStr(abc, abc) should be true")
	}
	if eqFn("abc", "xyz") {
		t.Error("eqStr(abc, xyz) should be false")
	}
	if !neFn("abc", "xyz") {
		t.Error("neStr(abc, xyz) should be true")
	}
	if neFn("abc", "abc") {
		t.Error("neStr(abc, abc) should be false")
	}
}

func TestQueryDurationClass(t *testing.T) {
	fn := getFunc("queryDurationClass").(func(any) string)
	if got := fn(120.0); got != "duration-critical" {
		t.Errorf("queryDurationClass(120) = %q, want duration-critical", got)
	}
	if got := fn(30.0); got != "duration-slow" {
		t.Errorf("queryDurationClass(30) = %q, want duration-slow", got)
	}
	if got := fn(5.0); got != "duration-warning" {
		t.Errorf("queryDurationClass(5) = %q, want duration-warning", got)
	}
	if got := fn(0.5); got != "duration-normal" {
		t.Errorf("queryDurationClass(0.5) = %q, want duration-normal", got)
	}
}

func TestLockModeBadge(t *testing.T) {
	fn := getFunc("lockModeBadge").(func(string) string)
	if got := fn("AccessExclusiveLock"); got != "badge-lock-accessexclusive" {
		t.Errorf("lockModeBadge(AccessExclusiveLock) = %q", got)
	}
	if got := fn("AccessShareLock"); got != "badge-lock-accessshare" {
		t.Errorf("lockModeBadge(AccessShareLock) = %q", got)
	}
}

func TestExplainNodeIcon(t *testing.T) {
	fn := getFunc("explainNodeIcon").(func(string) string)
	tests := map[string]string{
		"Seq Scan":        "view_list",
		"Index Scan":      "search",
		"Index Only Scan": "bolt",
		"Nested Loop":     "loop",
		"Hash Join":       "join",
		"Sort":            "sort",
		"Aggregate":       "functions",
		"Limit":           "filter_alt",
	}
	for input, expected := range tests {
		if got := fn(input); got != expected {
			t.Errorf("explainNodeIcon(%q) = %q, want %q", input, got, expected)
		}
	}
}

func TestWaitEventBadge(t *testing.T) {
	fn := getFunc("waitEventBadge").(func(string) string)
	if got := fn("Lock"); got != "badge-danger" {
		t.Errorf("waitEventBadge(Lock) = %q", got)
	}
	if got := fn("IO"); got != "badge-warning" {
		t.Errorf("waitEventBadge(IO) = %q", got)
	}
	if got := fn("Client"); got != "badge-info" {
		t.Errorf("waitEventBadge(Client) = %q", got)
	}
}

func TestConnectionUtilColor(t *testing.T) {
	fn := getFunc("connectionUtilColor").(func(any, any) string)
	if got := fn(85, 100); got != "var(--rose)" {
		t.Errorf("connectionUtilColor(85, 100) = %q", got)
	}
	if got := fn(65, 100); got != "var(--amber)" {
		t.Errorf("connectionUtilColor(65, 100) = %q", got)
	}
	if got := fn(30, 100); got != "var(--emerald)" {
		t.Errorf("connectionUtilColor(30, 100) = %q", got)
	}
	if got := fn(10, 0); got != "var(--text-muted)" {
		t.Errorf("connectionUtilColor(10, 0) = %q", got)
	}
}

func TestSeverityColor(t *testing.T) {
	fn := getFunc("severityColor").(func(string) string)
	if got := fn("CRITICAL"); got != "var(--rose)" {
		t.Errorf("severityColor(CRITICAL) = %q", got)
	}
	if got := fn("HIGH"); got != "var(--amber)" {
		t.Errorf("severityColor(HIGH) = %q", got)
	}
	if got := fn("LOW"); got != "var(--emerald)" {
		t.Errorf("severityColor(LOW) = %q", got)
	}
}

func TestTriggerTimingBadge(t *testing.T) {
	fn := getFunc("triggerTimingBadge").(func(string) string)
	if got := fn("BEFORE"); got != "badge-warning" {
		t.Errorf("triggerTimingBadge(BEFORE) = %q", got)
	}
	if got := fn("AFTER"); got != "badge-info" {
		t.Errorf("triggerTimingBadge(AFTER) = %q", got)
	}
}

func TestTriggerEventIcon(t *testing.T) {
	fn := getFunc("triggerEventIcon").(func(string) string)
	if got := fn("INSERT"); got != "add_circle" {
		t.Errorf("triggerEventIcon(INSERT) = %q", got)
	}
	if got := fn("DELETE"); got != "remove_circle" {
		t.Errorf("triggerEventIcon(DELETE) = %q", got)
	}
}

func TestConstraintIcon(t *testing.T) {
	fn := getFunc("constraintIcon").(func(string) string)
	if got := fn("PK"); got != "vpn_key" {
		t.Errorf("constraintIcon(PK) = %q", got)
	}
	if got := fn("FK"); got != "link" {
		t.Errorf("constraintIcon(FK) = %q", got)
	}
	if got := fn("UNIQUE"); got != "fingerprint" {
		t.Errorf("constraintIcon(UNIQUE) = %q", got)
	}
}

func TestList(t *testing.T) {
	fn := getFunc("list").(func(...any) []any)
	result := fn("a", "b", "c")
	if len(result) != 3 {
		t.Errorf("list(a,b,c) len = %d, want 3", len(result))
	}
}

func TestJoinStrings(t *testing.T) {
	fn := getFunc("joinStrings").(func([]string, string) string)
	if got := fn([]string{"a", "b", "c"}, ", "); got != "a, b, c" {
		t.Errorf("joinStrings = %q", got)
	}
}

func TestToUpperLower(t *testing.T) {
	upper := getFunc("toUpper").(func(string) string)
	lower := getFunc("toLower").(func(string) string)
	if got := upper("hello"); got != "HELLO" {
		t.Errorf("toUpper(hello) = %q", got)
	}
	if got := lower("HELLO"); got != "hello" {
		t.Errorf("toLower(HELLO) = %q", got)
	}
}
