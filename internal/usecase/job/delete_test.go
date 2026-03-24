package job_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUseCase_Delete(t *testing.T) {
	t.Parallel()

	id := uuid.New()

	tests := []struct {
		name      string
		id        uuid.UUID
		mockSetup func(*MockRepo)
		wantErr   bool
	}{
		{
			name: "success",
			id:   id,
			mockSetup: func(m *MockRepo) {
				m.On("Delete", mock.Anything, id).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "repo error",
			id:   id,
			mockSetup: func(m *MockRepo) {
				m.On("Delete", mock.Anything, id).Return(errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uc, repo := setup(t)
			tt.mockSetup(repo)

			err := uc.Delete(t.Context(), tt.id)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			repo.AssertExpectations(t)
		})
	}
}
