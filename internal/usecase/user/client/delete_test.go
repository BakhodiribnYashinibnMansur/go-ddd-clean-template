package client_test

import (
	"errors"
	"testing"

	"gct/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
"github.com/stretchr/testify/require"
)

func TestUseCase_Delete_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		filter       *domain.UserFilter
		repoError    error
		expectError  bool
		validateRepo func(t *testing.T)
	}{
		{
			name: "success_valid_id",
			filter: &domain.UserFilter{
				ID: uuidPtr(uuid.MustParse("00000000-0000-0000-0000-000000000001")),
			},
			repoError:   nil,
			expectError: false,
			validateRepo: func(t *testing.T) {
				// Validated in test setup
			},
		},
		{
			name: "success_zero_id",
			filter: &domain.UserFilter{
				ID: uuidPtr(uuid.UUID{}), // zero UUID
			},
			repoError:   nil,
			expectError: false,
			validateRepo: func(t *testing.T) {
				// Validated in test setup
			},
		},
		{
			name: "success_negative_id",
			filter: &domain.UserFilter{
				ID: uuidPtr(uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff")), // max UUID equivalent
			},
			repoError:   nil,
			expectError: false,
			validateRepo: func(t *testing.T) {
				// Validated in test setup
			},
		},
		{
			name:        "error_null_filter",
			filter:      &domain.UserFilter{}, // IDs are null
			repoError:   nil,
			expectError: true,
			validateRepo: func(t *testing.T) {
				// No repo calls expected
			},
		},
		{
			name: "error_repo_returns_error",
			filter: &domain.UserFilter{
				ID: uuidPtr(uuid.MustParse("00000000-0000-0000-0000-000000000001")),
			},
			repoError:   errors.New("delete failed"),
			expectError: true,
			validateRepo: func(t *testing.T) {
				// Validated in test setup
			},
		},
		{
			name: "success_nil_id_pointer",
			filter: &domain.UserFilter{
				ID: nil,
			},
			repoError:   nil,
			expectError: true, // Should error with nil ID
			validateRepo: func(t *testing.T) {
				// No repo calls expected
			},
		},
	}

	for _, tt := range tests {
		// parallel safety
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// arrange
			uc, clientRepo, _ := setup(t)
			ctx := t.Context()

			if tt.filter.ID != nil {
				clientRepo.On("Delete", mock.Anything, *tt.filter.ID).Return(tt.repoError).Once()
			}

			// act
			err := uc.Delete(ctx, tt.filter)

			// assert
			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			clientRepo.AssertExpectations(t)
		})
	}
}

// Helper function to create UUID pointer
func uuidPtr(u uuid.UUID) *uuid.UUID {
	return &u
}
