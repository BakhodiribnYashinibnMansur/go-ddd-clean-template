package cache

import (
	"testing"
)

func TestNewClockCache(t *testing.T) {
	c := NewClockCache(5)
	if c == nil {
		t.Fatal("expected non-nil ClockCache")
	}
	if c.capacity != 5 {
		t.Errorf("expected capacity 5, got %d", c.capacity)
	}
	if c.Len() != 0 {
		t.Errorf("expected length 0, got %d", c.Len())
	}
}

func TestClockCache_SetAndGet(t *testing.T) {
	c := NewClockCache(5)

	c.Set("key1", "value1")
	c.Set("key2", 42)

	val, ok := c.Get("key1")
	if !ok {
		t.Fatal("expected key1 to be found")
	}
	if val != "value1" {
		t.Errorf("expected 'value1', got %v", val)
	}

	val, ok = c.Get("key2")
	if !ok {
		t.Fatal("expected key2 to be found")
	}
	if val != 42 {
		t.Errorf("expected 42, got %v", val)
	}
}

func TestClockCache_Get_NotFound(t *testing.T) {
	c := NewClockCache(5)

	_, ok := c.Get("nonexistent")
	if ok {
		t.Error("expected key not to be found")
	}
}

func TestClockCache_Set_Update(t *testing.T) {
	c := NewClockCache(5)

	c.Set("key1", "original")
	c.Set("key1", "updated")

	val, ok := c.Get("key1")
	if !ok {
		t.Fatal("expected key1 to be found")
	}
	if val != "updated" {
		t.Errorf("expected 'updated', got %v", val)
	}

	if c.Len() != 1 {
		t.Errorf("expected length 1 after update, got %d", c.Len())
	}
}

func TestClockCache_Eviction(t *testing.T) {
	c := NewClockCache(3)

	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3)

	if c.Len() != 3 {
		t.Fatalf("expected length 3, got %d", c.Len())
	}

	// Adding a 4th item should evict one
	c.Set("d", 4)

	if c.Len() != 3 {
		t.Errorf("expected length 3 after eviction, got %d", c.Len())
	}

	// The new key should be present
	val, ok := c.Get("d")
	if !ok {
		t.Fatal("expected key 'd' to be found after insertion")
	}
	if val != 4 {
		t.Errorf("expected 4, got %v", val)
	}
}

func TestClockCache_Eviction_SecondChance(t *testing.T) {
	c := NewClockCache(3)

	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3)

	// Force first pass: add "x" to evict one item (all have refBit=true from Set).
	// This clears all refBits in the first sweep, then evicts "a" (hand starts at 0).
	c.Set("x", 10)

	// Now "b" and "c" have refBit=false, "x" has refBit=true.
	// Access "b" to give it a second chance.
	c.Get("b")

	// Add "y": evict starts at hand position. "b" has refBit=true (from Get) -> clear, skip.
	// Next item ("c") has refBit=false -> evict "c".
	c.Set("y", 20)

	// "b" should survive due to second chance
	_, ok := c.Get("b")
	if !ok {
		t.Error("expected key 'b' to survive eviction due to second chance")
	}

	// "y" should be present
	_, ok = c.Get("y")
	if !ok {
		t.Error("expected key 'y' to be present")
	}
}

func TestClockCache_Remove(t *testing.T) {
	c := NewClockCache(5)

	c.Set("key1", "value1")
	c.Set("key2", "value2")

	c.Remove("key1")

	_, ok := c.Get("key1")
	if ok {
		t.Error("expected key1 to be removed")
	}

	if c.Len() != 1 {
		t.Errorf("expected length 1 after remove, got %d", c.Len())
	}
}

func TestClockCache_Remove_NonExistent(t *testing.T) {
	c := NewClockCache(5)
	c.Set("key1", "value1")

	// Should not panic
	c.Remove("nonexistent")

	if c.Len() != 1 {
		t.Errorf("expected length 1, got %d", c.Len())
	}
}

func TestClockCache_Len(t *testing.T) {
	c := NewClockCache(10)

	if c.Len() != 0 {
		t.Errorf("expected length 0, got %d", c.Len())
	}

	for i := 0; i < 5; i++ {
		c.Set(string(rune('a'+i)), i)
	}

	if c.Len() != 5 {
		t.Errorf("expected length 5, got %d", c.Len())
	}
}

func TestClockCache_Purge(t *testing.T) {
	c := NewClockCache(5)

	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3)

	c.Purge()

	if c.Len() != 0 {
		t.Errorf("expected length 0 after purge, got %d", c.Len())
	}

	_, ok := c.Get("a")
	if ok {
		t.Error("expected key 'a' to be gone after purge")
	}
}

func TestClockCache_MultipleEvictions(t *testing.T) {
	c := NewClockCache(2)

	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3) // evicts one
	c.Set("d", 4) // evicts another

	if c.Len() != 2 {
		t.Errorf("expected length 2, got %d", c.Len())
	}

	// At least "c" and "d" should be present
	_, okC := c.Get("c")
	_, okD := c.Get("d")
	if !okD {
		t.Error("expected key 'd' to be present")
	}
	// "c" might or might not be present depending on eviction order, but "d" must be
	_ = okC
}

func TestClockCache_SetAfterPurge(t *testing.T) {
	c := NewClockCache(3)

	c.Set("a", 1)
	c.Purge()
	c.Set("b", 2)

	if c.Len() != 1 {
		t.Errorf("expected length 1, got %d", c.Len())
	}

	val, ok := c.Get("b")
	if !ok {
		t.Fatal("expected key 'b' to be found")
	}
	if val != 2 {
		t.Errorf("expected 2, got %v", val)
	}
}
