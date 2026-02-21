package admin

import (
	"encoding/json"
	"fmt"
	"html/template"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"

	"gct/internal/domain"

	"github.com/google/uuid"
)

// Shared Template Functions
var templateFuncs = template.FuncMap{
	"currUserEmail": func(u *domain.User) string {
		if u != nil {
			if u.Username != nil && *u.Username != "" {
				return *u.Username
			}
			if u.Email != nil && *u.Email != "" {
				return *u.Email
			}
			if u.Phone != nil && *u.Phone != "" {
				return *u.Phone
			}
		}
		return "Admin"
	},
	"derefString": func(s *string) string {
		if s == nil {
			return ""
		}
		return *s
	},
	"derefBool": func(b *bool) bool {
		if b == nil {
			return false
		}
		return *b
	},
	"derefDeviceType": func(d *domain.SessionDeviceType) string {
		if d == nil {
			return ""
		}
		return string(*d)
	},
	"add": func(a, b any) int64 {
		return toInt64(a) + toInt64(b)
	},
	"sub": func(a, b any) int64 {
		return toInt64(a) - toInt64(b)
	},
	"formatUUID": func(id any) string {
		if id == nil {
			return ""
		}
		switch v := id.(type) {
		case uuid.UUID:
			return v.String()
		case string:
			return v
		case []byte:
			if len(v) == 16 {
				u, err := uuid.FromBytes(v)
				if err == nil {
					return u.String()
				}
			}
			return string(v)
		default:
			return fmt.Sprintf("%v", v)
		}
	},
	"formatTime": func(t any) string {
		if t == nil {
			return ""
		}
		switch v := t.(type) {
		case time.Time:
			return v.Format("02 Jan 2006, 15:04")
		case *time.Time:
			if v != nil {
				return v.Format("02 Jan 2006, 15:04")
			}
		}
		return ""
	},
	"seq": func(start, end int) []int {
		var res []int
		for i := start; i <= end; i++ {
			res = append(res, i)
		}
		return res
	},
	"totalPages": func(total, limit int64) int64 {
		if limit == 0 {
			return 1
		}
		return int64(math.Ceil(float64(total) / float64(limit)))
	},
	"currPage": func(offset, limit int64) int64 {
		if limit == 0 {
			return 1
		}
		return (offset / limit) + 1
	},
	"toJSON": func(v any) template.JS {
		b, err := json.Marshal(v)
		if err != nil {
			return template.JS("{}")
		}
		return template.JS(b)
	},
	"paginationLink": func(currURL url.Values, page int64, limit int64) string {
		currURL.Set("page", strconv.FormatInt(page, 10))
		currURL.Set("limit", strconv.FormatInt(limit, 10))
		return "?" + currURL.Encode()
	},
	"default": func(d string, v any) any {
		if v == nil {
			return d
		}
		if s, ok := v.(*string); ok {
			if s == nil || *s == "" {
				return d
			}
			return *s
		}
		if s, ok := v.(string); ok {
			if s == "" {
				return d
			}
			return s
		}
		return v
	},
	"gt": func(a, b any) bool {
		return toInt64(a) > toInt64(b)
	},
	"lt": func(a, b any) bool {
		return toInt64(a) < toInt64(b)
	},
	"ge": func(a, b any) bool {
		return toInt64(a) >= toInt64(b)
	},
	"le": func(a, b any) bool {
		return toInt64(a) <= toInt64(b)
	},
	"eqNum": func(a, b any) bool {
		return toInt64(a) == toInt64(b)
	},
	"neNum": func(a, b any) bool {
		return toInt64(a) != toInt64(b)
	},
	"contains": func(s any, substr string) bool {
		var str string
		switch v := s.(type) {
		case string:
			str = v
		case *string:
			if v != nil {
				str = *v
			} else {
				return false
			}
		default:
			return false
		}
		return strings.Contains(str, substr)
	},
	"timeAgo": func(t time.Time) string {
		d := time.Since(t)
		switch {
		case d < time.Minute:
			return "Just now"
		case d < time.Hour:
			m := int(d.Minutes())
			if m == 1 {
				return "1m ago"
			}
			return fmt.Sprintf("%dm ago", m)
		case d < 24*time.Hour:
			h := int(d.Hours())
			if h == 1 {
				return "1h ago"
			}
			return fmt.Sprintf("%dh ago", h)
		case d < 7*24*time.Hour:
			days := int(d.Hours() / 24)
			if days == 1 {
				return "1d ago"
			}
			return fmt.Sprintf("%dd ago", days)
		default:
			return t.Format("Jan 02, 2006")
		}
	},
	"auditActionIcon": func(action domain.AuditActionType) string {
		switch action {
		case domain.AuditActionLogin:
			return "login"
		case domain.AuditActionLogout:
			return "logout"
		case domain.AuditActionSessionRevoke:
			return "cancel"
		case domain.AuditActionPasswordChange:
			return "password"
		case domain.AuditActionAccessGranted:
			return "check_circle"
		case domain.AuditActionAccessDenied:
			return "block"
		case domain.AuditActionUserCreate:
			return "person_add"
		case domain.AuditActionUserUpdate:
			return "edit"
		case domain.AuditActionUserDelete:
			return "person_remove"
		case domain.AuditActionRoleAssign:
			return "badge"
		case domain.AuditActionRoleRemove:
			return "remove_circle"
		case domain.AuditActionPolicyMatched, domain.AuditActionPolicyEvaluated:
			return "policy"
		case domain.AuditActionPolicyDenied:
			return "gpp_bad"
		case domain.AuditActionAdminChange:
			return "admin_panel_settings"
		default:
			return "info"
		}
	},
	"auditActionColor": func(action domain.AuditActionType) string {
		switch action {
		case domain.AuditActionLogin:
			return "var(--emerald)"
		case domain.AuditActionLogout:
			return "var(--sky)"
		case domain.AuditActionAccessGranted:
			return "var(--emerald)"
		case domain.AuditActionAccessDenied, domain.AuditActionPolicyDenied:
			return "var(--rose)"
		case domain.AuditActionUserCreate:
			return "var(--accent-400)"
		case domain.AuditActionUserDelete:
			return "var(--rose)"
		case domain.AuditActionAdminChange:
			return "var(--amber)"
		default:
			return "var(--text-muted)"
		}
	},
	"statusCodeColor": func(code int) string {
		switch {
		case code >= 500:
			return "badge-danger"
		case code >= 400:
			return "badge-warning"
		case code >= 300:
			return "badge-info"
		case code >= 200:
			return "badge-success"
		default:
			return "badge-neutral"
		}
	},
	"durationColor": func(ms int) string {
		switch {
		case ms > 1000:
			return "var(--rose)"
		case ms > 500:
			return "var(--amber)"
		default:
			return "var(--emerald)"
		}
	},
	"methodColor": func(method string) string {
		switch strings.ToUpper(method) {
		case "GET":
			return "badge-success"
		case "POST":
			return "badge-info"
		case "PUT":
			return "badge-warning"
		case "PATCH":
			return "badge-neutral"
		case "DELETE":
			return "badge-danger"
		default:
			return "badge-neutral"
		}
	},
	"shortUUID": func(id any) string {
		s := ""
		switch v := id.(type) {
		case uuid.UUID:
			s = v.String()
		case string:
			s = v
		default:
			s = fmt.Sprintf("%v", v)
		}
		if len(s) >= 8 {
			return s[:8]
		}
		return s
	},
	"prettyJSON": func(v any) string {
		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return "{}"
		}
		return string(b)
	},
	"mul": func(a, b any) int64 {
		return toInt64(a) * toInt64(b)
	},
	"div": func(a, b any) int64 {
		bv := toInt64(b)
		if bv == 0 {
			return 0
		}
		return toInt64(a) / bv
	},
	"derefUUID": func(u *uuid.UUID) string {
		if u == nil {
			return ""
		}
		return u.String()
	},
	"derefInt": func(i *int) int {
		if i == nil {
			return 0
		}
		return *i
	},
	"derefTime": func(t *time.Time) string {
		if t == nil {
			return ""
		}
		return t.Format("02 Jan 2006, 15:04")
	},
	"hasPrefix": func(s, prefix string) bool {
		return strings.HasPrefix(s, prefix)
	},
	"mapGet": func(m map[string]string, key string) string {
		return m[key]
	},

	// ─── Database Module Template Functions ───

	// formatBytes converts bytes to human-readable (KB, MB, GB, TB)
	"formatBytes": func(bytes any) string {
		b := toFloat64(bytes)
		if b == 0 {
			return "0 B"
		}
		units := []string{"B", "KB", "MB", "GB", "TB"}
		i := 0
		for b >= 1024 && i < len(units)-1 {
			b /= 1024
			i++
		}
		if i == 0 {
			return fmt.Sprintf("%.0f B", b)
		}
		return fmt.Sprintf("%.1f %s", b, units[i])
	},

	// formatPercent formats a float as a percentage string
	"formatPercent": func(v any) string {
		f := toFloat64(v)
		return fmt.Sprintf("%.2f%%", f)
	},

	// formatNumber formats an integer with comma separators
	"formatNumber": func(n any) string {
		v := toInt64(n)
		s := strconv.FormatInt(v, 10)
		if len(s) <= 3 {
			return s
		}
		var result strings.Builder
		for i, c := range s {
			if i > 0 && (len(s)-i)%3 == 0 {
				result.WriteRune(',')
			}
			result.WriteRune(c)
		}
		return result.String()
	},

	// formatDurationMs formats milliseconds to human-readable duration
	"formatDurationMs": func(ms any) string {
		f := toFloat64(ms)
		switch {
		case f < 1:
			return fmt.Sprintf("%.2f µs", f*1000)
		case f < 1000:
			return fmt.Sprintf("%.1f ms", f)
		case f < 60000:
			return fmt.Sprintf("%.2f s", f/1000)
		case f < 3600000:
			return fmt.Sprintf("%.1f min", f/60000)
		default:
			return fmt.Sprintf("%.1f h", f/3600000)
		}
	},

	// percentage calculates (a / b) * 100 safely
	"percentage": func(a, b any) float64 {
		av := toFloat64(a)
		bv := toFloat64(b)
		if bv == 0 {
			return 0
		}
		return (av / bv) * 100
	},

	// sizeBarWidth calculates proportional width for size bars (0-100)
	"sizeBarWidth": func(value, max any) int64 {
		v := toFloat64(value)
		m := toFloat64(max)
		if m == 0 {
			return 0
		}
		w := (v / m) * 100
		if w < 1 && v > 0 {
			return 1
		}
		return int64(w)
	},

	// columnTypeIcon returns Material Symbols icon name per PostgreSQL column data type (#454)
	"columnTypeIcon": func(dataType string) string {
		dt := strings.ToUpper(dataType)
		switch {
		case strings.Contains(dt, "UUID"):
			return "key"
		case strings.Contains(dt, "VARCHAR") || strings.Contains(dt, "TEXT") || strings.Contains(dt, "CHAR"):
			return "text_fields"
		case strings.Contains(dt, "TIMESTAMPTZ") || strings.Contains(dt, "TIMESTAMP"):
			return "schedule"
		case strings.Contains(dt, "DATE"):
			return "calendar_today"
		case strings.Contains(dt, "JSONB") || strings.Contains(dt, "JSON"):
			return "data_object"
		case strings.Contains(dt, "INET") || strings.Contains(dt, "CIDR"):
			return "language"
		case strings.Contains(dt, "BOOLEAN") || strings.Contains(dt, "BOOL"):
			return "toggle_on"
		case strings.Contains(dt, "BIGINT") || strings.Contains(dt, "INTEGER") || strings.Contains(dt, "SMALLINT") || strings.Contains(dt, "INT"):
			return "tag"
		case strings.Contains(dt, "NUMERIC") || strings.Contains(dt, "DECIMAL") || strings.Contains(dt, "REAL") || strings.Contains(dt, "DOUBLE"):
			return "calculate"
		case strings.Contains(dt, "BYTEA"):
			return "memory"
		case strings.Contains(dt, "ARRAY"):
			return "list"
		default:
			return "help_outline"
		}
	},

	// columnTypeBadge returns CSS badge class per column type family (#708)
	"columnTypeBadge": func(dataType string) string {
		dt := strings.ToUpper(dataType)
		switch {
		case strings.Contains(dt, "UUID"):
			return "badge-uuid"
		case strings.Contains(dt, "VARCHAR") || strings.Contains(dt, "TEXT") || strings.Contains(dt, "CHAR"):
			return "badge-text"
		case strings.Contains(dt, "TIMESTAMP") || strings.Contains(dt, "DATE"):
			return "badge-timestamp"
		case strings.Contains(dt, "JSONB") || strings.Contains(dt, "JSON"):
			return "badge-jsonb"
		case strings.Contains(dt, "INET") || strings.Contains(dt, "CIDR"):
			return "badge-inet"
		case strings.Contains(dt, "BOOLEAN") || strings.Contains(dt, "BOOL"):
			return "badge-boolean"
		case strings.Contains(dt, "BIGINT") || strings.Contains(dt, "INTEGER") || strings.Contains(dt, "SMALLINT") || strings.Contains(dt, "INT"):
			return "badge-integer"
		default:
			return "badge-neutral"
		}
	},

	// constraintBadge returns CSS badge class per constraint type (#709)
	"constraintBadge": func(constraintType string) string {
		switch strings.ToUpper(constraintType) {
		case "PK", "PRIMARY KEY", "P":
			return "badge-pk"
		case "FK", "FOREIGN KEY", "F":
			return "badge-fk"
		case "UNIQUE", "U":
			return "badge-unique"
		case "CHECK", "C":
			return "badge-check"
		case "NOT NULL":
			return "badge-notnull"
		case "EXCLUSION":
			return "badge-exclusion"
		default:
			return "badge-neutral"
		}
	},

	// constraintIcon returns Material icon per constraint type
	"constraintIcon": func(constraintType string) string {
		switch strings.ToUpper(constraintType) {
		case "PK", "PRIMARY KEY", "P":
			return "vpn_key"
		case "FK", "FOREIGN KEY", "F":
			return "link"
		case "UNIQUE", "U":
			return "fingerprint"
		case "CHECK", "C":
			return "verified"
		case "NOT NULL":
			return "block"
		default:
			return "rule"
		}
	},

	// indexTypeBadge returns CSS badge class per index type (#710)
	"indexTypeBadge": func(indexType string) string {
		switch strings.ToUpper(indexType) {
		case "BTREE":
			return "badge-btree"
		case "GIN":
			return "badge-gin"
		case "GIST":
			return "badge-gist"
		case "HASH":
			return "badge-hash"
		case "BRIN":
			return "badge-brin"
		default:
			return "badge-neutral"
		}
	},

	// lockModeBadge returns CSS badge class per lock mode (#711)
	"lockModeBadge": func(lockMode string) string {
		mode := strings.ToLower(lockMode)
		switch {
		case strings.Contains(mode, "accessexclusive"):
			return "badge-lock-accessexclusive"
		case strings.Contains(mode, "accessshare"):
			return "badge-lock-accessshare"
		case strings.Contains(mode, "rowexclusive"):
			return "badge-lock-rowexclusive"
		case strings.Contains(mode, "rowshare"):
			return "badge-lock-rowshare"
		case strings.Contains(mode, "exclusive"):
			return "badge-lock-exclusive"
		case strings.Contains(mode, "share"):
			return "badge-lock-share"
		default:
			return "badge-neutral"
		}
	},

	// queryStateClass returns CSS badge class per query state (#575)
	"queryStateClass": func(state string) string {
		switch strings.ToLower(state) {
		case "active":
			return "badge-success"
		case "idle":
			return "badge-neutral"
		case "idle in transaction":
			return "badge-warning"
		case "idle in transaction (aborted)":
			return "badge-danger"
		case "fastpath function call":
			return "badge-info"
		case "disabled":
			return "badge-neutral"
		default:
			return "badge-neutral"
		}
	},

	// queryDurationClass returns CSS class per duration threshold (#574)
	"queryDurationClass": func(seconds any) string {
		s := toFloat64(seconds)
		switch {
		case s > 60:
			return "duration-critical"
		case s > 10:
			return "duration-slow"
		case s > 1:
			return "duration-warning"
		default:
			return "duration-normal"
		}
	},

	// cacheHitClass returns CSS class per cache hit ratio (#522, #558)
	"cacheHitClass": func(ratio any) string {
		r := toFloat64(ratio)
		switch {
		case r >= 99:
			return "cache-excellent"
		case r >= 95:
			return "cache-good"
		case r >= 90:
			return "cache-warning"
		default:
			return "cache-critical"
		}
	},

	// cacheHitColor returns inline color per cache hit ratio
	"cacheHitColor": func(ratio any) string {
		r := toFloat64(ratio)
		switch {
		case r >= 99:
			return "var(--emerald)"
		case r >= 95:
			return "var(--amber)"
		default:
			return "var(--rose)"
		}
	},

	// deadTupleWarning returns true if dead tuple ratio is concerning (#514, #712)
	"deadTupleWarning": func(dead, live any) bool {
		d := toFloat64(dead)
		l := toFloat64(live)
		if l == 0 {
			return d > 0
		}
		return (d / l) > 0.2
	},

	// connectionUtilClass returns CSS class per connection utilization (#542, #718)
	"connectionUtilClass": func(active, max any) string {
		a := toFloat64(active)
		m := toFloat64(max)
		if m == 0 {
			return "util-normal"
		}
		pct := (a / m) * 100
		switch {
		case pct >= 90:
			return "util-critical"
		case pct >= 80:
			return "util-warning"
		case pct >= 60:
			return "util-elevated"
		default:
			return "util-normal"
		}
	},

	// connectionUtilColor returns inline color per connection utilization
	"connectionUtilColor": func(active, max any) string {
		a := toFloat64(active)
		m := toFloat64(max)
		if m == 0 {
			return "var(--text-muted)"
		}
		pct := (a / m) * 100
		switch {
		case pct >= 80:
			return "var(--rose)"
		case pct >= 60:
			return "var(--amber)"
		default:
			return "var(--emerald)"
		}
	},

	// truncateSQL truncates SQL for display with ellipsis (#577, #591)
	"truncateSQL": func(sql string, maxLen int) string {
		s := strings.TrimSpace(sql)
		s = strings.ReplaceAll(s, "\n", " ")
		s = strings.Join(strings.Fields(s), " ")
		if len(s) <= maxLen {
			return s
		}
		return s[:maxLen] + "..."
	},

	// fkActionBadge returns CSS class per FK action (#457)
	"fkActionBadge": func(action string) string {
		switch strings.ToUpper(action) {
		case "CASCADE":
			return "badge-danger"
		case "SET NULL":
			return "badge-warning"
		case "RESTRICT":
			return "badge-info"
		case "NO ACTION":
			return "badge-neutral"
		case "SET DEFAULT":
			return "badge-neutral"
		default:
			return "badge-neutral"
		}
	},

	// severityBadge returns CSS badge class per error severity
	"severityBadge": func(severity string) string {
		switch strings.ToUpper(severity) {
		case "CRITICAL":
			return "badge-danger"
		case "HIGH":
			return "badge-warning"
		case "MEDIUM":
			return "badge-info"
		case "LOW":
			return "badge-success"
		default:
			return "badge-neutral"
		}
	},

	// severityColor returns inline color per severity
	"severityColor": func(severity string) string {
		switch strings.ToUpper(severity) {
		case "CRITICAL":
			return "var(--rose)"
		case "HIGH":
			return "var(--amber)"
		case "MEDIUM":
			return "var(--sky)"
		case "LOW":
			return "var(--emerald)"
		default:
			return "var(--text-muted)"
		}
	},

	// boolIcon returns check/cross icon for boolean display (#501)
	"boolIcon": func(v any) template.HTML {
		var b bool
		switch val := v.(type) {
		case bool:
			b = val
		case *bool:
			if val != nil {
				b = *val
			}
		}
		if b {
			return template.HTML(`<i class="material-symbols-outlined" style="color: var(--emerald); font-size: 18px;">check_circle</i>`)
		}
		return template.HTML(`<i class="material-symbols-outlined" style="color: var(--rose); font-size: 18px;">cancel</i>`)
	},

	// nullDisplay returns styled NULL text (#506)
	"nullDisplay": func() template.HTML {
		return template.HTML(`<span style="color: var(--text-muted); font-style: italic; font-size: 12px;">NULL</span>`)
	},

	// triggerTimingBadge returns CSS class per trigger timing (#487)
	"triggerTimingBadge": func(timing string) string {
		switch strings.ToUpper(timing) {
		case "BEFORE":
			return "badge-warning"
		case "AFTER":
			return "badge-info"
		case "INSTEAD OF":
			return "badge-neutral"
		default:
			return "badge-neutral"
		}
	},

	// triggerEventIcon returns icon per trigger event (#491)
	"triggerEventIcon": func(event string) string {
		switch strings.ToUpper(event) {
		case "INSERT":
			return "add_circle"
		case "UPDATE":
			return "edit"
		case "DELETE":
			return "remove_circle"
		case "TRUNCATE":
			return "delete_sweep"
		default:
			return "bolt"
		}
	},

	// tableTypeIcon returns icon per table type (#451)
	"tableTypeIcon": func(tableType string) string {
		switch strings.ToUpper(tableType) {
		case "TABLE", "BASE TABLE":
			return "table_chart"
		case "VIEW":
			return "visibility"
		case "MATERIALIZED VIEW":
			return "layers"
		case "PARTITIONED TABLE":
			return "grid_view"
		case "FOREIGN TABLE":
			return "cloud"
		default:
			return "table_chart"
		}
	},

	// waitEventBadge returns CSS class per wait event type (#576)
	"waitEventBadge": func(eventType string) string {
		switch strings.ToUpper(eventType) {
		case "LOCK":
			return "badge-danger"
		case "IO":
			return "badge-warning"
		case "CLIENT":
			return "badge-info"
		case "IPC":
			return "badge-info"
		case "LWLOCK":
			return "badge-warning"
		case "BUFFERPIN":
			return "badge-neutral"
		default:
			return "badge-neutral"
		}
	},

	// explainNodeIcon returns icon per EXPLAIN node type (#672)
	"explainNodeIcon": func(nodeType string) string {
		nt := strings.ToLower(nodeType)
		switch {
		case strings.Contains(nt, "seq scan"):
			return "view_list"
		case strings.Contains(nt, "index only scan"):
			return "bolt"
		case strings.Contains(nt, "index scan"):
			return "search"
		case strings.Contains(nt, "bitmap"):
			return "grid_on"
		case strings.Contains(nt, "nested loop"):
			return "loop"
		case strings.Contains(nt, "hash join"):
			return "join"
		case strings.Contains(nt, "merge join"):
			return "merge_type"
		case strings.Contains(nt, "sort"):
			return "sort"
		case strings.Contains(nt, "aggregate"):
			return "functions"
		case strings.Contains(nt, "limit"):
			return "filter_alt"
		case strings.Contains(nt, "hash"):
			return "fingerprint"
		default:
			return "play_arrow"
		}
	},

	// toUpper converts string to uppercase
	"toUpper": func(s string) string {
		return strings.ToUpper(s)
	},

	// toLower converts string to lowercase
	"toLower": func(s string) string {
		return strings.ToLower(s)
	},

	// joinStrings joins a string slice with separator
	"joinStrings": func(items []string, sep string) string {
		return strings.Join(items, sep)
	},

	// mod returns a % b
	"mod": func(a, b any) int64 {
		bv := toInt64(b)
		if bv == 0 {
			return 0
		}
		return toInt64(a) % bv
	},

	// safeDiv returns a / b as float, safe for division by zero
	"safeDiv": func(a, b any) float64 {
		bv := toFloat64(b)
		if bv == 0 {
			return 0
		}
		return toFloat64(a) / bv
	},

	// floatFmt formats a float with given precision
	"floatFmt": func(f any, prec int) string {
		return fmt.Sprintf("%.*f", prec, toFloat64(f))
	},

	// dict creates a map from key-value pairs for passing to templates
	"dict": func(values ...any) map[string]any {
		if len(values)%2 != 0 {
			return nil
		}
		m := make(map[string]any, len(values)/2)
		for i := 0; i < len(values); i += 2 {
			key, ok := values[i].(string)
			if !ok {
				continue
			}
			m[key] = values[i+1]
		}
		return m
	},

	// list creates a slice from arguments
	"list": func(values ...any) []any {
		return values
	},

	// eq compares two strings
	"eqStr": func(a, b string) bool {
		return a == b
	},

	// neStr compares two strings for inequality
	"neStr": func(a, b string) bool {
		return a != b
	},

	// ternary returns a if condition is true, b otherwise
	"ternary": func(cond bool, a, b any) any {
		if cond {
			return a
		}
		return b
	},

	// safeHTML marks a string as safe HTML
	"safeHTML": func(s string) template.HTML {
		return template.HTML(s)
	},
}

func toInt64(v any) int64 {
	switch val := v.(type) {
	case int:
		return int64(val)
	case int64:
		return val
	case float64:
		return int64(val)
	case int32:
		return int64(val)
	case int16:
		return int64(val)
	case int8:
		return int64(val)
	case uint:
		return int64(val)
	case uint64:
		return int64(val)
	case uint32:
		return int64(val)
	default:
		return 0
	}
}

func toFloat64(v any) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case int32:
		return float64(val)
	case int16:
		return float64(val)
	case int8:
		return float64(val)
	case uint:
		return float64(val)
	case uint64:
		return float64(val)
	case uint32:
		return float64(val)
	default:
		return 0
	}
}
