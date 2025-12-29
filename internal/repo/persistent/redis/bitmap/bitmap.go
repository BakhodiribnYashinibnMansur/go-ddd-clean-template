package bitmap

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// Bitmap handles Redis Bitmap operations
type Bitmap struct {
	client *redis.Client
}

// New creates a new Bitmap instance
func New(client *redis.Client) *Bitmap {
	return &Bitmap{
		client: client,
	}
}

// SetBit sets or clears the bit at offset in the string value stored at key
func (b *Bitmap) SetBit(ctx context.Context, key string, offset int64, value int) (int64, error) {
	return b.client.SetBit(ctx, key, offset, value).Result()
}

// GetBit returns the bit value at offset in the string value stored at key
func (b *Bitmap) GetBit(ctx context.Context, key string, offset int64) (int64, error) {
	return b.client.GetBit(ctx, key, offset).Result()
}

// BitCount counts the number of set bits (population counting) in a string
func (b *Bitmap) BitCount(ctx context.Context, key string, start, end int64) (int64, error) {
	bitCount := &redis.BitCount{
		Start: start,
		End:   end,
	}
	return b.client.BitCount(ctx, key, bitCount).Result()
}

// BitCountAll counts all set bits in the bitmap
func (b *Bitmap) BitCountAll(ctx context.Context, key string) (int64, error) {
	return b.client.BitCount(ctx, key, nil).Result()
}

// BitPos finds the position of the first bit set to 1 or 0
func (b *Bitmap) BitPos(ctx context.Context, key string, bit, start, end int64) (int64, error) {
	return b.client.BitPos(ctx, key, bit, start, end).Result()
}

// BitPosAll finds the position of the first bit in entire bitmap
func (b *Bitmap) BitPosAll(ctx context.Context, key string, bit int64) (int64, error) {
	return b.client.BitPos(ctx, key, bit).Result()
}

// BitOpAnd performs bitwise AND operation between multiple keys
func (b *Bitmap) BitOpAnd(ctx context.Context, destKey string, keys ...string) (int64, error) {
	return b.client.BitOpAnd(ctx, destKey, keys...).Result()
}

// BitOpOr performs bitwise OR operation between multiple keys
func (b *Bitmap) BitOpOr(ctx context.Context, destKey string, keys ...string) (int64, error) {
	return b.client.BitOpOr(ctx, destKey, keys...).Result()
}

// BitOpXor performs bitwise XOR operation between multiple keys
func (b *Bitmap) BitOpXor(ctx context.Context, destKey string, keys ...string) (int64, error) {
	return b.client.BitOpXor(ctx, destKey, keys...).Result()
}

// BitOpNot performs bitwise NOT operation on a key
func (b *Bitmap) BitOpNot(ctx context.Context, destKey, key string) (int64, error) {
	return b.client.BitOpNot(ctx, destKey, key).Result()
}

// BitField performs arbitrary bit field integer operations
func (b *Bitmap) BitField(ctx context.Context, key string, args ...any) ([]int64, error) {
	return b.client.BitField(ctx, key, args...).Result()
}

// Delete removes a bitmap key
func (b *Bitmap) Delete(ctx context.Context, key string) error {
	return b.client.Del(ctx, key).Err()
}

// Exists checks if a bitmap key exists
func (b *Bitmap) Exists(ctx context.Context, key string) (bool, error) {
	count, err := b.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
