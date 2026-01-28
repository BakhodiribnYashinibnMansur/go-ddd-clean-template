package store

import (
	"github.com/redis/go-redis/v9"
)

func New(redisConn *redis.Client) *Repo {
	return &Repo{
		Primitive: &Primitives{
			String: NewPrimitive[string](redisConn),
			Int:    NewPrimitive[int64](redisConn),
			Byte:   NewPrimitive[[]byte](redisConn),
			Bool:   NewPrimitive[bool](redisConn),
			Float:  NewPrimitive[float64](redisConn),
		},
		Array: &Arrays{
			String: NewArray[string](redisConn),
			Int:    NewArray[int64](redisConn),
			Byte:   NewArray[[]byte](redisConn),
		},
		HashTable: &HashTables{
			String: NewHashTable[string](redisConn),
			Int:    NewHashTable[int64](redisConn),
			Byte:   NewHashTable[[]byte](redisConn),
		},
		Set: &Sets{
			String: NewSet[string](redisConn),
			Int:    NewSet[int64](redisConn),
			Byte:   NewSet[[]byte](redisConn),
		},
		Queue: &Queues{
			String: NewQueue[string](redisConn),
			Int:    NewQueue[int64](redisConn),
			Byte:   NewQueue[[]byte](redisConn),
		},
		List: &Lists{
			String: NewList[string](redisConn),
			Int:    NewList[int64](redisConn),
			Byte:   NewList[[]byte](redisConn),
		},
		PriorityQueue: &PriorityQueues{
			String: NewPriorityQueue[string](redisConn),
			Int:    NewPriorityQueue[int64](redisConn),
			Byte:   NewPriorityQueue[[]byte](redisConn),
		},
	}
}

type Arrays struct {
	String ArrayI[string]
	Int    ArrayI[int64]
	Byte   ArrayI[[]byte]
}

type HashTables struct {
	String HashTableI[string]
	Int    HashTableI[int64]
	Byte   HashTableI[[]byte]
}

type Sets struct {
	String SetI[string]
	Int    SetI[int64]
	Byte   SetI[[]byte]
}

type Queues struct {
	String QueueI[string]
	Int    QueueI[int64]
	Byte   QueueI[[]byte]
}

type Lists struct {
	String ListI[string]
	Int    ListI[int64]
	Byte   ListI[[]byte]
}

type PriorityQueues struct {
	String PriorityQueueI[string]
	Int    PriorityQueueI[int64]
	Byte   PriorityQueueI[[]byte]
}

type Repo struct {
	Primitive     *Primitives
	Array         *Arrays
	HashTable     *HashTables
	Set           *Sets
	Queue         *Queues
	List          *Lists
	PriorityQueue *PriorityQueues
}
