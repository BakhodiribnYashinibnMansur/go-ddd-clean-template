package sitesetting_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"gct/internal/domain"
	"gct/internal/usecase/sitesetting"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockUseCaseI mocks the sitesetting.UseCaseI interface for controller-level testing.
// Since the UseCase internally depends on *persistent.Repo (concrete), we test
// the UseCaseI interface contract directly via mock.
type MockUseCaseI struct {
	mock.Mock
}

func (m *MockUseCaseI) Get(ctx context.Context, filter *domain.SiteSettingFilter) (*domain.SiteSetting, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SiteSetting), args.Error(1)
}

func (m *MockUseCaseI) Gets(ctx context.Context, filter *domain.SiteSettingsFilter) ([]*domain.SiteSetting, int, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*domain.SiteSetting), args.Get(1).(int), args.Error(2)
}

func (m *MockUseCaseI) Update(ctx context.Context, setting *domain.SiteSetting) error {
	args := m.Called(ctx, setting)
	return args.Error(0)
}

func (m *MockUseCaseI) UpdateByKey(ctx context.Context, key, value string) error {
	args := m.Called(ctx, key, value)
	return args.Error(0)
}

func (m *MockUseCaseI) GetByKey(ctx context.Context, key string) (*domain.SiteSetting, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SiteSetting), args.Error(1)
}

// Verify MockUseCaseI implements sitesetting.UseCaseI at compile time.
var _ sitesetting.UseCaseI = (*MockUseCaseI)(nil)

// --- Get ---

func TestGet_Success(t *testing.T) {
	ctx := t.Context()
	uc := new(MockUseCaseI)

	id := uuid.New()
	now := time.Now()
	filter := &domain.SiteSettingFilter{ID: &id}

	expected := &domain.SiteSetting{
		ID: id, Key: "site_name", Value: "My Site",
		ValueType: "string", Category: "general",
		Description: "Site name", IsPublic: true,
		CreatedAt: now, UpdatedAt: now,
	}

	uc.On("Get", ctx, filter).Return(expected, nil)

	result, err := uc.Get(ctx, filter)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "site_name", result.Key)
	assert.Equal(t, "My Site", result.Value)
	assert.True(t, result.IsPublic)
	uc.AssertExpectations(t)
}

func TestGet_NotFound(t *testing.T) {
	ctx := t.Context()
	uc := new(MockUseCaseI)

	id := uuid.New()
	filter := &domain.SiteSettingFilter{ID: &id}

	uc.On("Get", ctx, filter).Return(nil, errors.New("not found"))

	result, err := uc.Get(ctx, filter)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not found")
	uc.AssertExpectations(t)
}

// --- GetByKey ---

func TestGetByKey_Success(t *testing.T) {
	ctx := t.Context()
	uc := new(MockUseCaseI)

	now := time.Now()
	expected := &domain.SiteSetting{
		ID: uuid.New(), Key: "maintenance_mode", Value: "false",
		ValueType: "boolean", Category: "maintenance",
		Description: "Enable maintenance", IsPublic: false,
		CreatedAt: now, UpdatedAt: now,
	}

	uc.On("GetByKey", ctx, "maintenance_mode").Return(expected, nil)

	result, err := uc.GetByKey(ctx, "maintenance_mode")

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "maintenance_mode", result.Key)
	assert.Equal(t, "false", result.Value)
	uc.AssertExpectations(t)
}

func TestGetByKey_NotFound(t *testing.T) {
	ctx := t.Context()
	uc := new(MockUseCaseI)

	uc.On("GetByKey", ctx, "nonexistent").Return(nil, errors.New("not found"))

	result, err := uc.GetByKey(ctx, "nonexistent")

	require.Error(t, err)
	assert.Nil(t, result)
	uc.AssertExpectations(t)
}

// --- Gets ---

func TestGets_Success(t *testing.T) {
	ctx := t.Context()
	uc := new(MockUseCaseI)

	category := "general"
	filter := &domain.SiteSettingsFilter{
		SiteSettingFilter: domain.SiteSettingFilter{
			Category: &category,
		},
	}

	now := time.Now()
	expected := []*domain.SiteSetting{
		{ID: uuid.New(), Key: "site_name", Value: "My Site", ValueType: "string", Category: "general", IsPublic: true, CreatedAt: now, UpdatedAt: now},
		{ID: uuid.New(), Key: "site_logo", Value: "/logo.png", ValueType: "string", Category: "general", IsPublic: true, CreatedAt: now, UpdatedAt: now},
	}

	uc.On("Gets", ctx, filter).Return(expected, 2, nil)

	results, count, err := uc.Gets(ctx, filter)

	require.NoError(t, err)
	assert.Equal(t, 2, count)
	assert.Len(t, results, 2)
	assert.Equal(t, "site_name", results[0].Key)
	uc.AssertExpectations(t)
}

func TestGets_Empty(t *testing.T) {
	ctx := t.Context()
	uc := new(MockUseCaseI)

	filter := &domain.SiteSettingsFilter{}

	uc.On("Gets", ctx, filter).Return([]*domain.SiteSetting{}, 0, nil)

	results, count, err := uc.Gets(ctx, filter)

	require.NoError(t, err)
	assert.Equal(t, 0, count)
	assert.Empty(t, results)
	uc.AssertExpectations(t)
}

func TestGets_Error(t *testing.T) {
	ctx := t.Context()
	uc := new(MockUseCaseI)

	filter := &domain.SiteSettingsFilter{}

	uc.On("Gets", ctx, filter).Return(nil, 0, errors.New("db error"))

	results, count, err := uc.Gets(ctx, filter)

	require.Error(t, err)
	assert.Equal(t, 0, count)
	assert.Nil(t, results)
	uc.AssertExpectations(t)
}

// --- Update ---

func TestUpdate_Success(t *testing.T) {
	ctx := t.Context()
	uc := new(MockUseCaseI)

	setting := &domain.SiteSetting{
		ID: uuid.New(), Key: "site_name", Value: "New Name",
		ValueType: "string", Category: "general",
		Description: "Site name", IsPublic: true,
	}

	uc.On("Update", ctx, setting).Return(nil)

	err := uc.Update(ctx, setting)

	require.NoError(t, err)
	uc.AssertExpectations(t)
}

func TestUpdate_Error(t *testing.T) {
	ctx := t.Context()
	uc := new(MockUseCaseI)

	setting := &domain.SiteSetting{
		ID: uuid.New(), Key: "site_name", Value: "Fail",
	}

	uc.On("Update", ctx, setting).Return(errors.New("update failed"))

	err := uc.Update(ctx, setting)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "update failed")
	uc.AssertExpectations(t)
}

// --- UpdateByKey ---

func TestUpdateByKey_Success(t *testing.T) {
	ctx := t.Context()
	uc := new(MockUseCaseI)

	uc.On("UpdateByKey", ctx, "site_name", "Updated Name").Return(nil)

	err := uc.UpdateByKey(ctx, "site_name", "Updated Name")

	require.NoError(t, err)
	uc.AssertExpectations(t)
}

func TestUpdateByKey_Error(t *testing.T) {
	ctx := t.Context()
	uc := new(MockUseCaseI)

	uc.On("UpdateByKey", ctx, "missing_key", "value").Return(errors.New("key not found"))

	err := uc.UpdateByKey(ctx, "missing_key", "value")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "key not found")
	uc.AssertExpectations(t)
}
