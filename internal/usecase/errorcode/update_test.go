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

func TestUpdate(t *testing.T) {
	t.Parallel()

	newMsg := "Updated message"
	newStatus := 502

	tests := []struct {
		name      string
		code      string
		input     repo.UpdateErrorCodeInput
		mockRes   *domain.ErrorCode
		mockErr   error
		wantErr   bool
		errSubstr string
	}{
		{
			name: "success",
			code: "ERR_001",
			input: repo.UpdateErrorCodeInput{
				Message:    &newMsg,
				HTTPStatus: &newStatus,
			},
			mockRes: &domain.ErrorCode{
				ID:         uuid.New(),
				Code:       "ERR_001",
				Message:    newMsg,
				HTTPStatus: newStatus,
				Category:   domain.CategorySystem,
				Severity:   domain.SeverityHigh,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
			mockErr: nil,
			wantErr: false,
		},
		{
			name: "not_found",
			code: "ERR_MISSING",
			input: repo.UpdateErrorCodeInput{
				Message: &newMsg,
			},
			mockRes:   nil,
			mockErr:   errors.New("no rows in result set"),
			wantErr:   true,
			errSubstr: "no rows",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			uc, mockRepo := setup(t)

			mockRepo.On("Update", ctx, tc.code, tc.input).Return(tc.mockRes, tc.mockErr)

			result, err := uc.Update(ctx, tc.code, tc.input)

			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), tc.errSubstr)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tc.code, result.Code)
				assert.Equal(t, newMsg, result.Message)
				assert.Equal(t, newStatus, result.HTTPStatus)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}
