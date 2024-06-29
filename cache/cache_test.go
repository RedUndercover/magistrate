package cache

import (
	"testing"
)

// TestCache tests the Cache functions
func TestCache(t *testing.T) {
	c := &Cache{results: make(map[string]interface{})}
	key := "testKey"
	value := "testValue"

	c.Set(key, value)
	result, found := c.Get(key)

	if !found {
		t.Errorf("Expected key %s to be found in cache", key)
	}

	if result != value {
		t.Errorf("Expected value %v to be %v", result, value)
	}
}
