package syncutil

import (
	"sync/atomic"
)

func NewBufferChan[T any](size int) *Chan[T] {
	return &Chan[T]{
		ch: make(chan T, size),
	}
}

func NewChan[T any]() *Chan[T] {
	return &Chan[T]{
		ch: make(chan T),
	}
}

type Chan[T any] struct {
	done uint32
	ch   chan T
}

func (o *Chan[T]) Close() {
	if atomic.LoadUint32(&o.done) == 0 {
		close(o.ch)
	}
	atomic.StoreUint32(&o.done, 1)
}

func (o *Chan[T]) Send(t T) {
	if atomic.LoadUint32(&o.done) == 0 {
		o.ch <- t
	}
}

func (o *Chan[T]) Recv() <-chan T {
	return o.ch
}
