package errorx

import (
	"errors"
	"testing"
)

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name     string
		appError *AppError
		want     string
	}{
		{
			name: "error without wrapped error",
			appError: &AppError{
				Code:    "1001",
				Message: "Bad request",
			},
			want: "[1001] Bad request",
		},
		{
			name: "error with wrapped error",
			appError: &AppError{
				Code:    "1002",
				Message: "Invalid input",
				Err:     errors.New("validation failed"),
			},
			want: "[1002] Invalid input: validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.appError.Error(); got != tt.want {
				t.Errorf("AppError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAppError_Unwrap(t *testing.T) {
	tests := []struct {
		name     string
		appError *AppError
		want     error
	}{
		{
			name: "error without wrapped error",
			appError: &AppError{
				Code:    "1001",
				Message: "Bad request",
			},
			want: nil,
		},
		{
			name: "error with wrapped error",
			appError: &AppError{
				Code:    "1002",
				Message: "Invalid input",
				Err:     errors.New("validation failed"),
			},
			want: errors.New("validation failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.appError.Unwrap(); !errors.Is(got, tt.want) && (got == nil || tt.want == nil || got.Error() != tt.want.Error()) {
				t.Errorf("AppError.Unwrap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAppError_WithField(t *testing.T) {
	tests := []struct {
		name     string
		appError *AppError
		key      string
		value    any
		want     map[string]any
	}{
		{
			name: "add field to nil fields",
			appError: &AppError{
				Code:    "1001",
				Message: "Bad request",
			},
			key:   "field1",
			value: "value1",
			want:  map[string]any{"field1": "value1"},
		},
		{
			name: "add field to existing fields",
			appError: &AppError{
				Code:    "1001",
				Message: "Bad request",
				Fields: map[string]any{
					"existing": "field",
				},
			},
			key:   "field1",
			value: "value1",
			want: map[string]any{
				"existing": "field",
				"field1":   "value1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.appError.WithField(tt.key, tt.value)
			if len(result.Fields) != len(tt.want) {
				t.Errorf("AppError.WithField() fields length = %v, want %v", len(result.Fields), len(tt.want))
			}
			for k, v := range tt.want {
				if result.Fields[k] != v {
					t.Errorf("AppError.WithField() field[%s] = %v, want %v", k, result.Fields[k], v)
				}
			}
		})
	}
}

func TestAppError_WithInput(t *testing.T) {
	tests := []struct {
		name     string
		appError *AppError
		input    any
		want     any
	}{
		{
			name: "add input data",
			appError: &AppError{
				Code:    "1001",
				Message: "Bad request",
			},
			input: "test input",
			want:  "test input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.appError.WithInput(tt.input)
			if result.Input != tt.want {
				t.Errorf("AppError.WithInput() = %v, want %v", result.Input, tt.want)
			}
		})
	}
}

func TestAppError_WithOutput(t *testing.T) {
	tests := []struct {
		name     string
		appError *AppError
		output   any
		want     any
	}{
		{
			name: "add output data",
			appError: &AppError{
				Code:    "1001",
				Message: "Bad request",
			},
			output: "test output",
			want:   "test output",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.appError.WithOutput(tt.output)
			if result.Output != tt.want {
				t.Errorf("AppError.WithOutput() = %v, want %v", result.Output, tt.want)
			}
		})
	}
}

func TestAppError_WithDetails(t *testing.T) {
	tests := []struct {
		name     string
		appError *AppError
		details  string
		want     string
	}{
		{
			name: "add details",
			appError: &AppError{
				Code:    "1001",
				Message: "Bad request",
			},
			details: "detailed error information",
			want:    "detailed error information",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.appError.WithDetails(tt.details)
			if result.Details != tt.want {
				t.Errorf("AppError.WithDetails() = %v, want %v", result.Details, tt.want)
			}
		})
	}
}

func TestNew(t *testing.T) {

	tests := []struct {
		name    string
		code    string
		message string
		want    *AppError
	}{
		{
			name:    "create new error with valid code",
			code:    ErrBadRequest,
			message: "Bad request",
			want: &AppError{
				Type:       ErrBadRequest,
				Code:       CodeBadRequest,
				Message:    "Bad request",
				HTTPStatus: 400,
				UserMsg:    "Bad request",
			},
		},
		{
			name:    "create new error with unknown code",
			code:    "UNKNOWN_ERROR",
			message: "Unknown error",
			want: &AppError{
				Type:       "UNKNOWN_ERROR",
				Code:       "1019",
				Message:    "Unknown error",
				HTTPStatus: 500,
				UserMsg:    "Unknown error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.code, tt.message)

			if got.Type != tt.want.Type {
				t.Errorf("New().Type = %v, want %v", got.Type, tt.want.Type)
			}
			if got.Code != tt.want.Code {
				t.Errorf("New().Code = %v, want %v", got.Code, tt.want.Code)
			}
			if got.Message != tt.want.Message {
				t.Errorf("New().Message = %v, want %v", got.Message, tt.want.Message)
			}
			if got.HTTPStatus != tt.want.HTTPStatus {
				t.Errorf("New().HTTPStatus = %v, want %v", got.HTTPStatus, tt.want.HTTPStatus)
			}
			if got.UserMsg != tt.want.UserMsg {
				t.Errorf("New().UserMsg = %v, want %v", got.UserMsg, tt.want.UserMsg)
			}
			if len(got.Stack) == 0 {
				t.Errorf("New().Stack should not be empty")
			}
		})
	}
}

func TestWrap(t *testing.T) {
	baseErr := errors.New("base error")

	tests := []struct {
		name    string
		err     error
		code    string
		message string
		want    *AppError
		wantNil bool
	}{
		{
			name:    "wrap non-nil error",
			err:     baseErr,
			code:    ErrBadRequest,
			message: "Bad request",
			want: &AppError{
				Type:       ErrBadRequest,
				Code:       CodeBadRequest,
				Message:    "Bad request",
				HTTPStatus: 400,
				UserMsg:    "Bad request",
				Err:        baseErr,
			},
			wantNil: false,
		},
		{
			name:    "wrap nil error",
			err:     nil,
			code:    ErrBadRequest,
			message: "Bad request",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Wrap(tt.err, tt.code, tt.message)

			if tt.wantNil {
				if got != nil {
					t.Errorf("Wrap() = %v, want nil", got)
				}
				return
			}

			if got.Type != tt.want.Type {
				t.Errorf("Wrap().Type = %v, want %v", got.Type, tt.want.Type)
			}
			if got.Code != tt.want.Code {
				t.Errorf("Wrap().Code = %v, want %v", got.Code, tt.want.Code)
			}
			if got.Message != tt.want.Message {
				t.Errorf("Wrap().Message = %v, want %v", got.Message, tt.want.Message)
			}
			if got.HTTPStatus != tt.want.HTTPStatus {
				t.Errorf("Wrap().HTTPStatus = %v, want %v", got.HTTPStatus, tt.want.HTTPStatus)
			}
			if got.UserMsg != tt.want.UserMsg {
				t.Errorf("Wrap().UserMsg = %v, want %v", got.UserMsg, tt.want.UserMsg)
			}
			if !errors.Is(got.Err, tt.want.Err) {
				t.Errorf("Wrap().Err = %v, want %v", got.Err, tt.want.Err)
			}
			if len(got.Stack) == 0 {
				t.Errorf("Wrap().Stack should not be empty")
			}
		})
	}
}

func TestIs(t *testing.T) {

	tests := []struct {
		name string
		err  error
		code string
		want bool
	}{
		{
			name: "matching AppError code",
			err:  New(ErrBadRequest, "Bad request"),
			code: ErrBadRequest,
			want: true,
		},
		{
			name: "non-matching AppError code",
			err:  New(ErrBadRequest, "Bad request"),
			code: ErrNotFound,
			want: false,
		},
		{
			name: "non-AppError",
			err:  errors.New("regular error"),
			code: ErrBadRequest,
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			code: ErrBadRequest,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Is(tt.err, tt.code); got != tt.want {
				t.Errorf("Is() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetCode(t *testing.T) {

	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "AppError with code",
			err:  New(ErrBadRequest, "Bad request"),
			want: ErrBadRequest,
		},
		{
			name: "non-AppError",
			err:  errors.New("regular error"),
			want: "",
		},
		{
			name: "nil error",
			err:  nil,
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetCode(tt.err); got != tt.want {
				t.Errorf("GetCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetHTTPStatus(t *testing.T) {
	tests := []struct {
		name string
		code string
		want int
	}{
		// 400 errors
		{"ErrBadRequest", ErrBadRequest, 400},
		{"ErrInvalidInput", ErrInvalidInput, 400},
		{"ErrValidation", ErrValidation, 400},
		{"ErrServiceInvalidInput", ErrServiceInvalidInput, 400},
		{"ErrServiceValidation", ErrServiceValidation, 400},

		// 401 errors
		{"ErrUnauthorized", ErrUnauthorized, 401},
		{"ErrInvalidToken", ErrInvalidToken, 401},
		{"ErrExpiredToken", ErrExpiredToken, 401},
		{"ErrRevokedToken", ErrRevokedToken, 401},
		{"ErrServiceUnauthorized", ErrServiceUnauthorized, 401},

		// 403 errors
		{"ErrForbidden", ErrForbidden, 403},
		{"ErrPermissionDenied", ErrPermissionDenied, 403},
		{"ErrDisabledAccount", ErrDisabledAccount, 403},
		{"ErrServiceForbidden", ErrServiceForbidden, 403},

		// 404 errors
		{"ErrNotFound", ErrNotFound, 404},
		{"ErrUserNotFound", ErrUserNotFound, 404},
		{"ErrSessionNotFound", ErrSessionNotFound, 404},
		{"ErrServiceNotFound", ErrServiceNotFound, 404},
		{"ErrBucketNotFound", ErrBucketNotFound, 404},
		{"ErrFileNotFound", ErrFileNotFound, 404},

		// 409 errors
		{"ErrConflict", ErrConflict, 409},
		{"ErrAlreadyExists", ErrAlreadyExists, 409},
		{"ErrServiceAlreadyExists", ErrServiceAlreadyExists, 409},
		{"ErrServiceConflict", ErrServiceConflict, 409},

		// 500 errors
		{"ErrInternal", ErrInternal, 500},
		{"ErrDatabase", ErrDatabase, 500},
		{"ErrUnknown", ErrUnknown, 500},
		{"ErrServiceUnknown", ErrServiceUnknown, 500},
		{"ErrServiceDependency", ErrServiceDependency, 500},

		// 504 errors
		{"ErrTimeout", ErrTimeout, 504},

		// Default
		{"Unknown error", "UNKNOWN_ERROR", 500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getHTTPStatus(tt.code); got != tt.want {
				t.Errorf("getHTTPStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetUserMessage(t *testing.T) {
	tests := []struct {
		name string
		code string
		want string
	}{
		// Common messages
		{"ErrBadRequest", ErrBadRequest, "Bad request"},
		{"ErrInvalidInput", ErrInvalidInput, "Invalid input provided"},
		{"ErrValidation", ErrValidation, "Validation failed"},

		// Auth messages
		{"ErrUnauthorized", ErrUnauthorized, "Authentication required"},
		{"ErrInvalidToken", ErrInvalidToken, "Invalid token"},
		{"ErrExpiredToken", ErrExpiredToken, "Token has expired"},
		{"ErrRevokedToken", ErrRevokedToken, "Token has been revoked"},
		{"ErrForbidden", ErrForbidden, "Access denied"},
		{"ErrPermissionDenied", ErrPermissionDenied, "You don't have permission to perform this action"},
		{"ErrDisabledAccount", ErrDisabledAccount, "Account is disabled"},

		// Resource messages
		{"ErrNotFound", ErrNotFound, "Not found"},
		{"ErrUserNotFound", ErrUserNotFound, "User not found"},
		{"ErrSessionNotFound", ErrSessionNotFound, "Session not found"},
		{"ErrConflict", ErrConflict, "Resource already exists"},
		{"ErrAlreadyExists", ErrAlreadyExists, "Already exists"},

		// System messages
		{"ErrInternal", ErrInternal, "Internal error"},
		{"ErrDatabase", ErrDatabase, "Database error"},
		{"ErrTimeout", ErrTimeout, "Request timeout"},
		{"ErrUnknown", ErrUnknown, "Unknown error"},

		// Default
		{"Unknown error", "UNKNOWN_ERROR", "Unknown error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getUserMessage(tt.code); got != tt.want {
				t.Errorf("getUserMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetNumericCode(t *testing.T) {
	tests := []struct {
		name string
		code string
		want string
	}{
		// Legacy codes
		{"ErrBadRequest", ErrBadRequest, CodeBadRequest},
		{"ErrUnauthorized", ErrUnauthorized, CodeUnauthorized},
		{"ErrNotFound", ErrNotFound, CodeNotFound},

		// Repository codes
		{"ErrRepoNotFound", ErrRepoNotFound, CodeRepoNotFound},
		{"ErrRepoDatabase", ErrRepoDatabase, CodeRepoDatabase},

		// Service codes
		{"ErrServiceNotFound", ErrServiceNotFound, CodeServiceNotFound},
		{"ErrServiceValidation", ErrServiceValidation, CodeServiceValidation},

		// Handler codes
		{"ErrHandlerNotFound", ErrHandlerNotFound, CodeHandlerNotFound},
		{"ErrHandlerBadRequest", ErrHandlerBadRequest, CodeHandlerBadRequest},

		// Default
		{"Unknown error", "UNKNOWN_ERROR", "1019"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetNumericCode(tt.code); got != tt.want {
				t.Errorf("GetNumericCode() = %v, want %v", got, tt.want)
			}
		})
	}
}
