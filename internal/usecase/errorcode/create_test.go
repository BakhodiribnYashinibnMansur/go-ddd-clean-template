package errorcode_test

import (
	"errors"
	"testing"
	"time"

	"gct/internal/domain"
	repo "gct/internal/repo/persistent/postgres/errorcode"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     repo.CreateErrorCodeInput
		mockRes   *domain.ErrorCode
		mockErr   error
		wantErr   bool
		errSubstr string
	}{
		{
			name: "success",
			input: repo.CreateErrorCodeInput{
				Code:       "ERR_001",
				Message:    "Something went wrong",
				HTTPStatus: 500,
				Category:   domain.CategorySystem,
				Severity:   domain.SeverityHigh,
				Retryable:  true,
				RetryAfter: 30,
				Suggestion: "Try again later",
			},
			mockRes: &domain.ErrorCode{
				ID:         uuid.New(),
				Code:       "ERR_001",
				Message:    "Something went wrong",
				HTTPStatus: 500,
				Category:   domain.CategorySystem,
				Severity:   domain.SeverityHigh,
				Retryable:  true,
				RetryAfter: 30,
				Suggestion: "Try again later",
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
			mockErr: nil,
			wantErr: false,
		},
		{
			name: "repo_error",
			input: repo.CreateErrorCodeInput{
				Code:       "ERR_DUP",
				Message:    "Duplicate",
				HTTPStatus: 409,
				Category:   domain.CategoryValidation,
				Severity:   domain.SeverityLow,
			},
			mockRes:   nil,
			mockErr:   errors.New("unique violation"),
			wantErr:   true,
			errSubstr: "unique violation",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			uc, mockRepo := setup(t)

			mockRepo.On("Create", ctx, tc.input).Return(tc.mockRes, tc.mockErr)

			result, err := uc.Create(ctx, tc.input)

			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), tc.errSubstr)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tc.input.Code, result.Code)
				assert.Equal(t, tc.input.Message, result.Message)
				assert.Equal(t, tc.input.HTTPStatus, result.HTTPStatus)
				assert.Equal(t, tc.input.Category, result.Category)
				assert.Equal(t, tc.input.Severity, result.Severity)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}
