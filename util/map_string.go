package util

import "sync"

type SafeStringMap struct {
	mu sync.Mutex
	db map[string]string
}

func NewSafeStringMap() SafeStringMap {
	return SafeStringMap{db: make(map[string]string)}
}
func (c *SafeStringMap) Get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if v, ok := c.db[key]; ok {
		return v, true
	}
	return "", false
}
func (c *SafeStringMap) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.db, key)
}
func (c *SafeStringMap) Range(fn func(key string, value string) bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, v := range c.db {
		if !fn(k, v) {
			break
		}
	}
}
func (c *SafeStringMap) Store(key string, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.db[key] = value
}
