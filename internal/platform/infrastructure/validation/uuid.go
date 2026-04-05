package validation

import "github.com/google/uuid"

// IsValidUUID checks if the string is a valid UUID.
func IsValidUUID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}
