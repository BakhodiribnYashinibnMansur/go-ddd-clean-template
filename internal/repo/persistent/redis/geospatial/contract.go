package geospatial

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// GeospatialI defines Redis Geospatial operations interface
type GeospatialI interface {
	GeoAdd(ctx context.Context, key string, locations ...*redis.GeoLocation) (int64, error)
	GeoPos(ctx context.Context, key string, members ...string) ([]*redis.GeoPos, error)
	GeoDist(ctx context.Context, key, member1, member2, unit string) (float64, error)
	GeoRadius(ctx context.Context, key string, longitude, latitude float64, query *redis.GeoRadiusQuery) ([]redis.GeoLocation, error)
	GeoRadiusByMember(ctx context.Context, key, member string, query *redis.GeoRadiusQuery) ([]redis.GeoLocation, error)
	GeoSearch(ctx context.Context, key string, q *redis.GeoSearchQuery) ([]string, error)
	GeoSearchLocation(ctx context.Context, key string, q *redis.GeoSearchLocationQuery) ([]redis.GeoLocation, error)
	GeoSearchStore(ctx context.Context, key, store string, q *redis.GeoSearchStoreQuery) (int64, error)
	GeoHash(ctx context.Context, key string, members ...string) ([]string, error)
	GeoRemove(ctx context.Context, key string, members ...string) (int64, error)
}
