package tcpproxy

import (
	"fmt"
	"net"
	"sync"
)

type SafeTcpConnMap struct {
	mu sync.Mutex
	db map[string]map[string]net.Conn //command, key
}

func NewSafeTcpConnMap() SafeTcpConnMap {
	return SafeTcpConnMap{
		db: make(map[string]map[string]net.Conn),
	}
}
func (c *SafeTcpConnMap) Store(key string, l_remote_add string, b net.Conn) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if v, ok := c.db[key]; ok {
		v[l_remote_add] = b
	} else {
		c.db[key] = make(map[string]net.Conn)
		c.db[key][l_remote_add] = b
	}
}
func (c *SafeTcpConnMap) Get(key string, l_remote_add string) (net.Conn, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if val, ok := c.db[key]; ok {
		if v, ok := val[l_remote_add]; ok {
			return v, nil
		}
	}
	return nil, fmt.Errorf("TcpConnMap cannot get for key %v", key)
}
func (c *SafeTcpConnMap) GetAll(key string) map[string]net.Conn {
	c.mu.Lock()
	defer c.mu.Unlock()
	if val, ok := c.db[key]; ok {
		return val
	}
	return map[string]net.Conn{}
}
func (c *SafeTcpConnMap) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.db, key)
}
