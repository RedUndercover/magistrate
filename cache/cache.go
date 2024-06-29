package cache

import (
	"sync"
)

// Cache stores the results of plugin scans.
type Cache struct {
	mu      sync.Mutex
	results map[string]interface{}
}

var PluginRegistry = Cache{
	results: make(map[string]interface{}),
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	result, found := c.results[key]
	return result, found
}

func (c *Cache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.results[key] = value
}
