package events

// FieldChange represents a single field-level mutation in a domain entity.
// Each changed field in an update produces one FieldChange entry.
// For create operations, OldValue is empty. For delete, NewValue is empty.
type FieldChange struct {
	FieldName string `json:"field_name"`
	OldValue  string `json:"old_value"`
	NewValue  string `json:"new_value"`
}

// RedactedValue is the sentinel used for sensitive fields (password, tokens)
// so the fact of the change is recorded without exposing actual values.
const RedactedValue = "[REDACTED]"
