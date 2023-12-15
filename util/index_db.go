package util

import (
	"fmt"
	"sync"
)

type SafeIndexDb struct {
	mu sync.Mutex
	db map[string][]byte //command, key
}

func NewSafeIndexDb() SafeIndexDb {
	return SafeIndexDb{
		db: make(map[string][]byte),
	}
}
func (c *SafeIndexDb) Store(key string, b []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.db[key] = b
}
func (c *SafeIndexDb) Get(key string) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if val, ok := c.db[key]; ok {
		return val, nil
	}
	return nil, fmt.Errorf("IndexDb cannot get for key %v", key)
}
func (c *SafeIndexDb) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.db, key)
}
