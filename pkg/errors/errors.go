package errors

import (
	"context"
	"errors"
	"fmt"
	"runtime"
)

// AppError custom error strukturasi
type AppError struct {
	Type       string         // Error type (masalan: "USER_NOT_FOUND")
	Code       string         // Numeric error code (masalan: "4041")
	Message    string         // Developer message
	HTTPStatus int            // HTTP status code
	UserMsg    string         // Foydalanuvchi uchun xabar
	Details    string         // Batafsil tushuntirish
	Fields     map[string]any // Qo'shimcha ma'lumotlar
	Err        error          // Wrapped error
	Stack      []uintptr      // Stack trace
}

// Error interface implementation
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap wrapped errorni qaytaradi
func (e *AppError) Unwrap() error {
	return e.Err
}

// WithField field qo'shadi
func (e *AppError) WithField(key string, value any) *AppError {
	if e.Fields == nil {
		e.Fields = make(map[string]any)
	}
	e.Fields[key] = value
	return e
}

// WithDetails batafsil ma'lumot qo'shadi
func (e *AppError) WithDetails(details string) *AppError {
	e.Details = details
	return e
}

// New creates new error
func New(ctx context.Context, code string, message string) *AppError {
	return &AppError{
		Type:       code,
		Code:       getNumericCode(code),
		Message:    message,
		HTTPStatus: getHTTPStatus(code),
		UserMsg:    getUserMessage(code),
		Stack:      captureStack(),
	}
}

// Wrap mavjud errorni wrap qiladi
func Wrap(ctx context.Context, err error, code string, message string) *AppError {
	if err == nil {
		return nil
	}

	return &AppError{
		Type:       code,
		Code:       getNumericCode(code),
		Message:    message,
		HTTPStatus: getHTTPStatus(code),
		UserMsg:    getUserMessage(code),
		Err:        err,
		Stack:      captureStack(),
	}
}

// Is checks if error matches the code
func Is(err error, code string) bool {
	var e *AppError
	if errors.As(err, &e) {
		return e.Type == code
	}
	return false
}

// GetCode returns error code from error
func GetCode(err error) string {
	var e *AppError
	if errors.As(err, &e) {
		return e.Type
	}
	return ""
}

// captureStack stack trace yaratadi
func captureStack() []uintptr {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	return pcs[0:n]
}

// getHTTPStatus kod bo'yicha HTTP status qaytaradi
func getHTTPStatus(code string) int {
	switch code {
	// 400 errors
	case ErrBadRequest, ErrInvalidInput, ErrValidation:
		return 400

	// 401 errors
	case ErrUnauthorized, ErrInvalidToken, ErrExpiredToken, ErrRevokedToken:
		return 401

	// 403 errors
	case ErrForbidden, ErrPermissionDenied, ErrDisabledAccount:
		return 403

	// 404 errors
	case ErrNotFound, ErrUserNotFound, ErrSessionNotFound:
		return 404

	// 409 errors
	case ErrConflict, ErrAlreadyExists:
		return 409

	// 500 errors
	case ErrInternal, ErrDatabase, ErrUnknown:
		return 500

	// 504 errors
	case ErrTimeout:
		return 504

	default:
		return 500
	}
}

// getUserMessage kod bo'yicha foydalanuvchi xabarini qaytaradi
func getUserMessage(code string) string {
	switch code {
	case ErrBadRequest:
		return "Noto'g'ri so'rov"
	case ErrInvalidInput:
		return "Noto'g'ri ma'lumot kiritilgan"
	case ErrValidation:
		return "Ma'lumotlar validatsiyadan o'tmadi"
	case ErrUnauthorized:
		return "Autentifikatsiya talab qilinadi"
	case ErrInvalidToken:
		return "Noto'g'ri token"
	case ErrExpiredToken:
		return "Token muddati tugagan"
	case ErrRevokedToken:
		return "Token bekor qilingan"
	case ErrForbidden:
		return "Ruxsat yo'q"
	case ErrPermissionDenied:
		return "Ushbu amalni bajarish uchun ruxsatingiz yo'q"
	case ErrDisabledAccount:
		return "Akkaunt faol emas"
	case ErrNotFound:
		return "Topilmadi"
	case ErrUserNotFound:
		return "Foydalanuvchi topilmadi"
	case ErrSessionNotFound:
		return "Sessiya topilmadi"
	case ErrConflict:
		return "Ma'lumot allaqachon mavjud"
	case ErrAlreadyExists:
		return "Allaqachon mavjud"
	case ErrInternal:
		return "Ichki xatolik"
	case ErrDatabase:
		return "Ma'lumotlar bazasi xatolik"
	case ErrTimeout:
		return "Vaqt tugadi"
	case ErrUnknown:
		return "Noma'lum xatolik"
	default:
		return "Xatolik yuz berdi"
	}
}

// getNumericCode returns numeric code by error type
func getNumericCode(code string) string {
	// Repository layer codes (2xxx)
	switch code {
	case ErrRepoNotFound:
		return CodeRepoNotFound
	case ErrRepoAlreadyExists:
		return CodeRepoAlreadyExists
	case ErrRepoDatabase:
		return CodeRepoDatabase
	case ErrRepoTimeout:
		return CodeRepoTimeout
	case ErrRepoConnection:
		return CodeRepoConnection
	case ErrRepoTransaction:
		return CodeRepoTransaction
	case ErrRepoConstraint:
		return CodeRepoConstraint
	case ErrRepoUnknown:
		return CodeRepoUnknown
	}

	// Service layer codes (3xxx)
	switch code {
	case ErrServiceInvalidInput:
		return CodeServiceInvalidInput
	case ErrServiceValidation:
		return CodeServiceValidation
	case ErrServiceNotFound:
		return CodeServiceNotFound
	case ErrServiceAlreadyExists:
		return CodeServiceAlreadyExists
	case ErrServiceUnauthorized:
		return CodeServiceUnauthorized
	case ErrServiceForbidden:
		return CodeServiceForbidden
	case ErrServiceConflict:
		return CodeServiceConflict
	case ErrServiceBusinessRule:
		return CodeServiceBusinessRule
	case ErrServiceDependency:
		return CodeServiceDependency
	case ErrServiceUnknown:
		return CodeServiceUnknown
	}

	// Handler layer codes (4xxx, 5xxx)
	switch code {
	case ErrHandlerBadRequest:
		return CodeHandlerBadRequest
	case ErrHandlerUnauthorized:
		return CodeHandlerUnauthorized
	case ErrHandlerForbidden:
		return CodeHandlerForbidden
	case ErrHandlerNotFound:
		return CodeHandlerNotFound
	case ErrHandlerConflict:
		return CodeHandlerConflict
	case ErrHandlerInternal:
		return CodeHandlerInternal
	case ErrHandlerUnknown:
		return CodeHandlerUnknown
	}

	// Old/legacy codes
	switch code {
	case ErrBadRequest:
		return CodeBadRequest
	case ErrInvalidInput:
		return CodeInvalidInput
	case ErrValidation:
		return CodeValidation
	case ErrUnauthorized:
		return CodeUnauthorized
	case ErrInvalidToken:
		return CodeInvalidToken
	case ErrExpiredToken:
		return CodeExpiredToken
	case ErrRevokedToken:
		return CodeRevokedToken
	case ErrForbidden:
		return CodeForbidden
	case ErrPermissionDenied:
		return CodePermissionDenied
	case ErrDisabledAccount:
		return CodeDisabledAccount
	case ErrNotFound:
		return CodeNotFound
	case ErrUserNotFound:
		return CodeUserNotFound
	case ErrSessionNotFound:
		return CodeSessionNotFound
	case ErrConflict:
		return CodeConflict
	case ErrAlreadyExists:
		return CodeAlreadyExists
	case ErrInternal:
		return CodeInternal
	case ErrDatabase:
		return CodeDatabase
	case ErrTimeout:
		return CodeTimeout
	case ErrUnknown:
		return CodeUnknown
	default:
		return "9999"
	}
}
