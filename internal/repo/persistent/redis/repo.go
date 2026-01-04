package redis

import (
	"gct/internal/repo/persistent/redis/bitmap"
	"gct/internal/repo/persistent/redis/geospatial"
	"gct/internal/repo/persistent/redis/hyperloglog"
	"gct/internal/repo/persistent/redis/pubsub"
	"gct/internal/repo/persistent/redis/store"
	"gct/internal/repo/persistent/redis/stream"
	"gct/pkg/logger"
	"github.com/redis/go-redis/v9"
)

type Repo struct {
	Primitive     *store.Primitives
	Array         *store.Arrays
	HashTable     *store.HashTables
	Set           *store.Sets
	PriorityQueue *store.PriorityQueues
	Queue         *store.Queues
	List          *store.Lists
	PubSub        pubsub.PubSubI
	Stream        stream.StreamI
	Geospatial    geospatial.GeospatialI
	HyperLogLog   hyperloglog.HyperLogLogI
	Bitmap        bitmap.BitmapI
}

func New(redisConn *redis.Client, logger logger.Log) *Repo {
	s := store.New(redisConn)
	return &Repo{
		Primitive:     s.Primitive,
		Array:         s.Array,
		HashTable:     s.HashTable,
		Set:           s.Set,
		PriorityQueue: s.PriorityQueue,
		Queue:         s.Queue,
		List:          s.List,
		PubSub:        pubsub.New(redisConn),
		Stream:        stream.New(redisConn),
		Geospatial:    geospatial.New(redisConn),
		HyperLogLog:   hyperloglog.New(redisConn),
		Bitmap:        bitmap.New(redisConn),
	}
}
