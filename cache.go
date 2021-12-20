package vey

import (
	"sync"
)

// MemCache implements Cache interface
type MemCache struct {
	m      sync.Mutex
	values map[string]Cached
}

func NewMemCache() Cache {
	return &MemCache{
		values: make(map[string]Cached),
	}
}

func (c *MemCache) Set(key string, val Cached) error {
	c.m.Lock()
	defer c.m.Unlock()
	c.values[key] = val
	return nil
}

func (c *MemCache) Get(key string) (Cached, error) {
	c.m.Lock()
	defer c.m.Unlock()
	return c.values[key], nil
}
