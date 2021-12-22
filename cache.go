package vey

import (
	"encoding/base64"
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

func (c *MemCache) Set(key []byte, val Cached) error {
	c.m.Lock()
	defer c.m.Unlock()
	str := base64.StdEncoding.EncodeToString(key)
	c.values[str] = val
	return nil
}

func (c *MemCache) Get(key []byte) (Cached, error) {
	c.m.Lock()
	defer c.m.Unlock()
	str := base64.StdEncoding.EncodeToString(key)
	return c.values[str], nil
}
