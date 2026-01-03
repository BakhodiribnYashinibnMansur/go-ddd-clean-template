package errors

import (
	"context"
	"errors"
	"testing"
)

func TestNewServiceError(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		code    string
		message string
		want    *AppError
	}{
		{
			name:    "create new service error",
			code:    ErrServiceNotFound,
			message: "Resource not found",
			want: &AppError{
				Type:       ErrServiceNotFound,
				Code:       CodeServiceNotFound,
				Message:    "Resource not found",
				HTTPStatus: 404,
				UserMsg:    "Not found",
			},
		},
		{
			name:    "create service validation error",
			code:    ErrServiceValidation,
			message: "Validation failed",
			want: &AppError{
				Type:       ErrServiceValidation,
				Code:       CodeServiceValidation,
				Message:    "Validation failed",
				HTTPStatus: 400,
				UserMsg:    "Validation failed",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewServiceError(ctx, tt.code, tt.message)

			if got.Type != tt.want.Type {
				t.Errorf("NewServiceError().Type = %v, want %v", got.Type, tt.want.Type)
			}
			if got.Code != tt.want.Code {
				t.Errorf("NewServiceError().Code = %v, want %v", got.Code, tt.want.Code)
			}
			if got.Message != tt.want.Message {
				t.Errorf("NewServiceError().Message = %v, want %v", got.Message, tt.want.Message)
			}
			if got.HTTPStatus != tt.want.HTTPStatus {
				t.Errorf("NewServiceError().HTTPStatus = %v, want %v", got.HTTPStatus, tt.want.HTTPStatus)
			}
			if got.UserMsg != tt.want.UserMsg {
				t.Errorf("NewServiceError().UserMsg = %v, want %v", got.UserMsg, tt.want.UserMsg)
			}
		})
	}
}

func TestWrapServiceError(t *testing.T) {
	ctx := context.Background()
	baseErr := errors.New("service error")

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
			code:    ErrServiceUnknown,
			message: "Service operation failed",
			want: &AppError{
				Type:       ErrServiceUnknown,
				Code:       CodeServiceUnknown,
				Message:    "Service operation failed",
				HTTPStatus: 500,
				UserMsg:    "Unknown error",
				Err:        baseErr,
			},
			wantNil: false,
		},
		{
			name:    "wrap nil error",
			err:     nil,
			code:    ErrServiceUnknown,
			message: "Service operation failed",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WrapServiceError(ctx, tt.err, tt.code, tt.message)

			if tt.wantNil {
				if got != nil {
					t.Errorf("WrapServiceError() = %v, want nil", got)
				}
				return
			}

			if got.Type != tt.want.Type {
				t.Errorf("WrapServiceError().Type = %v, want %v", got.Type, tt.want.Type)
			}
			if got.Code != tt.want.Code {
				t.Errorf("WrapServiceError().Code = %v, want %v", got.Code, tt.want.Code)
			}
			if got.Message != tt.want.Message {
				t.Errorf("WrapServiceError().Message = %v, want %v", got.Message, tt.want.Message)
			}
			if got.HTTPStatus != tt.want.HTTPStatus {
				t.Errorf("WrapServiceError().HTTPStatus = %v, want %v", got.HTTPStatus, tt.want.HTTPStatus)
			}
			if got.UserMsg != tt.want.UserMsg {
				t.Errorf("WrapServiceError().UserMsg = %v, want %v", got.UserMsg, tt.want.UserMsg)
			}
			if got.Err != tt.want.Err {
				t.Errorf("WrapServiceError().Err = %v, want %v", got.Err, tt.want.Err)
			}
		})
	}
}

func TestMapRepoToServiceError(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		err      error
		wantType string
		wantCode string
	}{
		{
			name:     "nil error",
			err:      nil,
			wantType: "",
			wantCode: "",
		},
		{
			name:     "repo not found error",
			err:      NewRepoError(ctx, ErrRepoNotFound, "Record not found"),
			wantType: ErrServiceNotFound,
			wantCode: CodeServiceNotFound,
		},
		{
			name:     "repo already exists error",
			err:      NewRepoError(ctx, ErrRepoAlreadyExists, "Record already exists"),
			wantType: ErrServiceAlreadyExists,
			wantCode: CodeServiceAlreadyExists,
		},
		{
			name:     "repo constraint error",
			err:      NewRepoError(ctx, ErrRepoConstraint, "Constraint violation"),
			wantType: ErrServiceConflict,
			wantCode: CodeServiceConflict,
		},
		{
			name:     "repo database error",
			err:      NewRepoError(ctx, ErrRepoDatabase, "Database error"),
			wantType: ErrServiceDependency,
			wantCode: CodeServiceDependency,
		},
		{
			name:     "non-AppError",
			err:      errors.New("regular error"),
			wantType: ErrServiceUnknown,
			wantCode: CodeServiceUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapRepoToServiceError(ctx, tt.err)

			if tt.wantType == "" {
				if got != nil {
					t.Errorf("MapRepoToServiceError() = %v, want nil", got)
				}
				return
			}

			if got.Type != tt.wantType {
				t.Errorf("MapRepoToServiceError().Type = %v, want %v", got.Type, tt.wantType)
			}
			if got.Code != tt.wantCode {
				t.Errorf("MapRepoToServiceError().Code = %v, want %v", got.Code, tt.wantCode)
			}
		})
	}
}

func TestServiceMessages(t *testing.T) {
	tests := []struct {
		name string
		code string
		want string
	}{
		{"ErrServiceInvalidInput", ErrServiceInvalidInput, "Invalid input provided"},
		{"ErrServiceValidation", ErrServiceValidation, "Validation failed"},
		{"ErrServiceNotFound", ErrServiceNotFound, "Resource not found"},
		{"ErrServiceAlreadyExists", ErrServiceAlreadyExists, "Resource already exists"},
		{"ErrServiceUnauthorized", ErrServiceUnauthorized, "Authentication required"},
		{"ErrServiceForbidden", ErrServiceForbidden, "Permission denied"},
		{"ErrServiceConflict", ErrServiceConflict, "Resource conflict"},
		{"ErrServiceBusinessRule", ErrServiceBusinessRule, "Business rule violation"},
		{"ErrServiceDependency", ErrServiceDependency, "Dependency service error"},
		{"ErrServiceUnknown", ErrServiceUnknown, "Unknown service error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := serviceMessages[tt.code]; got != tt.want {
				t.Errorf("serviceMessages[%s] = %v, want %v", tt.code, got, tt.want)
			}
		})
	}
}
