package logger

// F provides type-safe structured log fields for operation logging.
type F struct {
	Op       string // operation name (e.g., "CreateAnnouncement")
	Entity   string // domain entity type (e.g., "announcement")
	EntityID any    // entity identifier, when available
	Err      error  // the original error
}

// KV converts F into a flat key-value slice suitable for structured logging methods.
func (f F) KV() []any {
	fields := make([]any, 0, 8)
	if f.Op != "" {
		fields = append(fields, "operation", f.Op)
	}
	if f.Entity != "" {
		fields = append(fields, "entity", f.Entity)
	}
	if f.EntityID != nil {
		fields = append(fields, "entity_id", f.EntityID)
	}
	if f.Err != nil {
		fields = append(fields, "error", f.Err)
	}
	return fields
}
