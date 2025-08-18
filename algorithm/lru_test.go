package algorithm

import "testing"

func TestLRUCache(t *testing.T) {
	cache := NewLRUCache[int, string](2)

	cache.Put(1, "one")
	cache.Put(2, "two")

	if val, ok := cache.Get(1); !ok || val != "one" {
		t.Errorf("Expected to get 'one' for key 1, got '%s'", val)
	}

	cache.Put(3, "three") // This should evict key 2

	if _, ok := cache.Get(2); ok {
		t.Errorf("Expected key 2 to be evicted")
	}

	cache.Put(4, "four") // This should evict key 1

	if _, ok := cache.Get(1); ok {
		t.Errorf("Expected key 1 to be evicted")
	}

	if val, ok := cache.Get(3); !ok || val != "three" {
		t.Errorf("Expected to get 'three' for key 3, got '%s'", val)
	}

	if val, ok := cache.Get(4); !ok || val != "four" {
		t.Errorf("Expected to get 'four' for key 4, got '%s'", val)
	}
}
