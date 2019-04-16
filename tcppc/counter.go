package tcppc

import (
	"sync"
)

var (
	counter = NewSessionCounter()
)

type SessionCounter struct {
	Count uint
	mutex sync.RWMutex
}

func NewSessionCounter() *SessionCounter {
	return &SessionCounter{}
}

func (c *SessionCounter) inc() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.Count += 1
}

func (c *SessionCounter) dec() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.Count -= 1
}

func (c *SessionCounter) count() uint {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.Count
}
