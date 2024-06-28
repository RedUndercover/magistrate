package cache

import (
	"testing"
)

// TestCache tests the Cache functions
func TestCache(t *testing.T) {
	c := &Cache{results: make(map[string][]interface{})}
	key := "testKey"
	value := []interface{}{"value1", "value2"}

	c.Set(key, value)
	result, found := c.Get(key)

	if !found {
		t.Errorf("Expected key %s to be found in cache", key)
	}

	if len(result) != len(value) {
		t.Errorf("Expected cache value length to be %d, got %d", len(value), len(result))
	}
}
