package cache

import (
	"fmt"
	"testing"
)

func BenchmarkLRUCache_Set(b *testing.B) {
	c := NewLRUCache(1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Set(fmt.Sprintf("key-%d", i), i)
	}
}

func BenchmarkLRUCache_Get(b *testing.B) {
	c := NewLRUCache(1024)
	for i := 0; i < 1024; i++ {
		c.Set(fmt.Sprintf("key-%d", i), i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Get(fmt.Sprintf("key-%d", i%1024))
	}
}

func BenchmarkLRUCache_SetParallel(b *testing.B) {
	c := NewLRUCache(1024)

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			c.Set(fmt.Sprintf("key-%d", i), i)
			i++
		}
	})
}
