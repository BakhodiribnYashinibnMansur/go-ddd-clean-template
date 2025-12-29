package geospatial

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// Geospatial handles Redis Geospatial operations
type Geospatial struct {
	client *redis.Client
}

// New creates a new Geospatial instance
func New(client *redis.Client) *Geospatial {
	return &Geospatial{
		client: client,
	}
}

// GeoAdd adds one or more geospatial items to the specified key
func (g *Geospatial) GeoAdd(ctx context.Context, key string, locations ...*redis.GeoLocation) (int64, error) {
	return g.client.GeoAdd(ctx, key, locations...).Result()
}

// GeoPos returns the positions (longitude, latitude) of members
func (g *Geospatial) GeoPos(ctx context.Context, key string, members ...string) ([]*redis.GeoPos, error) {
	return g.client.GeoPos(ctx, key, members...).Result()
}

// GeoDist returns the distance between two members
func (g *Geospatial) GeoDist(ctx context.Context, key, member1, member2, unit string) (float64, error) {
	return g.client.GeoDist(ctx, key, member1, member2, unit).Result()
}

// GeoRadius returns members within the radius of a point
func (g *Geospatial) GeoRadius(ctx context.Context, key string, longitude, latitude float64, query *redis.GeoRadiusQuery) ([]redis.GeoLocation, error) {
	return g.client.GeoRadius(ctx, key, longitude, latitude, query).Result()
}

// GeoRadiusByMember returns members within the radius of a member
func (g *Geospatial) GeoRadiusByMember(ctx context.Context, key, member string, query *redis.GeoRadiusQuery) ([]redis.GeoLocation, error) {
	return g.client.GeoRadiusByMember(ctx, key, member, query).Result()
}

// GeoSearch searches for members using modern geo commands
func (g *Geospatial) GeoSearch(ctx context.Context, key string, q *redis.GeoSearchQuery) ([]string, error) {
	return g.client.GeoSearch(ctx, key, q).Result()
}

// GeoSearchLocation searches and returns detailed location info
func (g *Geospatial) GeoSearchLocation(ctx context.Context, key string, q *redis.GeoSearchLocationQuery) ([]redis.GeoLocation, error) {
	return g.client.GeoSearchLocation(ctx, key, q).Result()
}

// GeoSearchStore stores search results in another key
func (g *Geospatial) GeoSearchStore(ctx context.Context, key, store string, q *redis.GeoSearchStoreQuery) (int64, error) {
	return g.client.GeoSearchStore(ctx, key, store, q).Result()
}

// GeoHash returns geohash strings for members
func (g *Geospatial) GeoHash(ctx context.Context, key string, members ...string) ([]string, error) {
	return g.client.GeoHash(ctx, key, members...).Result()
}

// GeoRemove removes members from geospatial index (uses ZREM internally)
func (g *Geospatial) GeoRemove(ctx context.Context, key string, members ...string) (int64, error) {
	return g.client.ZRem(ctx, key, members).Result()
}
