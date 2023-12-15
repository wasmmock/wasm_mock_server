package util

import "sync"

type SafeHitMap struct {
	mu sync.Mutex
	db []Hit
}

func NewSafeHitMap() SafeHitMap {
	return SafeHitMap{db: make([]Hit, 0)}
}

func (c *SafeHitMap) Range(fn func(hit Hit) bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, v := range c.db {
		if !fn(v) {
			break
		}
	}
}
func (c *SafeHitMap) Clone() []Hit {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.db
}
func (c *SafeHitMap) Store(hit Hit) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.db = append(c.db, hit)
}
