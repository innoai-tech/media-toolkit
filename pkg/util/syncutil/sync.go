package syncutil

import "sync"

type ValueMutex[T any] struct {
	value T
	rw    sync.Mutex
}

func (c *ValueMutex[T]) Get() T {
	c.rw.Lock()
	defer c.rw.Unlock()
	return c.value
}

func (c *ValueMutex[T]) Set(v T) T {
	c.rw.Lock()
	defer c.rw.Unlock()
	c.value = v
	return v
}
