package util

import (
	"context"
	"sync"

	"github.com/wapc/wapc-go"
)

type SafeInstance struct {
	mu       sync.Mutex
	instance wapc.Instance
}

func NewSafeInstance(i wapc.Instance) SafeInstance {
	return SafeInstance{
		instance: i,
	}
}
func (c *SafeInstance) Invoke(ctx context.Context, g string, b []byte) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.instance.Invoke(ctx, g, b)
}
func (c *SafeInstance) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	ctx := context.Background()
	c.instance.Close(ctx)
}

type SafeInstanceMap struct {
	mu        sync.Mutex
	instances map[string]*SafeInstance
}

func (c *SafeInstanceMap) Get(key string) (*SafeInstance, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if v, ok := c.instances[key]; ok {
		return v, true
	}
	return &SafeInstance{}, false
}
func (c *SafeInstanceMap) GetAll() map[string]*SafeInstance {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.instances
}
func (c *SafeInstanceMap) Set(key string, instance *SafeInstance) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.instances[key] = instance
}
func (c *SafeInstanceMap) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.instances, key)
}
func NewSafeInstanceMap() SafeInstanceMap {
	return SafeInstanceMap{
		instances: make(map[string]*SafeInstance),
	}
}
