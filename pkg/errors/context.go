package errors

import (
	"path/filepath"
	"runtime"
	"strings"
)

const (
	unknownValue = "unknown"
)

// WithSource adds source file and function information to error
// This helps track where the error originated in the codebase
func WithSource(err *AppError, file, function string) *AppError {
	if err == nil {
		return nil
	}
	return err.
		WithField("file", file).
		WithField("function", function)
}

// GetCaller returns the file and function name of the caller
// Skip levels: 0 = GetCaller itself, 1 = caller of GetCaller, etc.
// Returns relative path from project root (internal/repo/...) and short function name
func GetCaller(skip int) (file, function string) {
	pc, fullPath, _, ok := runtime.Caller(skip + 1)
	if !ok {
		return unknownValue, unknownValue
	}

	// Extract relative path from project root
	// Look for common project markers: internal/, pkg/, cmd/
	file = fullPath
	if idx := strings.Index(fullPath, "/internal/"); idx != -1 {
		file = fullPath[idx+1:] // Remove leading slash, keep "internal/..."
	} else if idx := strings.Index(fullPath, "/pkg/"); idx != -1 {
		file = fullPath[idx+1:]
	} else if idx := strings.Index(fullPath, "/cmd/"); idx != -1 {
		file = fullPath[idx+1:]
	} else {
		// If no marker found, just use the filename
		file = filepath.Base(fullPath)
	}

	// Get function name
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return file, unknownValue
	}

	// Extract short function name (e.g., "Repo.Get" instead of full path)
	funcName := fn.Name()
	// Split by "/" and take last part, then split by "." and take last 2 parts
	parts := strings.Split(funcName, "/")
	if len(parts) > 0 {
		lastPart := parts[len(parts)-1]
		// Now get package.Function or Struct.Method
		dotParts := strings.Split(lastPart, ".")
		if len(dotParts) >= 2 {
			// Return last 2 parts: "Repo.Get" or "package.Function"
			funcName = strings.Join(dotParts[len(dotParts)-2:], ".")
		}
	}

	return file, funcName
}

// WithCaller automatically adds caller information to error
// Usage: errors.WithCaller(err, 0) - will get the caller of WithCaller
func WithCaller(err *AppError, skip int) *AppError {
	if err == nil {
		return nil
	}
	file, function := GetCaller(skip + 1)
	return err.
		WithField("file", file).
		WithField("function", function)
}

// AutoSource automatically adds caller info with proper skip level
// This should be called directly from repository/service functions
func AutoSource(err *AppError) *AppError {
	if err == nil {
		return nil
	}
	// skip = 1 means we skip AutoSource itself and get the actual caller
	file, function := GetCaller(1)
	return err.
		WithField("file", file).
		WithField("function", function)
}
