package bitmap

import (
	"context"
	"fmt"

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
	result, err := b.client.SetBit(ctx, key, offset, value).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to set bit at offset %d in bitmap %s: %w", offset, key, err)
	}
	return result, nil
}

// GetBit returns the bit value at offset in the string value stored at key
func (b *Bitmap) GetBit(ctx context.Context, key string, offset int64) (int64, error) {
	result, err := b.client.GetBit(ctx, key, offset).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get bit at offset %d from bitmap %s: %w", offset, key, err)
	}
	return result, nil
}

// BitCount counts the number of set bits (population counting) in a string
func (b *Bitmap) BitCount(ctx context.Context, key string, start, end int64) (int64, error) {
	bitCount := &redis.BitCount{
		Start: start,
		End:   end,
	}
	result, err := b.client.BitCount(ctx, key, bitCount).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to count bits in range [%d, %d] in bitmap %s: %w", start, end, key, err)
	}
	return result, nil
}

// BitCountAll counts all set bits in the bitmap
func (b *Bitmap) BitCountAll(ctx context.Context, key string) (int64, error) {
	result, err := b.client.BitCount(ctx, key, nil).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to count bits in bitmap %s: %w", key, err)
	}
	return result, nil
}

// BitPos finds the position of the first bit set to 1 or 0
func (b *Bitmap) BitPos(ctx context.Context, key string, bit, start, end int64) (int64, error) {
	result, err := b.client.BitPos(ctx, key, bit, start, end).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to find bit position in bitmap %s: %w", key, err)
	}
	return result, nil
}

// BitPosAll finds the position of the first bit in entire bitmap
func (b *Bitmap) BitPosAll(ctx context.Context, key string, bit int64) (int64, error) {
	result, err := b.client.BitPos(ctx, key, bit).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to find bit position in bitmap %s: %w", key, err)
	}
	return result, nil
}

// BitOpAnd performs bitwise AND operation between multiple keys
func (b *Bitmap) BitOpAnd(ctx context.Context, destKey string, keys ...string) (int64, error) {
	result, err := b.client.BitOpAnd(ctx, destKey, keys...).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to perform bitwise AND operation on keys %v: %w", keys, err)
	}
	return result, nil
}

// BitOpOr performs bitwise OR operation between multiple keys
func (b *Bitmap) BitOpOr(ctx context.Context, destKey string, keys ...string) (int64, error) {
	result, err := b.client.BitOpOr(ctx, destKey, keys...).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to perform bitwise OR operation on keys %v: %w", keys, err)
	}
	return result, nil
}

// BitOpXor performs bitwise XOR operation between multiple keys
func (b *Bitmap) BitOpXor(ctx context.Context, destKey string, keys ...string) (int64, error) {
	result, err := b.client.BitOpXor(ctx, destKey, keys...).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to perform bitwise XOR operation on keys %v: %w", keys, err)
	}
	return result, nil
}

// BitOpNot performs bitwise NOT operation on a key
func (b *Bitmap) BitOpNot(ctx context.Context, destKey, key string) (int64, error) {
	result, err := b.client.BitOpNot(ctx, destKey, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to perform bitwise NOT operation on key %s: %w", key, err)
	}
	return result, nil
}

// BitField performs bitwise operations on specific fields
func (b *Bitmap) BitField(ctx context.Context, key string, args ...any) ([]int64, error) {
	result, err := b.client.BitField(ctx, key, args...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to perform bit field operations on key %s: %w", key, err)
	}
	return result, nil
}

// Delete removes the bitmap key
func (b *Bitmap) Delete(ctx context.Context, key string) error {
	if err := b.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete bitmap key %s: %w", key, err)
	}
	return nil
}

// Exists checks if a bitmap key exists
func (b *Bitmap) Exists(ctx context.Context, key string) (bool, error) {
	count, err := b.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check existence of bitmap key %s: %w", key, err)
	}
	return count > 0, nil
}
