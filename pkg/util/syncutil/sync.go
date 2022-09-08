package syncutil

import (
	"sync/atomic"
)

type Value[T any] struct {
	v atomic.Value
}

func (c *Value[T]) Load() (found T) {
	if v, ok := c.v.Load().(T); ok {
		return v
	}
	return
}

func (c *Value[T]) Store(v T) T {
	c.v.Store(v)
	return v
}

func (c *Value[T]) LoadOrStore(v T) T {
	found := c.v.Load()
	if found == nil {
		return c.Store(v)
	}
	return found.(T)
}
