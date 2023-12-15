package util

import (
	"sync"

	"github.com/gorilla/websocket"
)

type SafeWsConn struct {
	mu       sync.Mutex
	conn     *websocket.Conn
	hostCall bool
}

func NewSafeWsConn(i *websocket.Conn) SafeWsConn {
	return SafeWsConn{
		conn: i,
	}
}
func (c *SafeWsConn) ReadJSON(v interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn.ReadJSON(v)
}
func (c *SafeWsConn) Lock() {
	c.mu.Lock()
}
func (c *SafeWsConn) UnLock() {
	c.mu.Unlock()
}
func (c *SafeWsConn) ReadMessage() (int, []byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn.ReadMessage()
}
func (c *SafeWsConn) WriteMessage(m int, b []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn.WriteMessage(m, b)
}
func (c *SafeWsConn) WriteJSON(v interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.conn.WriteJSON(v)
}
func (c *SafeWsConn) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.conn.Close()
}
func (c *SafeWsConn) SetHostCall(v bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.hostCall = v
}
func (c *SafeWsConn) HostCall() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.hostCall
}

type SafeWsConnMap struct {
	mu    sync.Mutex
	conns map[string]*SafeWsConn
	len   int
}

func NewSafeWsConnMap() SafeWsConnMap {
	return SafeWsConnMap{
		conns: make(map[string]*SafeWsConn),
		len:   0,
	}
}
func (c *SafeWsConnMap) Get(key string) (*SafeWsConn, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if v, ok := c.conns[key]; ok {
		return v, true
	}
	return &SafeWsConn{}, false
}
func (c *SafeWsConnMap) GetAll() map[string]*SafeWsConn {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conns
}
func (c *SafeWsConnMap) Set(key string, conn *SafeWsConn) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.conns[key] = conn
	c.len += 1
}
func (c *SafeWsConnMap) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.conns, key)
	if c.len > 0 {
		c.len -= 1
	}
}
func (c *SafeWsConnMap) SizeHint() int {
	return c.len
}
