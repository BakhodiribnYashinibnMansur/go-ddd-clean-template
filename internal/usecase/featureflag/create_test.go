package featureflag_test

import (
	"errors"
	"testing"

	"gct/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUseCase_Create(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		req       domain.CreateFeatureFlagRequest
		mockSetup func(*MockRepo)
		wantErr   bool
	}{
		{
			name: "success",
			req: domain.CreateFeatureFlagRequest{
				Key:         "enable_dark_mode",
				Name:        "Dark Mode",
				Type:        "boolean",
				Value:       "true",
				Description: "Enable dark mode for users",
				IsActive:    true,
			},
			mockSetup: func(m *MockRepo) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*domain.FeatureFlag")).
					Run(func(args mock.Arguments) {
						f := args.Get(1).(*domain.FeatureFlag)
						assert.NotEmpty(t, f.ID)
						assert.Equal(t, "enable_dark_mode", f.Key)
						assert.Equal(t, "Dark Mode", f.Name)
						assert.Equal(t, "boolean", f.Type)
						assert.Equal(t, "true", f.Value)
						assert.Equal(t, "Enable dark mode for users", f.Description)
						assert.True(t, f.IsActive)
					}).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "repo error",
			req: domain.CreateFeatureFlagRequest{
				Key:         "enable_dark_mode",
				Name:        "Dark Mode",
				Type:        "boolean",
				Value:       "true",
				Description: "Enable dark mode for users",
				IsActive:    false,
			},
			mockSetup: func(m *MockRepo) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*domain.FeatureFlag")).
					Return(errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uc, repo := setup(t)
			tt.mockSetup(repo)

			result, err := uc.Create(t.Context(), tt.req)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.req.Key, result.Key)
				assert.Equal(t, tt.req.Name, result.Name)
				assert.Equal(t, tt.req.Type, result.Type)
				assert.Equal(t, tt.req.Value, result.Value)
				assert.Equal(t, tt.req.Description, result.Description)
				assert.Equal(t, tt.req.IsActive, result.IsActive)
			}

			repo.AssertExpectations(t)
		})
	}
}
