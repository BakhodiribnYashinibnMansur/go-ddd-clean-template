package validator

import (
	"testing"

	apperrors "gct/internal/kernel/infrastructure/errorx"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testStruct struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age" validate:"gte=0,lte=150"`
}

type testPhoneStruct struct {
	Phone string `json:"phone" validate:"required,phone"`
}

type testPasswordStruct struct {
	Password string `json:"password" validate:"required,strong_password"`
}

func TestValidateStruct_Success(t *testing.T) {
	s := testStruct{
		Name:  "Alice",
		Email: "alice@example.com",
		Age:   30,
	}
	err := ValidateStruct(s)
	assert.NoError(t, err)
}

func TestValidateStruct_MissingRequired(t *testing.T) {
	s := testStruct{
		Name:  "",
		Email: "alice@example.com",
		Age:   25,
	}
	err := ValidateStruct(s)
	require.Error(t, err)

	var appErr *apperrors.AppError
	assert.ErrorAs(t, err, &appErr)
	assert.Contains(t, appErr.Message, "Validation failed")
}

func TestValidateStruct_InvalidEmail(t *testing.T) {
	s := testStruct{
		Name:  "Bob",
		Email: "not-an-email",
		Age:   25,
	}
	err := ValidateStruct(s)
	require.Error(t, err)

	var appErr *apperrors.AppError
	assert.ErrorAs(t, err, &appErr)
	assert.Contains(t, appErr.Message, "email")
}

func TestValidateStruct_AgeOutOfRange(t *testing.T) {
	s := testStruct{
		Name:  "Charlie",
		Email: "charlie@example.com",
		Age:   -1,
	}
	err := ValidateStruct(s)
	require.Error(t, err)
}

func TestValidateStruct_MultipleErrors(t *testing.T) {
	s := testStruct{
		Name:  "",
		Email: "bad",
		Age:   200,
	}
	err := ValidateStruct(s)
	require.Error(t, err)

	var appErr *apperrors.AppError
	assert.ErrorAs(t, err, &appErr)
	// Should contain multiple validation errors
	assert.Contains(t, appErr.Message, "Validation failed")
}

func TestValidateStruct_NilInput(t *testing.T) {
	// Passing a non-struct should cause an internal error
	err := ValidateStruct("not a struct")
	require.Error(t, err)
}

func TestValidateStruct_Pointer(t *testing.T) {
	s := &testStruct{
		Name:  "Diana",
		Email: "diana@example.com",
		Age:   28,
	}
	err := ValidateStruct(s)
	assert.NoError(t, err)
}

func TestValidateStruct_JsonTagUsedAsFieldName(t *testing.T) {
	s := testStruct{
		Name:  "",
		Email: "valid@example.com",
		Age:   25,
	}
	err := ValidateStruct(s)
	require.Error(t, err)

	var appErr *apperrors.AppError
	assert.ErrorAs(t, err, &appErr)
	// The error message should use the json tag name "name", not "Name"
	assert.Contains(t, appErr.Message, "name")
}

func TestValidateStruct_CustomPhoneValidation(t *testing.T) {
	tests := []struct {
		name    string
		phone   string
		wantErr bool
	}{
		{"valid_phone", "+12025551234", false},
		{"invalid_phone", "abc", true},
		{"empty_phone", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testPhoneStruct{Phone: tt.phone}
			err := ValidateStruct(s)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateStruct_CustomPasswordValidation(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"strong_password", "MyStr0ng!Pass", false},
		{"weak_password", "weak", true},
		{"empty_password", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testPasswordStruct{Password: tt.password}
			err := ValidateStruct(s)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
