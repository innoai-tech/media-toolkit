package syncutil

import "sync"

type Map[K any, V any] struct {
	sync.Map
}

func (x *Map[K, V]) Load(k K) (found V, ok bool) {
	got, ok := x.Map.Load(k)
	if ok {
		return got.(V), true
	}
	return
}
func (x *Map[K, V]) Store(k K, v V) {
	x.Map.Store(k, v)
}

func (x *Map[K, V]) LoadOrStore(k K, v V) (found V, ok bool) {
	got, ok := x.Map.LoadOrStore(k, v)
	if ok {
		return got.(V), true
	}
	return
}

func (x *Map[K, V]) LoadAndDelete(k K) (found V, ok bool) {
	got, ok := x.Map.LoadAndDelete(k)
	if ok {
		return got.(V), true
	}
	return
}
