package geospatial

import (
	"context"
	"fmt"

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
	result, err := g.client.GeoAdd(ctx, key, locations...).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to add geospatial items to key %s: %w", key, err)
	}
	return result, nil
}

// GeoPos returns the positions (longitude, latitude) of members
func (g *Geospatial) GeoPos(ctx context.Context, key string, members ...string) ([]*redis.GeoPos, error) {
	result, err := g.client.GeoPos(ctx, key, members...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get positions for members %v in key %s: %w", members, key, err)
	}
	return result, nil
}

// GeoDist returns the distance between two members
func (g *Geospatial) GeoDist(ctx context.Context, key, member1, member2, unit string) (float64, error) {
	result, err := g.client.GeoDist(ctx, key, member1, member2, unit).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to calculate distance between %s and %s in key %s: %w", member1, member2, key, err)
	}
	return result, nil
}

// GeoRadius returns members within the radius of a point
func (g *Geospatial) GeoRadius(ctx context.Context, key string, longitude, latitude float64, query *redis.GeoRadiusQuery) ([]redis.GeoLocation, error) {
	result, err := g.client.GeoRadius(ctx, key, longitude, latitude, query).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get members within radius of point (%f, %f) in key %s: %w", longitude, latitude, key, err)
	}
	return result, nil
}

// GeoRadiusByMember returns members within the radius of a member
func (g *Geospatial) GeoRadiusByMember(ctx context.Context, key, member string, query *redis.GeoRadiusQuery) ([]redis.GeoLocation, error) {
	result, err := g.client.GeoRadiusByMember(ctx, key, member, query).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get members within radius of member %s in key %s: %w", member, key, err)
	}
	return result, nil
}

// GeoSearch searches for members using modern geo commands
func (g *Geospatial) GeoSearch(ctx context.Context, key string, q *redis.GeoSearchQuery) ([]string, error) {
	result, err := g.client.GeoSearch(ctx, key, q).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to search geospatial data in key %s: %w", key, err)
	}
	return result, nil
}

// GeoSearchLocation searches and returns detailed location info
func (g *Geospatial) GeoSearchLocation(ctx context.Context, key string, q *redis.GeoSearchLocationQuery) ([]redis.GeoLocation, error) {
	result, err := g.client.GeoSearchLocation(ctx, key, q).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to search geospatial location data in key %s: %w", key, err)
	}
	return result, nil
}

// GeoSearchStore stores search results in another key
func (g *Geospatial) GeoSearchStore(ctx context.Context, key, store string, q *redis.GeoSearchStoreQuery) (int64, error) {
	result, err := g.client.GeoSearchStore(ctx, key, store, q).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to store geospatial search results from key %s to %s: %w", key, store, err)
	}
	return result, nil
}

// GeoHash returns the geohash string of members
func (g *Geospatial) GeoHash(ctx context.Context, key string, members ...string) ([]string, error) {
	result, err := g.client.GeoHash(ctx, key, members...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get geohash for members %v in key %s: %w", members, key, err)
	}
	return result, nil
}

// GeoRemove removes members from the geospatial key
func (g *Geospatial) GeoRemove(ctx context.Context, key string, members ...string) (int64, error) {
	result, err := g.client.ZRem(ctx, key, members).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to remove members %v from geospatial key %s: %w", members, key, err)
	}
	return result, nil
}
