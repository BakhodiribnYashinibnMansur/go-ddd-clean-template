package errorcode_test

import (
	"errors"
	"testing"
	"time"

	"gct/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetByCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		code      string
		mockRes   *domain.ErrorCode
		mockErr   error
		wantErr   bool
		errSubstr string
	}{
		{
			name: "success",
			code: "ERR_001",
			mockRes: &domain.ErrorCode{
				ID:         uuid.New(),
				Code:       "ERR_001",
				Message:    "Something went wrong",
				HTTPStatus: 500,
				Category:   domain.CategorySystem,
				Severity:   domain.SeverityHigh,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
			mockErr: nil,
			wantErr: false,
		},
		{
			name:      "not_found",
			code:      "ERR_MISSING",
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

			mockRepo.On("GetByCode", ctx, tc.code).Return(tc.mockRes, tc.mockErr)

			result, err := uc.GetByCode(ctx, tc.code)

			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), tc.errSubstr)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tc.code, result.Code)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}
