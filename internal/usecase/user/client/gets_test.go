package client_test

import (
	"errors"
	"testing"

	"gct/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUseCase_Gets_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		filter         *domain.UsersFilter
		mockUsers      []*domain.User
		mockTotal      int
		repoError      error
		expectError    bool
		validateResult func(t *testing.T, users []*domain.User, count int)
	}{
		{
			name: "success_with_pagination",
			filter: &domain.UsersFilter{
				Pagination: &domain.Pagination{
					Limit:  10,
					Offset: 0,
				},
			},
			mockUsers: []*domain.User{
				{ID: uuid.MustParse("00000000-0000-0000-0000-000000000001"), Phone: stringPtr("111111111")},
				{ID: uuid.MustParse("00000000-0000-0000-0000-000000000002"), Phone: stringPtr("222222222")},
			},
			mockTotal:   2,
			repoError:   nil,
			expectError: false,
			validateResult: func(t *testing.T, users []*domain.User, count int) {
				require.Len(t, users, 2)
				require.Equal(t, 2, count)
				require.Equal(t, uuid.MustParse("00000000-0000-0000-0000-000000000001"), users[0].ID)
				require.Equal(t, uuid.MustParse("00000000-0000-0000-0000-000000000002"), users[1].ID)
			},
		},
		{
			name: "success_empty_result",
			filter: &domain.UsersFilter{
				Pagination: &domain.Pagination{
					Limit:  10,
					Offset: 0,
				},
			},
			mockUsers:   []*domain.User{},
			mockTotal:   0,
			repoError:   nil,
			expectError: false,
			validateResult: func(t *testing.T, users []*domain.User, count int) {
				require.Len(t, users, 0)
				require.Equal(t, 0, count)
			},
		},
		{
			name: "success_with_offset",
			filter: &domain.UsersFilter{
				Pagination: &domain.Pagination{
					Limit:  5,
					Offset: 10,
				},
			},
			mockUsers: []*domain.User{
				{ID: uuid.MustParse("00000000-0000-0000-0000-00000000000b"), Phone: stringPtr("111111111")},
				{ID: uuid.MustParse("00000000-0000-0000-0000-00000000000c"), Phone: stringPtr("222222222")},
				{ID: uuid.MustParse("00000000-0000-0000-0000-00000000000d"), Phone: stringPtr("333333333")},
			},
			mockTotal:   25,
			repoError:   nil,
			expectError: false,
			validateResult: func(t *testing.T, users []*domain.User, count int) {
				require.Len(t, users, 3)
				require.Equal(t, 25, count)
				require.Equal(t, uuid.MustParse("00000000-0000-0000-0000-00000000000b"), users[0].ID)
			},
		},
		{
			name: "success_nil_pagination",
			filter: &domain.UsersFilter{
				Pagination: nil,
			},
			mockUsers: []*domain.User{
				{ID: uuid.MustParse("00000000-0000-0000-0000-000000000001"), Phone: stringPtr("111111111")},
			},
			mockTotal:   1,
			repoError:   nil,
			expectError: false,
			validateResult: func(t *testing.T, users []*domain.User, count int) {
				require.Len(t, users, 1)
				require.Equal(t, 1, count)
			},
		},
		{
			name: "error_repo_returns_error",
			filter: &domain.UsersFilter{
				Pagination: &domain.Pagination{
					Limit:  10,
					Offset: 0,
				},
			},
			mockUsers:   nil,
			mockTotal:   0,
			repoError:   errors.New("database error"),
			expectError: true,
			validateResult: func(t *testing.T, users []*domain.User, count int) {
				// No validation expected on error
			},
		},
		{
			name: "success_large_limit",
			filter: &domain.UsersFilter{
				Pagination: &domain.Pagination{
					Limit:  100,
					Offset: 0,
				},
			},
			mockUsers:   make([]*domain.User, 50),
			mockTotal:   150,
			repoError:   nil,
			expectError: false,
			validateResult: func(t *testing.T, users []*domain.User, count int) {
				require.Len(t, users, 50)
				require.Equal(t, 150, count)
			},
		},
		{
			name: "success_zero_limit",
			filter: &domain.UsersFilter{
				Pagination: &domain.Pagination{
					Limit:  0,
					Offset: 0,
				},
			},
			mockUsers:   []*domain.User{},
			mockTotal:   0,
			repoError:   nil,
			expectError: false,
			validateResult: func(t *testing.T, users []*domain.User, count int) {
				require.Len(t, users, 0)
				require.Equal(t, 0, count)
			},
		},
		{
			name: "success_with_usernames",
			filter: &domain.UsersFilter{
				Pagination: &domain.Pagination{
					Limit:  10,
					Offset: 0,
				},
			},
			mockUsers: []*domain.User{
				{ID: uuid.MustParse("00000000-0000-0000-0000-000000000001"), Phone: stringPtr("111111111"), Username: stringPtr("user1")},
				{ID: uuid.MustParse("00000000-0000-0000-0000-000000000002"), Phone: stringPtr("222222222"), Username: stringPtr("user2")},
				{ID: uuid.MustParse("00000000-0000-0000-0000-000000000003"), Phone: stringPtr("333333333"), Username: nil},
			},
			mockTotal:   3,
			repoError:   nil,
			expectError: false,
			validateResult: func(t *testing.T, users []*domain.User, count int) {
				require.Len(t, users, 3)
				require.Equal(t, 3, count)
				require.NotNil(t, users[0].Username)
				require.Equal(t, "user1", *users[0].Username)
				require.Nil(t, users[2].Username)
			},
		},
	}

	for _, tt := range tests {
		tt := tt // parallel safety
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// arrange
			uc, clientRepo, _ := setup(t)
			ctx := t.Context()

			clientRepo.On("Gets", ctx, tt.filter).Return(tt.mockUsers, tt.mockTotal, tt.repoError).Once()

			// act
			users, count, err := uc.Gets(ctx, tt.filter)

			// assert
			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, users, count)
				}
			}

			clientRepo.AssertExpectations(t)
		})
	}
}
