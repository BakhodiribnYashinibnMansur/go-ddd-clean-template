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

func TestList(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name      string
		mockRes   []*domain.ErrorCode
		mockErr   error
		wantErr   bool
		wantLen   int
		errSubstr string
	}{
		{
			name: "success_multiple",
			mockRes: []*domain.ErrorCode{
				{
					ID:         uuid.New(),
					Code:       "ERR_001",
					Message:    "First error",
					HTTPStatus: 500,
					Category:   domain.CategorySystem,
					Severity:   domain.SeverityHigh,
					CreatedAt:  now,
					UpdatedAt:  now,
				},
				{
					ID:         uuid.New(),
					Code:       "ERR_002",
					Message:    "Second error",
					HTTPStatus: 400,
					Category:   domain.CategoryValidation,
					Severity:   domain.SeverityLow,
					CreatedAt:  now,
					UpdatedAt:  now,
				},
			},
			mockErr: nil,
			wantErr: false,
			wantLen: 2,
		},
		{
			name:    "success_empty",
			mockRes: []*domain.ErrorCode{},
			mockErr: nil,
			wantErr: false,
			wantLen: 0,
		},
		{
			name:      "repo_error",
			mockRes:   nil,
			mockErr:   errors.New("database connection lost"),
			wantErr:   true,
			errSubstr: "database connection lost",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			uc, mockRepo := setup(t)

			mockRepo.On("List", ctx).Return(tc.mockRes, tc.mockErr)

			result, err := uc.List(ctx)

			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), tc.errSubstr)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Len(t, result, tc.wantLen)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}
