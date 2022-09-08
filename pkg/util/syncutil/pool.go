package syncutil

import (
	"sync"
)

func NewPool[T any](new func() (T, error)) *Pool[T] {
	p := &Pool[T]{new: new}
	return p
}

type Pool[T any] struct {
	Err   error
	new   func() (T, error)
	value any
	mut   sync.RWMutex
}

func (p *Pool[T]) Get() (ret T, err error) {
	p.mut.Lock()
	defer p.mut.Unlock()

	if p.value == nil {
		r, e := p.new()
		if e != nil {
			err = e
			return
		}
		p.value = r
		ret = r
		return
	}
	ret = p.value.(T)
	return
}

func (p *Pool[T]) Put(v T) {
	p.mut.Lock()
	defer p.mut.Unlock()

	p.value = v
}

func (p *Pool[T]) MustGet() T {
	ret, err := p.Get()
	if err != nil {
		panic(err)
	}
	return ret
}
