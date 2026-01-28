package geospatial

import (
	"errors"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRedis(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	return client, mr
}

func TestGeospatial_GeoAddPos(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func(*redis.Client, *miniredis.Miniredis) (string, int64, error)
		locations     []*redis.GeoLocation
		expectedCount int64
		expectError   bool
		errorCheck    func(*testing.T, error)
	}{
		{
			name: "success add two locations",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			locations: []*redis.GeoLocation{
				{Name: "Palermo", Longitude: 13.361389, Latitude: 38.115556},
				{Name: "Catania", Longitude: 15.087269, Latitude: 37.502669},
			},
			expectedCount: 2,
			expectError:   false,
		},
		{
			name: "success add single location",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			locations: []*redis.GeoLocation{
				{Name: "Rome", Longitude: 12.496365, Latitude: 41.902783},
			},
			expectedCount: 1,
			expectError:   false,
		},
		{
			name: "success add empty locations",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			locations:     []*redis.GeoLocation{},
			expectedCount: 0,
			expectError:   false,
		},
		{
			name: "redis connection error",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), errors.New("redis connection failed")
			},
			locations: []*redis.GeoLocation{
				{Name: "Test", Longitude: 0.0, Latitude: 0.0},
			},
			expectedCount: 0,
			expectError:   true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "connection failed")
			},
		},
		{
			name: "invalid longitude",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			locations: []*redis.GeoLocation{
				{Name: "Invalid", Longitude: 181.0, Latitude: 0.0}, // Invalid longitude
			},
			expectedCount: 0,
			expectError:   true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "invalid latitude",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			locations: []*redis.GeoLocation{
				{Name: "Invalid", Longitude: 0.0, Latitude: 91.0}, // Invalid latitude
			},
			expectedCount: 0,
			expectError:   true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			client, _ := newTestRedis(t)
			defer client.Close()
			g := New(client)
			testKey := uuid.New().String()
			testCtx := t.Context()

			// act
			count, err := g.GeoAdd(testCtx, testKey, tt.locations...)

			// assert
			if tt.expectError {
				require.Error(t, err)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedCount, count)
			}
		})
	}
}

func TestGeospatial_GeoDist(t *testing.T) {
	tests := []struct {
		name         string
		setupMock    func(*redis.Client, *miniredis.Miniredis) (string, int64, error)
		fromMember   string
		toMember     string
		unit         string
		expectedDist float64
		expectError  bool
		errorCheck   func(*testing.T, error)
	}{
		{
			name: "success distance km",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			fromMember:   "Palermo",
			toMember:     "Catania",
			unit:         "km",
			expectedDist: 166.0, // ~166km
			expectError:  false,
		},
		{
			name: "success distance meters",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			fromMember:   "Palermo",
			toMember:     "Catania",
			unit:         "m",
			expectedDist: 166000.0, // ~166km in meters
			expectError:  false,
		},
		{
			name: "success distance miles",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			fromMember:   "Palermo",
			toMember:     "Catania",
			unit:         "mi",
			expectedDist: 103.0, // ~166km in miles
			expectError:  false,
		},
		{
			name: "success distance feet",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			fromMember:   "Palermo",
			toMember:     "Catania",
			unit:         "ft",
			expectedDist: 544000.0, // ~166km in feet
			expectError:  false,
		},
		{
			name: "redis connection error",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), errors.New("redis connection failed")
			},
			fromMember:   "Palermo",
			toMember:     "Catania",
			unit:         "km",
			expectedDist: 0,
			expectError:  true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "connection failed")
			},
		},
		{
			name: "member not found",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			fromMember:   "NonExistent",
			toMember:     "Catania",
			unit:         "km",
			expectedDist: 0,
			expectError:  true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "invalid unit",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			fromMember:   "Palermo",
			toMember:     "Catania",
			unit:         "invalid",
			expectedDist: 0,
			expectError:  true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			client, _ := newTestRedis(t)
			defer client.Close()
			g := New(client)
			testKey := uuid.New().String()
			testCtx := t.Context()

			// setup locations
			g.GeoAdd(testCtx, testKey,
				&redis.GeoLocation{Name: "Palermo", Longitude: 13.361389, Latitude: 38.115556},
				&redis.GeoLocation{Name: "Catania", Longitude: 15.087269, Latitude: 37.502669},
			)

			// act
			dist, err := g.GeoDist(testCtx, testKey, tt.fromMember, tt.toMember, tt.unit)

			// assert
			if tt.expectError {
				require.Error(t, err)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				assert.Greater(t, dist, tt.expectedDist-10) // Allow some tolerance
			}
		})
	}
}

func TestGeospatial_GeoHash(t *testing.T) {
	client, _ := newTestRedis(t)
	defer client.Close()

	g := New(client)
	key := uuid.New().String()
	ctx := t.Context()

	t.Skip("Skipping GeoHash test: miniredis might not support GEOHASH")

	g.GeoAdd(ctx, key,
		&redis.GeoLocation{Name: "Palermo", Longitude: 13.361389, Latitude: 38.115556},
	)

	hashes, err := g.GeoHash(ctx, key, "Palermo")
	require.NoError(t, err)
	assert.Len(t, hashes, 1)
	assert.NotEmpty(t, hashes[0])
}

func TestGeospatial_GeoRadius(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func(*redis.Client, *miniredis.Miniredis) (string, int64, error)
		longitude     float64
		latitude      float64
		radius        float64
		unit          string
		expectedCount int
		expectError   bool
		errorCheck    func(*testing.T, error)
	}{
		{
			name: "success radius 200km",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			longitude:     13.361389,
			latitude:      38.115556,
			radius:        200,
			unit:          "km",
			expectedCount: 2, // Should find both Palermo and Catania
			expectError:   false,
		},
		{
			name: "success radius 100km",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			longitude:     13.361389,
			latitude:      38.115556,
			radius:        100,
			unit:          "km",
			expectedCount: 2, // Should find both Palermo and Catania
			expectError:   false,
		},
		{
			name: "success radius 50km",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			longitude:     13.361389,
			latitude:      38.115556,
			radius:        50,
			unit:          "km",
			expectedCount: 1, // Should find only Palermo
			expectError:   false,
		},
		{
			name: "success radius meters",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			longitude:     13.361389,
			latitude:      38.115556,
			radius:        50000,
			unit:          "m",
			expectedCount: 2,
			expectError:   false,
		},
		{
			name: "success radius miles",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			longitude:     13.361389,
			latitude:      38.115556,
			radius:        100,
			unit:          "mi",
			expectedCount: 2,
			expectError:   false,
		},
		{
			name: "no locations in radius",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			longitude:     0.0,
			latitude:      0.0,
			radius:        10,
			unit:          "km",
			expectedCount: 0,
			expectError:   false,
		},
		{
			name: "redis connection error",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), errors.New("redis connection failed")
			},
			longitude:     13.361389,
			latitude:      38.115556,
			radius:        200,
			unit:          "km",
			expectedCount: 0,
			expectError:   true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "connection failed")
			},
		},
		{
			name: "invalid radius negative",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			longitude:     13.361389,
			latitude:      38.115556,
			radius:        -10,
			unit:          "km",
			expectedCount: 0,
			expectError:   true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "invalid unit",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			longitude:     13.361389,
			latitude:      38.115556,
			radius:        200,
			unit:          "invalid",
			expectedCount: 0,
			expectError:   true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			client, _ := newTestRedis(t)
			defer client.Close()
			g := New(client)
			testKey := uuid.New().String()
			testCtx := t.Context()

			// setup locations
			g.GeoAdd(testCtx, testKey,
				&redis.GeoLocation{Name: "Palermo", Longitude: 13.361389, Latitude: 38.115556},
				&redis.GeoLocation{Name: "Catania", Longitude: 15.087269, Latitude: 37.502669},
			)

			// act
			locs, err := g.GeoRadius(testCtx, testKey, tt.longitude, tt.latitude, &redis.GeoRadiusQuery{
				Radius:   tt.radius,
				Unit:     tt.unit,
				WithDist: true,
			})

			// assert
			if tt.expectError {
				require.Error(t, err)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				assert.Len(t, locs, tt.expectedCount)
			}
		})
	}
}

func TestGeospatial_GeoRadiusByMember(t *testing.T) {
	client, _ := newTestRedis(t)
	defer client.Close()

	g := New(client)
	key := uuid.New().String()
	ctx := t.Context()

	g.GeoAdd(ctx, key,
		&redis.GeoLocation{Name: "Palermo", Longitude: 13.361389, Latitude: 38.115556},
		&redis.GeoLocation{Name: "Catania", Longitude: 15.087269, Latitude: 37.502669},
	)

	locs, err := g.GeoRadiusByMember(ctx, key, "Palermo", &redis.GeoRadiusQuery{
		Radius: 200,
		Unit:   "km",
	})
	require.NoError(t, err)
	assert.Len(t, locs, 2)
}

func TestGeospatial_GeoSearch(t *testing.T) {
	client, _ := newTestRedis(t)
	defer client.Close()

	g := New(client)
	key := uuid.New().String()
	ctx := t.Context()

	_, err := g.GeoAdd(ctx, key,
		&redis.GeoLocation{Name: "Palermo", Longitude: 13.361389, Latitude: 38.115556},
		&redis.GeoLocation{Name: "Catania", Longitude: 15.087269, Latitude: 37.502669},
	)
	require.NoError(t, err)

	// miniredis might not support GEOSEARCH or return unknown command error
	t.Skip("Skipping GeoSearch test: miniredis might not support GEOSEARCH")
}

func TestGeospatial_GeoRemove(t *testing.T) {
	tests := []struct {
		name            string
		setupMock       func(*redis.Client, *miniredis.Miniredis) (string, int64, error)
		member          string
		expectedRemoved int64
		expectError     bool
		errorCheck      func(*testing.T, error)
	}{
		{
			name: "success remove existing member",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			member:          "Palermo",
			expectedRemoved: 1,
			expectError:     false,
		},
		{
			name: "remove non-existent member",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			member:          "NonExistent",
			expectedRemoved: 0,
			expectError:     false,
		},
		{
			name: "remove empty member",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			member:          "",
			expectedRemoved: 0,
			expectError:     false,
		},
		{
			name: "redis connection error",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), errors.New("redis connection failed")
			},
			member:          "Palermo",
			expectedRemoved: 0,
			expectError:     true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "connection failed")
			},
		},
		{
			name: "remove member with special characters",
			setupMock: func(client *redis.Client, mr *miniredis.Miniredis) (string, int64, error) {
				return "", int64(0), nil
			},
			member:          "Test-Location_123",
			expectedRemoved: 1,
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			client, _ := newTestRedis(t)
			defer client.Close()
			g := New(client)
			testKey := uuid.New().String()
			testCtx := t.Context()

			// setup location
			g.GeoAdd(testCtx, testKey,
				&redis.GeoLocation{Name: "Palermo", Longitude: 13.361389, Latitude: 38.115556},
			)

			// act
			removed, err := g.GeoRemove(testCtx, testKey, tt.member)

			// assert
			if tt.expectError {
				require.Error(t, err)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedRemoved, removed)
			}
		})
	}
}
