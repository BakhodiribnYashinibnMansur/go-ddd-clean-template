package errors

import (
	"errors"
	"testing"
)

func TestNewRepoError(t *testing.T) {

	tests := []struct {
		name    string
		code    string
		message string
		want    *AppError
	}{
		{
			name:    "create new repository error",
			code:    ErrRepoNotFound,
			message: "Record not found",
			want: &AppError{
				Type:       ErrRepoNotFound,
				Code:       CodeRepoNotFound,
				Message:    "Record not found",
				HTTPStatus: 404,
				UserMsg:    "Not found",
			},
		},
		{
			name:    "create repository database error",
			code:    ErrRepoDatabase,
			message: "Database connection failed",
			want: &AppError{
				Type:       ErrRepoDatabase,
				Code:       CodeRepoDatabase,
				Message:    "Database connection failed",
				HTTPStatus: 500,
				UserMsg:    "Database error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewRepoError(tt.code, tt.message)

			if got.Type != tt.want.Type {
				t.Errorf("NewRepoError().Type = %v, want %v", got.Type, tt.want.Type)
			}
			if got.Code != tt.want.Code {
				t.Errorf("NewRepoError().Code = %v, want %v", got.Code, tt.want.Code)
			}
			if got.Message != tt.want.Message {
				t.Errorf("NewRepoError().Message = %v, want %v", got.Message, tt.want.Message)
			}
			if got.HTTPStatus != tt.want.HTTPStatus {
				t.Errorf("NewRepoError().HTTPStatus = %v, want %v", got.HTTPStatus, tt.want.HTTPStatus)
			}
			if got.UserMsg != tt.want.UserMsg {
				t.Errorf("NewRepoError().UserMsg = %v, want %v", got.UserMsg, tt.want.UserMsg)
			}
		})
	}
}

func TestWrapRepoError(t *testing.T) {
	baseErr := errors.New("database error")

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
			code:    ErrRepoDatabase,
			message: "Database operation failed",
			want: &AppError{
				Type:       ErrRepoDatabase,
				Code:       CodeRepoDatabase,
				Message:    "Database operation failed",
				HTTPStatus: 500,
				UserMsg:    "Database error",
				Err:        baseErr,
			},
			wantNil: false,
		},
		{
			name:    "wrap nil error",
			err:     nil,
			code:    ErrRepoDatabase,
			message: "Database operation failed",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WrapRepoError(tt.err, tt.code, tt.message)

			if tt.wantNil {
				if got != nil {
					t.Errorf("WrapRepoError() = %v, want nil", got)
				}
				return
			}

			if got.Type != tt.want.Type {
				t.Errorf("WrapRepoError().Type = %v, want %v", got.Type, tt.want.Type)
			}
			if got.Code != tt.want.Code {
				t.Errorf("WrapRepoError().Code = %v, want %v", got.Code, tt.want.Code)
			}
			if got.Message != tt.want.Message {
				t.Errorf("WrapRepoError().Message = %v, want %v", got.Message, tt.want.Message)
			}
			if got.HTTPStatus != tt.want.HTTPStatus {
				t.Errorf("WrapRepoError().HTTPStatus = %v, want %v", got.HTTPStatus, tt.want.HTTPStatus)
			}
			if got.UserMsg != tt.want.UserMsg {
				t.Errorf("WrapRepoError().UserMsg = %v, want %v", got.UserMsg, tt.want.UserMsg)
			}
			if !errors.Is(got.Err, tt.want.Err) {
				t.Errorf("WrapRepoError().Err = %v, want %v", got.Err, tt.want.Err)
			}
		})
	}
}

func TestRepoMessages(t *testing.T) {
	tests := []struct {
		name string
		code string
		want string
	}{
		{"ErrRepoNotFound", ErrRepoNotFound, "Record not found in database"},
		{"ErrRepoAlreadyExists", ErrRepoAlreadyExists, "Record already exists"},
		{"ErrRepoDatabase", ErrRepoDatabase, "Database error occurred"},
		{"ErrRepoTimeout", ErrRepoTimeout, "Database operation timeout"},
		{"ErrRepoConnection", ErrRepoConnection, "Database connection error"},
		{"ErrRepoTransaction", ErrRepoTransaction, "Transaction error"},
		{"ErrRepoConstraint", ErrRepoConstraint, "Database constraint violation"},
		{"ErrRepoUnknown", ErrRepoUnknown, "Unknown repository error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := repoMessages[tt.code]; got != tt.want {
				t.Errorf("repoMessages[%s] = %v, want %v", tt.code, got, tt.want)
			}
		})
	}
}
