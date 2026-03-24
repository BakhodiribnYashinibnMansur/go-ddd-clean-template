package errorcode_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDelete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		code      string
		mockErr   error
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "success",
			code:    "ERR_001",
			mockErr: nil,
			wantErr: false,
		},
		{
			name:      "repo_error",
			code:      "ERR_MISSING",
			mockErr:   errors.New("delete failed"),
			wantErr:   true,
			errSubstr: "delete failed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			uc, mockRepo := setup(t)

			mockRepo.On("Delete", ctx, tc.code).Return(tc.mockErr)

			err := uc.Delete(ctx, tc.code)

			if tc.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errSubstr)
			} else {
				require.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}
