package featureflag_test

import (
	"errors"
	"testing"
	"time"

	"gct/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUseCase_Update(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	now := time.Now()
	newName := "Updated Name"
	newValue := "false"
	newDesc := "Updated description"

	tests := []struct {
		name      string
		id        uuid.UUID
		req       domain.UpdateFeatureFlagRequest
		mockSetup func(*MockRepo)
		wantErr   bool
		check     func(*testing.T, *domain.FeatureFlag)
	}{
		{
			name: "success partial update",
			id:   id,
			req: domain.UpdateFeatureFlagRequest{
				Name:        &newName,
				Value:       &newValue,
				Description: &newDesc,
			},
			mockSetup: func(m *MockRepo) {
				m.On("GetByID", mock.Anything, id).
					Return(&domain.FeatureFlag{
						ID:          id,
						Key:         "enable_dark_mode",
						Name:        "Old Name",
						Type:        "boolean",
						Value:       "true",
						Description: "Old description",
						IsActive:    true,
						CreatedAt:   now,
						UpdatedAt:   now,
					}, nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("*domain.FeatureFlag")).
					Return(nil)
			},
			wantErr: false,
			check: func(t *testing.T, f *domain.FeatureFlag) {
				assert.Equal(t, newName, f.Name)
				assert.Equal(t, newValue, f.Value)
				assert.Equal(t, newDesc, f.Description)
				assert.Equal(t, "boolean", f.Type)       // unchanged
				assert.Equal(t, "enable_dark_mode", f.Key) // unchanged
				assert.True(t, f.IsActive)                  // unchanged
			},
		},
		{
			name: "not found on get",
			id:   id,
			req: domain.UpdateFeatureFlagRequest{
				Name: &newName,
			},
			mockSetup: func(m *MockRepo) {
				m.On("GetByID", mock.Anything, id).
					Return(nil, errors.New("not found"))
			},
			wantErr: true,
		},
		{
			name: "repo error on update",
			id:   id,
			req: domain.UpdateFeatureFlagRequest{
				Name: &newName,
			},
			mockSetup: func(m *MockRepo) {
				m.On("GetByID", mock.Anything, id).
					Return(&domain.FeatureFlag{
						ID:          id,
						Key:         "enable_dark_mode",
						Name:        "Old Name",
						Type:        "boolean",
						Value:       "true",
						Description: "Old description",
						IsActive:    true,
						CreatedAt:   now,
						UpdatedAt:   now,
					}, nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("*domain.FeatureFlag")).
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

			result, err := uc.Update(t.Context(), tt.id, tt.req)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				if tt.check != nil {
					tt.check(t, result)
				}
			}

			repo.AssertExpectations(t)
		})
	}
}
