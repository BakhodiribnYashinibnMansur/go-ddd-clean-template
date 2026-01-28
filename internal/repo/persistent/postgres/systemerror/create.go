package systemerror

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"

	"gct/consts"
	"gct/internal/repo/schema"
	"gct/pkg/errorx"
)

// Create satisfies errorx.Repository interface
func (r *Repo) Create(ctx context.Context, input errorx.LogErrorInput) error {
	stackTrace := string(debug.Stack())

	message := input.Message
	if input.Err != nil {
		message = fmt.Sprintf("%s: %v", input.Message, input.Err)
	}

	_, err := r.CreateSystemError(ctx, CreateSystemErrorInput{
		Code:        input.Code,
		Message:     message,
		StackTrace:  &stackTrace,
		Metadata:    input.Metadata,
		Severity:    input.Severity,
		ServiceName: input.ServiceName,
		RequestID:   input.RequestID,
		UserID:      input.UserID,
		IPAddress:   input.IPAddress,
		Path:        input.Path,
		Method:      input.Method,
	})
	return err
}

// CreateSystemError inserts a new system error record into the database
func (r *Repo) CreateSystemError(ctx context.Context, input CreateSystemErrorInput) (*SystemError, error) {
	// Convert metadata to JSONB
	var metadataJSON []byte
	var err error
	if input.Metadata != nil {
		metadataJSON, err = json.Marshal(input.Metadata)
		if err != nil {
			r.logger.Error("failed to marshal metadata", "error", err)
			return nil, err
		}
	}

	// Set default severity if not provided
	if input.Severity == "" {
		input.Severity = consts.SeverityError
	}

	// Set default service name if not provided
	if input.ServiceName == "" {
		input.ServiceName = consts.ServiceNameAPI
	}

	query, args, err := r.db.Builder.
		Insert(schema.TableSystemError).
		Columns(
			schema.SystemErrorCode,
			schema.SystemErrorMessage,
			schema.SystemErrorStackTrace,
			schema.SystemErrorMetadata,
			schema.SystemErrorSeverity,
			schema.SystemErrorServiceName,
			schema.SystemErrorRequestID,
			schema.SystemErrorUserID,
			schema.SystemErrorIPAddress,
			schema.SystemErrorPath,
			schema.SystemErrorMethod,
		).
		Values(
			input.Code,
			input.Message,
			input.StackTrace,
			metadataJSON,
			input.Severity,
			input.ServiceName,
			input.RequestID,
			input.UserID,
			input.IPAddress,
			input.Path,
			input.Method,
		).
		Suffix("RETURNING " +
			schema.SystemErrorID + ", " +
			schema.SystemErrorCode + ", " +
			schema.SystemErrorMessage + ", " +
			schema.SystemErrorStackTrace + ", " +
			schema.SystemErrorMetadata + ", " +
			schema.SystemErrorSeverity + ", " +
			schema.SystemErrorServiceName + ", " +
			schema.SystemErrorRequestID + ", " +
			schema.SystemErrorUserID + ", " +
			schema.SystemErrorIPAddress + ", " +
			schema.SystemErrorPath + ", " +
			schema.SystemErrorMethod + ", " +
			schema.SystemErrorIsResolved + ", " +
			schema.SystemErrorResolvedAt + ", " +
			schema.SystemErrorResolvedBy + ", " +
			schema.SystemErrorCreatedAt).
		ToSql()

	if err != nil {
		r.logger.Error("failed to build create query", "error", err)
		return nil, err
	}

	var se SystemError
	err = r.db.Pool.QueryRow(ctx, query, args...).Scan(
		&se.ID,
		&se.Code,
		&se.Message,
		&se.StackTrace,
		&se.Metadata,
		&se.Severity,
		&se.ServiceName,
		&se.RequestID,
		&se.UserID,
		&se.IPAddress,
		&se.Path,
		&se.Method,
		&se.IsResolved,
		&se.ResolvedAt,
		&se.ResolvedBy,
		&se.CreatedAt,
	)

	if err != nil {
		r.logger.Error("failed to create system error record", "error", err, "code", input.Code)
		return nil, err
	}

	r.logger.Info("system error logged to database",
		"error_id", se.ID,
		"code", se.Code,
		"severity", se.Severity,
	)

	return &se, nil
}
