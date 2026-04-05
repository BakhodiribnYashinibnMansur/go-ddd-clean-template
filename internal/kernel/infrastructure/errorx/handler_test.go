package errorx

import (
	"errors"
	"testing"
)

func TestNewHandlerError(t *testing.T) {

	tests := []struct {
		name    string
		code    string
		message string
		want    *AppError
	}{
		{
			name:    "create new handler error",
			code:    ErrHandlerNotFound,
			message: "Resource not found",
			want: &AppError{
				Type:       ErrHandlerNotFound,
				Code:       CodeHandlerNotFound,
				Message:    "Resource not found",
				HTTPStatus: 404,
				UserMsg:    "Not found",
			},
		},
		{
			name:    "create handler bad request error",
			code:    ErrHandlerBadRequest,
			message: "Bad request",
			want: &AppError{
				Type:       ErrHandlerBadRequest,
				Code:       CodeHandlerBadRequest,
				Message:    "Bad request",
				HTTPStatus: 400,
				UserMsg:    "Bad request",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewHandlerError(tt.code, tt.message)

			if got.Type != tt.want.Type {
				t.Errorf("NewHandlerError().Type = %v, want %v", got.Type, tt.want.Type)
			}
			if got.Code != tt.want.Code {
				t.Errorf("NewHandlerError().Code = %v, want %v", got.Code, tt.want.Code)
			}
			if got.Message != tt.want.Message {
				t.Errorf("NewHandlerError().Message = %v, want %v", got.Message, tt.want.Message)
			}
			if got.HTTPStatus != tt.want.HTTPStatus {
				t.Errorf("NewHandlerError().HTTPStatus = %v, want %v", got.HTTPStatus, tt.want.HTTPStatus)
			}
			if got.UserMsg != tt.want.UserMsg {
				t.Errorf("NewHandlerError().UserMsg = %v, want %v", got.UserMsg, tt.want.UserMsg)
			}
		})
	}
}

func TestWrapHandlerError(t *testing.T) {
	baseErr := errors.New("handler error")

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
			code:    ErrHandlerInternal,
			message: "Internal server error",
			want: &AppError{
				Type:       ErrHandlerInternal,
				Code:       CodeHandlerInternal,
				Message:    "Internal server error",
				HTTPStatus: 500,
				UserMsg:    "Internal error",
				Err:        baseErr,
			},
			wantNil: false,
		},
		{
			name:    "wrap nil error",
			err:     nil,
			code:    ErrHandlerInternal,
			message: "Internal server error",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WrapHandlerError(tt.err, tt.code, tt.message)

			if tt.wantNil {
				if got != nil {
					t.Errorf("WrapHandlerError() = %v, want nil", got)
				}
				return
			}

			if got.Type != tt.want.Type {
				t.Errorf("WrapHandlerError().Type = %v, want %v", got.Type, tt.want.Type)
			}
			if got.Code != tt.want.Code {
				t.Errorf("WrapHandlerError().Code = %v, want %v", got.Code, tt.want.Code)
			}
			if got.Message != tt.want.Message {
				t.Errorf("WrapHandlerError().Message = %v, want %v", got.Message, tt.want.Message)
			}
			if got.HTTPStatus != tt.want.HTTPStatus {
				t.Errorf("WrapHandlerError().HTTPStatus = %v, want %v", got.HTTPStatus, tt.want.HTTPStatus)
			}
			if got.UserMsg != tt.want.UserMsg {
				t.Errorf("WrapHandlerError().UserMsg = %v, want %v", got.UserMsg, tt.want.UserMsg)
			}
			if !errors.Is(got.Err, tt.want.Err) {
				t.Errorf("WrapHandlerError().Err = %v, want %v", got.Err, tt.want.Err)
			}
		})
	}
}

func TestMapServiceToHandlerError(t *testing.T) {

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
			name:     "service not found error",
			err:      NewServiceError(ErrServiceNotFound, "Resource not found"),
			wantType: ErrHandlerNotFound,
			wantCode: CodeHandlerNotFound,
		},
		{
			name:     "service invalid input error",
			err:      NewServiceError(ErrServiceInvalidInput, "Invalid input"),
			wantType: ErrHandlerBadRequest,
			wantCode: CodeHandlerBadRequest,
		},
		{
			name:     "service validation error",
			err:      NewServiceError(ErrServiceValidation, "Validation failed"),
			wantType: ErrHandlerBadRequest,
			wantCode: CodeHandlerBadRequest,
		},
		{
			name:     "service unauthorized error",
			err:      NewServiceError(ErrServiceUnauthorized, "Unauthorized"),
			wantType: ErrHandlerUnauthorized,
			wantCode: CodeHandlerUnauthorized,
		},
		{
			name:     "service forbidden error",
			err:      NewServiceError(ErrServiceForbidden, "Forbidden"),
			wantType: ErrHandlerForbidden,
			wantCode: CodeHandlerForbidden,
		},
		{
			name:     "service conflict error",
			err:      NewServiceError(ErrServiceConflict, "Conflict"),
			wantType: ErrHandlerConflict,
			wantCode: CodeHandlerConflict,
		},
		{
			name:     "service already exists error",
			err:      NewServiceError(ErrServiceAlreadyExists, "Already exists"),
			wantType: ErrHandlerConflict,
			wantCode: CodeHandlerConflict,
		},
		{
			name:     "service unknown error",
			err:      NewServiceError(ErrServiceUnknown, "Unknown error"),
			wantType: ErrHandlerInternal,
			wantCode: CodeHandlerInternal,
		},
		{
			name:     "non-AppError",
			err:      errors.New("regular error"),
			wantType: ErrHandlerInternal,
			wantCode: CodeHandlerInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapServiceToHandlerError(tt.err)

			if tt.wantType == "" {
				if got != nil {
					t.Errorf("MapServiceToHandlerError() = %v, want nil", got)
				}
				return
			}

			if got.Type != tt.wantType {
				t.Errorf("MapServiceToHandlerError().Type = %v, want %v", got.Type, tt.wantType)
			}
			if got.Code != tt.wantCode {
				t.Errorf("MapServiceToHandlerError().Code = %v, want %v", got.Code, tt.wantCode)
			}
		})
	}
}

func TestMapToHTTPStatus(t *testing.T) {
	tests := []struct {
		name string
		code string
		want int
	}{
		// 400 errors
		{"ErrHandlerBadRequest", ErrHandlerBadRequest, 400},
		{"ErrServiceInvalidInput", ErrServiceInvalidInput, 400},
		{"ErrServiceValidation", ErrServiceValidation, 400},

		// 401 errors
		{"ErrHandlerUnauthorized", ErrHandlerUnauthorized, 401},
		{"ErrServiceUnauthorized", ErrServiceUnauthorized, 401},

		// 403 errors
		{"ErrHandlerForbidden", ErrHandlerForbidden, 403},
		{"ErrServiceForbidden", ErrServiceForbidden, 403},

		// 404 errors
		{"ErrHandlerNotFound", ErrHandlerNotFound, 404},
		{"ErrServiceNotFound", ErrServiceNotFound, 404},
		{"ErrRepoNotFound", ErrRepoNotFound, 404},

		// 409 errors
		{"ErrHandlerConflict", ErrHandlerConflict, 409},
		{"ErrServiceConflict", ErrServiceConflict, 409},
		{"ErrServiceAlreadyExists", ErrServiceAlreadyExists, 409},
		{"ErrRepoAlreadyExists", ErrRepoAlreadyExists, 409},

		// 500 errors
		{"ErrHandlerInternal", ErrHandlerInternal, 500},
		{"ErrServiceUnknown", ErrServiceUnknown, 500},
		{"ErrRepoDatabase", ErrRepoDatabase, 500},
		{"ErrRepoUnknown", ErrRepoUnknown, 500},

		// 504 errors
		{"ErrRepoTimeout", ErrRepoTimeout, 504},

		// Default
		{"Unknown error", "UNKNOWN_ERROR", 500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MapToHTTPStatus(tt.code); got != tt.want {
				t.Errorf("MapToHTTPStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandlerMessages(t *testing.T) {
	tests := []struct {
		name string
		code string
		want string
	}{
		{"ErrHandlerBadRequest", ErrHandlerBadRequest, "Bad request"},
		{"ErrHandlerUnauthorized", ErrHandlerUnauthorized, "Unauthorized access"},
		{"ErrHandlerForbidden", ErrHandlerForbidden, "Forbidden access"},
		{"ErrHandlerNotFound", ErrHandlerNotFound, "Resource not found"},
		{"ErrHandlerConflict", ErrHandlerConflict, "Resource conflict"},
		{"ErrHandlerInternal", ErrHandlerInternal, "Internal server error"},
		{"ErrHandlerUnknown", ErrHandlerUnknown, "Unknown error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := handlerMessages[tt.code]; got != tt.want {
				t.Errorf("handlerMessages[%s] = %v, want %v", tt.code, got, tt.want)
			}
		})
	}
}
